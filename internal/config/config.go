package config

import (
	"flag"
	"fmt"
)

type Config struct {
	EthereumRPC string `json:"ethereum_rpc"`
	Port        uint   `json:"port"`
	Beneficiary string `json:"beneficiary"`
}

func Load() (*Config, error) {
	// Initialize empty config
	config := &Config{}

	// Define flags
	rpc := flag.String("rpc", "", "Ethereum RPC URL")
	port := flag.Uint("port", 3000, "Port")
	beneficiary := flag.String("beneficiary", "", "Beneficiary Address")

	// Parse all defined flags
	flag.Parse()

	// Set flags into config
	config.EthereumRPC = *rpc
	config.Port = *port
	config.Beneficiary = *beneficiary

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
