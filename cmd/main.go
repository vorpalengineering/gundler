package main

import (
	"fmt"
	"log"

	"github.com/vorpalengineering/gundler/internal/config"
)

func main() {
	fmt.Println("Starting Gundler...")

	// Setup config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Print config
	cfg.Print()

	fmt.Println("Gundler Startup Complete")
}
