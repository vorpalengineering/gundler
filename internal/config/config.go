package config

import (
	"flag"
	"fmt"
)

type Config struct {
	EthereumRPC string `json:"ethereum_rpc"`
	Port        uint   `json:"port"`
}

func Load() (*Config, error) {
	// Initialize empty config
	config := &Config{}

	// Define flags
	rpc := flag.String("rpc", "", "Ethereum RPC URL")
	port := flag.Uint("port", 3000, "Port")

	// Parse all defined flags
	flag.Parse()

	// Set flags into config
	config.EthereumRPC = *rpc
	config.Port = *port

	// Validate config
	err := config.Validate()
	if err != nil {
		return nil, fmt.Errorf("error validating config")
	}

	return config, nil
}

func (cfg *Config) Validate() error {
	// Check required flags
	if cfg.EthereumRPC == "" {
		return fmt.Errorf("ethereum rpc is required")
	}

	return nil
}

func (cfg *Config) Print() {
	// Print config summary
	fmt.Println("Gundler Config:")
	fmt.Println("===============")
	fmt.Printf("Ethereum RPC: %s\n", cfg.EthereumRPC)
	fmt.Printf("Port: %v\n", cfg.Port)
	fmt.Println("===============")
}
