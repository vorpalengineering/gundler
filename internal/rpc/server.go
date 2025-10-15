package rpc

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

type RPCServer struct {
	server *http.Server
}

func New() (*RPCServer, error) {
	// Initialize multiplexer
	mux := http.NewServeMux()

	// Register healthcheck route
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Received healthcheck request")
		fmt.Printf("Header: %v\n", r.Header)

		// Write response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Register base route
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Gundler RPC Server"))
	})

	return &RPCServer{
		server: &http.Server{
			Addr:    "localhost:8080",
			Handler: mux,
		},
	}, nil
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
