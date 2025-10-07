package config

import (
	"flag"
	"fmt"
)

type Config struct {
	EthereumRPC string `json:"ethereum_rpc"`
}

func Load() (*Config, error) {
	// Initialize empty config
	config := &Config{}

	// Define flags
	rpc := flag.String("rpc", "", "Ethereum RPC URL")

	// Parse all defined flags
	flag.Parse()

	// Set flags into config
	config.EthereumRPC = *rpc

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
