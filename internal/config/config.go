package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vorpalengineering/gundler/internal/types"
)

type GundlerConfig struct {
	EthereumRPC          string   `json:"ethereum_rpc"`
	Port                 uint     `json:"port"`
	Beneficiary          string   `json:"beneficiary"`
	SupportedEntryPoints []string `json:"supported_entry_points"`
}

func Load() (*GundlerConfig, error) {
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
	config := &GundlerConfig{}
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

func (cfg *GundlerConfig) Validate() error {
	// Check required fields
	if cfg.EthereumRPC == "" {
		return fmt.Errorf("ethereum_rpc is required")
	}
	if cfg.Beneficiary == "" {
		return fmt.Errorf("beneficiary is required")
	}
	if len(cfg.SupportedEntryPoints) == 0 {
		return fmt.Errorf("supported_entry_points must contain at least one entry point address")
	}
	for _, epStr := range cfg.SupportedEntryPoints {
		entryPoint := common.HexToAddress(epStr)
		err := types.ValidateEntryPointAddress(entryPoint)
		if err != nil {
			return fmt.Errorf("entrypoint address %s is invalid", epStr)
		}
	}

	return nil
}

func (cfg *GundlerConfig) Print() {
	// Print config summary
	fmt.Println("======= Gundler Config ========")
	fmt.Printf("Ethereum RPC: %s\n", cfg.EthereumRPC)
	fmt.Printf("Port: %v\n", cfg.Port)
	fmt.Printf("Beneficiary: %v\n", cfg.Beneficiary)
	fmt.Printf("Supported Entry Points: %v\n", cfg.SupportedEntryPoints)
	fmt.Println("===============================")
}
