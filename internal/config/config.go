package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

type Config struct {
	EthereumRPC string `json:"ethereum_rpc"`
	Port        uint   `json:"port"`
	Beneficiary string `json:"beneficiary"`
}

func Load() (*Config, error) {
	// Define config file flag
	configPath := flag.String("config", "./config.json", "Path to config file")

	// Parse flags
	flag.Parse()

	// Read config file
	data, err := os.ReadFile(*configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse JSON
	config := &Config{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Validate config
	err = config.Validate()
	if err != nil {
		return nil, fmt.Errorf("error validating config: %w", err)
	}

	return config, nil
}

func (cfg *Config) Validate() error {
	// Check required flags
	if cfg.EthereumRPC == "" {
		return fmt.Errorf("ethereum rpc flag is required")
	}
	if cfg.Beneficiary == "" {
		return fmt.Errorf("beneficiary flag is required")
	}

	return nil
}

func (cfg *Config) Print() {
	// Print config summary
	fmt.Println("======= Gundler Config ========")
	fmt.Printf("Ethereum RPC: %s\n", cfg.EthereumRPC)
	fmt.Printf("Port: %v\n", cfg.Port)
	fmt.Printf("Beneficiary: %v\n", cfg.Beneficiary)
	fmt.Println("===============================")
}
