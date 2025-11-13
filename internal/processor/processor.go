package processor

import (
	"context"
)

type Processor interface {
	Start(ctx context.Context) error
	Stop() error
	Pause()
	Unpause()
	IsPaused() bool
}
