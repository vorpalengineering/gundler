package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vorpalengineering/gundler/internal/config"
	"github.com/vorpalengineering/gundler/internal/rpc"
)

func main() {
	fmt.Println("Starting Gundler...")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Print config
	cfg.Print()

	// Start RPC Server
	rpc, err := rpc.NewRPCServer(cfg.Port, cfg.EthereumRPC)
	if err != nil {
		log.Fatalf("Failed to create RPC Server: %v", err)
	}
	if err := rpc.Start(); err != nil {
		log.Fatalf("Failed to start RPC Server: %v", err)
	}

	// Start processor
	// processor := processor.NewBasicProcessor(mempool, ethClient, entryPoint, 1*time.Second)
	// if err := processor.Start(context.Background()); err != nil {
	// 	log.Fatalf("Failed to start processor: %v", err)
	// }

	fmt.Println("Gundler Startup Complete")

	// Wait for interrupt signal
	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel

	fmt.Println("\nShutting down Gundler...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// if err := processor.Stop(); err != nil {
	// 	log.Printf("Failed to stop processor: %v", err)
	// }

	if err := rpc.Shutdown(ctx); err != nil {
		log.Fatalf("RPC Server forced to shutdown: %v", err)
	}

	fmt.Println("Gundler stopped")
}
