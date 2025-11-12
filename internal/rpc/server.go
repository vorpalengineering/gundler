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
	"github.com/vorpalengineering/gundler/internal/keypool"
	"github.com/vorpalengineering/gundler/internal/mempool"
	"github.com/vorpalengineering/gundler/internal/processor"
	"github.com/vorpalengineering/gundler/internal/simulation"
	"github.com/vorpalengineering/gundler/pkg/types"
)

type RPCServer struct {
	server               *http.Server
	ethClient            *ethclient.Client
	mempools             map[string]*mempool.Mempool // entryPointAddress => Mempool
	processors           map[string]processor.Processor
	chainID              *big.Int
	supportedEntryPoints []string
	mode                 string
}

func NewRPCServer(
	port uint,
	ethRPC string,
	supportedEntryPoints []string,
	mode string,
	maxBundleSize uint,
	ethClient *ethclient.Client,
	chainID *big.Int,
	keyPool *keypool.KeyPool,
) (*RPCServer, error) {

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
		normalizedAddress := entryPoint.Hex()
		mempools[normalizedAddress] = mempool.NewMempool(entryPoint, chainID)

		// Create simulator for this processor
		simulator := simulation.NewSimulator(ethClient, chainID)

		// Create processor
		processors[normalizedAddress] = processor.NewBasicProcessor(
			mempools[normalizedAddress],
			ethClient,
			1*time.Second,
			maxBundleSize,
			keyPool,
			simulator,
		)
		if err := processors[normalizedAddress].Start(context.Background()); err != nil {
			log.Fatalf("Failed to start processor: %v", err)
		}

		log.Printf("Initialized mempool and processor for entry point: %s", normalizedAddress)
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
		mode:                 mode,
	}

	// Register base route
	mux.HandleFunc("/", rpc.handleRPCRequest)

	// Log debug methods availability
	if mode == "DEBUG" {
		log.Println("Debug RPC methods enabled: debug_mempools, debug_pause, debug_clear")
	}

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
	var req types.RPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rpc.sendError(w, nil, -32700, "Parse error")
		return
	}

	// Route to appropriate handler
	var result any
	var err *types.RPCError

	// Handle standard RPC methods
	switch req.Method {
	case "eth_chainId":
		result, err = rpc.handleChainId()
	case "eth_supportedEntryPoints":
		result, err = rpc.handleSupportedEntryPoints()
	case "eth_sendUserOperation":
		result, err = rpc.handleSendUserOperation(req.Params)
	default:
		// Check for debug methods if in DEBUG mode
		if rpc.mode == "DEBUG" {
			switch req.Method {
			case "debug_mempools":
				result, err = rpc.handleDebugMempools()
			case "debug_pause":
				result, err = rpc.handleDebugPause()
			case "debug_clear":
				result, err = rpc.handleDebugClear()
			default:
				err = &types.RPCError{
					Code:    -32601,
					Message: "Method not found",
				}
			}
		} else {
			err = &types.RPCError{
				Code:    -32601,
				Message: "Method not found",
			}
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
	resp := &types.RPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (rpc *RPCServer) sendError(w http.ResponseWriter, id any, code int, message string) {
	resp := &types.RPCResponse{
		JSONRPC: "2.0",
		Error: &types.RPCError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (rpc *RPCServer) handleChainId() (string, *types.RPCError) {
	return fmt.Sprintf("0x%x", rpc.chainID), nil
}

func (rpc *RPCServer) handleSupportedEntryPoints() ([]string, *types.RPCError) {
	return rpc.supportedEntryPoints, nil
}

func (rpc *RPCServer) handleSendUserOperation(params json.RawMessage) (string, *types.RPCError) {
	// Parse json params
	var rawParams []json.RawMessage
	if err := json.Unmarshal(params, &rawParams); err != nil {
		return "", &types.RPCError{
			Code:    -32602,
			Message: "Invalid parameters",
		}
	}

	if len(rawParams) != 2 {
		return "", &types.RPCError{
			Code:    -32602,
			Message: "Expected 2 parameters: [userOp, entryPoint]",
		}
	}

	// Parse UserOperation from rawParams[0]
	var userOp types.UserOperation
	if err := json.Unmarshal(rawParams[0], &userOp); err != nil {
		return "", &types.RPCError{
			Code:    -32602,
			Message: "Error unmarshalling userOp",
		}
	}

	// Parse EntryPoint address from rawParams[1]
	var entryPointStr string
	if err := json.Unmarshal(rawParams[1], &entryPointStr); err != nil {
		return "", &types.RPCError{
			Code:    -32602,
			Message: "Error unmarshalling entryPoint",
		}
	}
	entryPoint := common.HexToAddress(entryPointStr)

	// Validate EntryPoint address
	err := types.ValidateEntryPointAddress(entryPoint)
	if err != nil {
		return "", &types.RPCError{
			Code:    -32602,
			Message: "Invalid EntryPoint Address",
		}
	}

	// TODO: Validate UserOperation

	// Add to mempool (use normalized address for lookup)
	normalizedAddress := entryPoint.Hex()

	// Check if mempool exists
	mempool, exists := rpc.mempools[normalizedAddress]
	if !exists {
		// Log all available mempool keys for debugging
		log.Printf("Mempool not found! Available mempool keys:")
		for key := range rpc.mempools {
			log.Printf("  - %s", key)
		}
		return "", &types.RPCError{
			Code:    -32602,
			Message: fmt.Sprintf("No mempool found for entry point: %s", normalizedAddress),
		}
	}

	if err := mempool.Add(&userOp); err != nil {
		return "", &types.RPCError{
			Code:    -32602,
			Message: fmt.Sprintf("Failed adding userOp to mempool: %v", err),
		}
	}

	// Calculate userOp hash
	userOpHash := userOp.Hash(entryPoint, rpc.chainID)

	log.Printf("UserOp %s validated and added to mempool. Mempool size: %v", userOpHash.Hex(), mempool.Size())

	return userOpHash.Hex(), nil
}
