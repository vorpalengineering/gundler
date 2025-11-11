package processor

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vorpalengineering/gundler/pkg/types"
)

type Processor interface {
	Start(ctx context.Context) error
	Stop() error
	Pause()
	Unpause()
	IsPaused() bool
}

type Bundle struct {
	UserOps    []*types.UserOperation
	EntryPoint common.Address
}

type SimulationResult struct {
	Success bool
	Error   error
	GasUsed uint64
}
