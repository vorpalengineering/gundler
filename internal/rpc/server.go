package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/vorpalengineering/gundler/internal/types"
)

type RPCServer struct {
	server    *http.Server
	ethClient *ethclient.Client
}

type RPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      any             `json:"id"`
}

type RPCResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	Result  any       `json:"result,omitempty"`
	Error   *RPCError `json:"error,omitempty"`
	ID      any       `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func New(port uint, ethRPC string) (*RPCServer, error) {
	// Dial ethereum client
	ethClient, err := ethclient.Dial(ethRPC)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ethereum client: %w", err)
	}

	// Initialize mux handler
	mux := http.NewServeMux()

	// Register healthcheck route
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	rpc := &RPCServer{
		server: &http.Server{
			Addr:    fmt.Sprintf("localhost:%v", port),
			Handler: mux,
		},
		ethClient: ethClient,
	}

	// Register base route
	mux.HandleFunc("/", rpc.handleRPCRequest)

	return rpc, nil
}

func (rpc *RPCServer) Start() error {
	fmt.Printf("Starting RPC Server on: %s\n", rpc.server.Addr)

	go func() {
		if err := rpc.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("RPC Server error: %v", err)
		}
	}()

	return nil
}

func (rpc *RPCServer) Shutdown(ctx context.Context) error {
	fmt.Println("Shutting down RPC Server...")

	// Close ethereum client connection
	if rpc.ethClient != nil {
		rpc.ethClient.Close()
	}

	return rpc.server.Shutdown(ctx)
}

func (rpc *RPCServer) handleRPCRequest(w http.ResponseWriter, r *http.Request) {
	// Restrict to POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode request as JSON
	var req RPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rpc.sendError(w, nil, -32700, "Parse error")
		return
	}

	// Route to appropriate handler
	var result any
	var err *RPCError

	switch req.Method {
	case "eth_chainId":
		result, err = rpc.handleChainId()
	case "eth_supportedEntryPoints":
		result, err = rpc.handleSupportedEntryPoints()
	case "eth_sendUserOperation":
		result, err = rpc.handleSendUserOperation(req.Params)
	default:
		err = &RPCError{
			Code:    -32601,
			Message: "Method not found",
		}
	}

	// Send response
	if err != nil {
		rpc.sendError(w, req.ID, err.Code, err.Message)
		return
	}

	rpc.sendResult(w, req.ID, result)
}

func (rpc *RPCServer) sendResult(w http.ResponseWriter, id any, result any) {
	resp := &RPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (rpc *RPCServer) sendError(w http.ResponseWriter, id any, code int, message string) {
	resp := &RPCResponse{
		JSONRPC: "2.0",
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (rpc *RPCServer) handleChainId() (string, *RPCError) {
	ctx := context.Background()
	chainID, err := rpc.ethClient.ChainID(ctx)
	if err != nil {
		return "", &RPCError{
			Code:    -32000,
			Message: fmt.Sprintf("failed to get chain id: %v", err),
		}
	}

	return fmt.Sprintf("0x%x", chainID), nil
}

func (rpc *RPCServer) handleSupportedEntryPoints() ([]string, *RPCError) {
	return []string{
		"0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789", // v0.6
		"0x0000000071727De22E5E9d8BAf0edAc6f37da032", // v0.7
	}, nil
}

func (rpc *RPCServer) handleSendUserOperation(params json.RawMessage) (string, *RPCError) {
	// Parse json params
	var rawParams []json.RawMessage
	if err := json.Unmarshal(params, &rawParams); err != nil {
		return "", &RPCError{
			Code:    -32602,
			Message: "Invalid parameters",
		}
	}

	if len(rawParams) != 2 {
		return "", &RPCError{
			Code:    -32602,
			Message: "Expected 2 parameters: [userOp, entryPoint]",
		}
	}

	// Parse UserOperation from rawParams[0]
	var userOp types.UserOperation
	if err := json.Unmarshal(rawParams[0], &userOp); err != nil {
		return "", &RPCError{
			Code:    -32602,
			Message: "Error unmarshalling userOp",
		}
	}

	// TODO: Parse EntryPoint address from rawParams[1]
	// TODO: Validate UserOperation
	// TODO: Add to mempool
	// TODO: Return userOp hash

	return "0x0000000000000000000000000000000000000000000000000000000000000000", nil
}
