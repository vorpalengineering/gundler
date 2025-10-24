package processor

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/vorpalengineering/gundler/internal/mempool"
	"github.com/vorpalengineering/gundler/internal/types"
)

type BasicProcessor struct {
	mempool     *mempool.Mempool
	ethClient   *ethclient.Client
	interval    time.Duration
	stopChannel chan struct{}
	doneChannel chan struct{}
}

func NewBasicProcessor(
	mempool *mempool.Mempool,
	ethClient *ethclient.Client,
	interval time.Duration,
) *BasicProcessor {
	return &BasicProcessor{
		mempool:     mempool,
		ethClient:   ethClient,
		interval:    interval,
		stopChannel: make(chan struct{}),
		doneChannel: make(chan struct{}),
	}
}

func (processor *BasicProcessor) Start(ctx context.Context) error {
	log.Printf("Starting Basic Processor with %v interval", processor.interval)

	go processor.run(ctx)

	return nil
}

func (processor *BasicProcessor) Stop() error {
	log.Println("Stopping Basic Processor...")
	close(processor.stopChannel)
	<-processor.doneChannel
	log.Println("Basic Processor Stopped")
	return nil
}

func (processor *BasicProcessor) run(ctx context.Context) {
	defer close(processor.doneChannel)

	ticker := time.NewTicker(processor.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-processor.stopChannel:
			return
		case <-ticker.C:
			if err := processor.processOnce(ctx); err != nil {
				log.Printf("Processing error: %v", err)
			}
		}
	}
}

func (processor *BasicProcessor) processOnce(ctx context.Context) error {
	// Check mempool size
	mempoolSize := processor.mempool.Size()
	if mempoolSize == 0 {
		return nil
	}

	// Create Bundle
	userOpRange, err := processor.mempool.GetRange(0, mempoolSize)
	if err != nil {
		return fmt.Errorf("error getting userOp range: %v", err)
	}
	bundle := processor.createBundle(userOpRange)

	// TODO: simulate bundle

	// Submit Bundle to Chain
	err = processor.submitBundle(ctx, bundle)
	if err != nil {
		return fmt.Errorf("error submitting bundle: %v", err)
	}

	return nil
}

func (processor *BasicProcessor) createBundle(userOps []*types.UserOperation) *Bundle {
	return &Bundle{
		UserOps:    userOps,
		EntryPoint: processor.mempool.EntryPoint,
	}
}

func (processor *BasicProcessor) simulateBundle(ctx context.Context, bundle *Bundle) error {
	return nil
}

func (processor *BasicProcessor) submitBundle(ctx context.Context, bundle *Bundle) error {
	// TODO: submit to chain and get result

	// Remove userOp range from mempool
	// processor.mempool.RemoveByIndexRange()
	return nil
}
