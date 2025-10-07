package types

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

type EntryPoint struct {
	Address common.Address
	Version string
}

// Deterministic EntryPoint addresses
var (
	EntryPointV06Address = common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")
	EntryPointV07Address = common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032")
	EntryPointV08Address = common.HexToAddress("0x4337084D9E255Ff0702461CF8895CE9E3b5Ff108")
)

func NewEntryPoint(version string) (*EntryPoint, error) {
	var entryPointAddress common.Address

	// Select address based on version
	switch version {
	case "V06":
		entryPointAddress = EntryPointV06Address
	case "V07":
		entryPointAddress = EntryPointV07Address
	case "V08":
		entryPointAddress = EntryPointV08Address
	default:
		return nil, fmt.Errorf("unsupported EntryPoint version: %s", version)
	}

	// Construct EntryPoint
	return &EntryPoint{
		Address: entryPointAddress,
		Version: version,
	}, nil
}
