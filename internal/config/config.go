package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vorpalengineering/gundler/pkg/types"
)

// Mode represents the runtime mode
type Mode string

const (
	ModeDebug Mode = "DEBUG"
	ModeDev   Mode = "DEV"
	ModeProd  Mode = "PROD"
)

// IsValid checks if the mode is valid
func (m Mode) IsValid() bool {
	switch m {
	case ModeDebug, ModeDev, ModeProd:
		return true
	default:
		return false
	}
}

type GundlerConfig struct {
	EthereumRPC          string   `json:"ethereum_rpc"`
	Port                 uint     `json:"port"`
	Beneficiary          string   `json:"beneficiary"`
	SupportedEntryPoints []string `json:"supported_entry_points"`
	Mode                 Mode     `json:"mode"`
	MaxBundleSize        uint     `json:"max_bundle_size"`
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
	if cfg.Mode == "" {
		return fmt.Errorf("mode is required")
	}
	if !cfg.Mode.IsValid() {
		return fmt.Errorf("mode must be one of: DEBUG, DEV, PROD (got: %s)", cfg.Mode)
	}

	// Set default MaxBundleSize if not provided
	if cfg.MaxBundleSize == 0 {
		cfg.MaxBundleSize = 5
	}

	return nil
}

func (cfg *GundlerConfig) Print() {
	// Print config summary
	fmt.Println("======= Gundler Config ========")
	fmt.Printf("Mode: %s\n", cfg.Mode)
	fmt.Printf("Ethereum RPC: %s\n", cfg.EthereumRPC)
	fmt.Printf("Port: %v\n", cfg.Port)
	fmt.Printf("Beneficiary: %v\n", cfg.Beneficiary)
	fmt.Printf("Supported Entry Points: %v\n", cfg.SupportedEntryPoints)
	fmt.Printf("Max Bundle Size: %v\n", cfg.MaxBundleSize)
	fmt.Println("===============================")
}

func LoadPrivateKeys() ([]string, error) {
	privKeysEnv := os.Getenv("GUNDLER_PRIV_KEYS")
	if privKeysEnv == "" {
		return nil, fmt.Errorf("GUNDLER_PRIV_KEYS environment variable is required")
	}

	// Split by comma and trim whitespace
	privKeyStrings := strings.Split(privKeysEnv, ",")
	result := make([]string, 0, len(privKeyStrings))

	for _, pk := range privKeyStrings {
		trimmed := strings.TrimSpace(pk)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("GUNDLER_PRIV_KEYS environment variable is empty")
	}

	return result, nil
}
