package processor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/vorpalengineering/gundler/internal/keypool"
	"github.com/vorpalengineering/gundler/internal/mempool"
	"github.com/vorpalengineering/gundler/internal/simulation"
	"github.com/vorpalengineering/gundler/pkg/types"
)

type BasicProcessor struct {
	mempool       *mempool.Mempool
	ethClient     *ethclient.Client
	interval      time.Duration
	stopChannel   chan struct{}
	doneChannel   chan struct{}
	paused        bool
	pauseMutex    sync.RWMutex
	maxBundleSize uint
	keyPool       *keypool.KeyPool
	simulator     *simulation.Simulator
}

func NewBasicProcessor(
	mempool *mempool.Mempool,
	ethClient *ethclient.Client,
	interval time.Duration,
	maxBundleSize uint,
	keyPool *keypool.KeyPool,
	simulator *simulation.Simulator,
) *BasicProcessor {
	return &BasicProcessor{
		mempool:       mempool,
		ethClient:     ethClient,
		interval:      interval,
		stopChannel:   make(chan struct{}),
		doneChannel:   make(chan struct{}),
		maxBundleSize: maxBundleSize,
		keyPool:       keyPool,
		simulator:     simulator,
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
	// Check if paused
	if processor.IsPaused() {
		return nil
	}

	// Check mempool size
	mempoolSize := processor.mempool.Size()
	if mempoolSize == 0 {
		return nil
	}

	// Calculate bundle size (min of mempool size and max bundle size)
	bundleSize := int(processor.maxBundleSize)
	if mempoolSize < bundleSize {
		bundleSize = mempoolSize
	}

	// Get userops from mempool by range
	userOps, err := processor.mempool.GetRange(0, bundleSize)
	if err != nil {
		return fmt.Errorf("error getting userOps by range: %v", err)
	}

	// Create Bundle from mempool userops
	bundle := processor.createBundle(userOps)

	// Simulate bundle
	simResult, err := processor.simulator.SimulateBundle(ctx, bundle)
	if err != nil {
		return fmt.Errorf("error simulating bundle: %v", err)
	}

	if !simResult.Success {
		log.Printf("Bundle simulation failed: %v", simResult.Error)
		return fmt.Errorf("bundle simulation failed: %v", simResult.Error)
	}

	log.Printf("Bundle simulation successful, gas used: %d", simResult.GasUsed)

	// Submit Bundle to Chain
	err = processor.submitBundle(ctx, bundle, bundleSize)
	if err != nil {
		return fmt.Errorf("error submitting bundle: %v", err)
	}

	return nil
}

func (processor *BasicProcessor) createBundle(userOps []*types.UserOperation) *simulation.Bundle {
	return &simulation.Bundle{
		UserOps:    userOps,
		EntryPoint: processor.mempool.EntryPoint,
	}
}

func (processor *BasicProcessor) simulateBundle(ctx context.Context, bundle *simulation.Bundle) error {
	return nil
}

func (processor *BasicProcessor) submitBundle(ctx context.Context, bundle *simulation.Bundle, bundleSize int) error {
	log.Printf("Submitting bundle to chain... size: %v", len(bundle.UserOps))

	// TODO: Build actual transaction for EntryPoint.handleOps() call
	// For now, this is a placeholder that demonstrates keypool usage
	// In a real implementation, you would:
	// 1. Pack the userOps into the EntryPoint.handleOps() call data
	// 2. Create a transaction with proper gas limits and values
	// 3. Submit via keypool

	// Placeholder: We'll add actual transaction building in a future update
	log.Printf("TODO: Build and submit transaction for %d userOps to EntryPoint %s",
		len(bundle.UserOps), bundle.EntryPoint.Hex())

	// Example of how keypool will be used:
	// tx := ethtypes.NewTransaction(...)
	// txHash, keyAddress, err := processor.keyPool.SubmitTransaction(ctx, tx)
	// if err != nil {
	//     return fmt.Errorf("failed to submit bundle transaction: %w", err)
	// }
	// defer processor.keyPool.ReleaseKey(keyAddress)
	// log.Printf("Bundle submitted: tx=%s, key=%s", txHash.Hex(), keyAddress.Hex())

	// Remove bundled userOps from mempool
	err := processor.mempool.RemoveByIndexRange(0, bundleSize)
	if err != nil {
		return fmt.Errorf("error removing bundled userops: %v", err)
	}

	return nil
}

func (processor *BasicProcessor) Pause() {
	processor.pauseMutex.Lock()
	defer processor.pauseMutex.Unlock()

	if !processor.paused {
		processor.paused = true
		log.Println("Basic Processor Paused")
	}
}

func (processor *BasicProcessor) Unpause() {
	processor.pauseMutex.Lock()
	defer processor.pauseMutex.Unlock()

	if processor.paused {
		processor.paused = false
		log.Println("Basic Processor Unpaused")
	}
}

func (processor *BasicProcessor) IsPaused() bool {
	processor.pauseMutex.RLock()
	defer processor.pauseMutex.RUnlock()

	return processor.paused
}
