package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/vorpalengineering/gundler/internal/mempool"
	"github.com/vorpalengineering/gundler/internal/processor"
	"github.com/vorpalengineering/gundler/internal/types"
)

type RPCServer struct {
	server               *http.Server
	ethClient            *ethclient.Client
	mempools             map[string]*mempool.Mempool // entryPointAddress => Mempool
	processors           map[string]processor.Processor
	chainID              *big.Int
	supportedEntryPoints []string
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

func NewRPCServer(port uint, ethRPC string, supportedEntryPoints []string) (*RPCServer, error) {
	// Dial ethereum client
	ethClient, err := ethclient.Dial(ethRPC)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ethereum client: %w", err)
	}

	// Fetch chain id
	ctx := context.Background()
	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain id from rpc: %v", err)
	}
	log.Printf("Connected to chain ID: %v", chainID)

	// Initialize mux handler
	mux := http.NewServeMux()

	// Register healthcheck route
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Initialize mempool and processor for each supported entrypoint
	mempools := make(map[string]*mempool.Mempool, len(supportedEntryPoints))
	processors := make(map[string]processor.Processor, len(supportedEntryPoints))
	for _, epStr := range supportedEntryPoints {
		// Create mempool
		entryPoint := common.HexToAddress(epStr)
		mempools[epStr] = mempool.NewMempool(entryPoint, chainID)

		// Create processor
		processors[epStr] = processor.NewBasicProcessor(mempools[epStr], ethClient, 1*time.Second)
		if err := processors[epStr].Start(context.Background()); err != nil {
			log.Fatalf("Failed to start processor: %v", err)
		}

		log.Printf("Initialized mempool and processor for entry point: %s", epStr)
	}

	rpc := &RPCServer{
		server: &http.Server{
			Addr:    fmt.Sprintf("localhost:%v", port),
			Handler: mux,
		},
		ethClient:            ethClient,
		mempools:             mempools,
		processors:           processors,
		chainID:              chainID,
		supportedEntryPoints: supportedEntryPoints,
	}

	// Register base route
	mux.HandleFunc("/", rpc.handleRPCRequest)

	// Register debug endpoints
	mux.HandleFunc("/debug_mempools", rpc.handleDebugMempools)
	mux.HandleFunc("/debug_pause", rpc.handleDebugPause)
	mux.HandleFunc("/debug_clear", rpc.handleDebugClear)

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

	// Stop processors
	for _, proc := range rpc.processors {
		if err := proc.Stop(); err != nil {
			log.Printf("Failed to stop processor: %v", err)
		}
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
	return fmt.Sprintf("0x%x", rpc.chainID), nil
}

func (rpc *RPCServer) handleSupportedEntryPoints() ([]string, *RPCError) {
	return rpc.supportedEntryPoints, nil
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

	// Parse EntryPoint address from rawParams[1]
	var entryPointStr string
	if err := json.Unmarshal(rawParams[1], &entryPointStr); err != nil {
		return "", &RPCError{
			Code:    -32602,
			Message: "Error unmarshalling entryPoint",
		}
	}
	entryPoint := common.HexToAddress(entryPointStr)

	// Validate EntryPoint address
	err := types.ValidateEntryPointAddress(entryPoint)
	if err != nil {
		return "", &RPCError{
			Code:    -32602,
			Message: "Invalid EntryPoint Address",
		}
	}

	// TODO: Validate UserOperation

	// Add to mempool
	if err := rpc.mempools[entryPointStr].Add(&userOp); err != nil {
		return "", &RPCError{
			Code:    -32602,
			Message: fmt.Sprintf("Failed adding userOp to mempool: %v", err),
		}
	}

	// Calculate userOp hash
	userOpHash := userOp.Hash(entryPoint, rpc.chainID)

	log.Printf("UserOp %s validated and added to mempool. Mempool size: %v", userOpHash.Hex(), rpc.mempools[entryPointStr].Size())

	return userOpHash.Hex(), nil
}
