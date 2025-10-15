package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type RPCServer struct {
	server *http.Server
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

func New() *RPCServer {
	// Initialize mux handler
	mux := http.NewServeMux()

	// Register healthcheck route
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	rpc := &RPCServer{
		server: &http.Server{
			Addr:    "localhost:8080",
			Handler: mux,
		},
	}

	// Register base route
	mux.HandleFunc("/", rpc.handleRPC)

	return rpc
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
	return rpc.server.Shutdown(ctx)
}

func (rpc *RPCServer) handleRPC(w http.ResponseWriter, r *http.Request) {
	// Restrict to POST requests
	fmt.Println("\nHandling RPC Request")
}
