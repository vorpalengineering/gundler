package mempool

import (
	"fmt"
	"sync"

	"github.com/vorpalengineering/gundler/internal/types"
)

type Mempool struct {
	mutex   sync.RWMutex
	userOps []*types.UserOperation
	// userOpsByHash map[common.Hash]int
	// userOpsBySender map[common.Address]*types.UserOperation
}

func NewMempool() *Mempool {
	return &Mempool{
		userOps: make([]*types.UserOperation, 0),
	}
}

func (pool *Mempool) Add(userOp *types.UserOperation) error {
	// Acquire write lock
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	// Validate user operation
	if err := pool.validateUserOp(userOp); err != nil {
		return fmt.Errorf("userOp validation failed: %w", err)
	}

	// TODO: Check for duplicates
	// TODO: Check pending userOps from sender

	// Append userop to array
	pool.userOps = append(pool.userOps, userOp)

	return nil
}

func (pool *Mempool) RemoveByIndex(index int) {
	// Acquire write lock
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if index >= 0 && index < len(pool.userOps) {
		pool.userOps = append(pool.userOps[:index], pool.userOps[index+1:]...)
	}
}

func (pool *Mempool) GetAll() []*types.UserOperation {
	// Acquire read lock
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	// Get all user operations in mempool
	// Create a copy to avoid external modifications
	ops := make([]*types.UserOperation, len(pool.userOps))
	copy(ops, pool.userOps)

	return ops
}

func (pool *Mempool) GetRange(begin int, end int) ([]*types.UserOperation, error) {
	// Validate
	if begin > end || end > len(pool.userOps) {
		return nil, fmt.Errorf("invalid range bounds")
	}

	// Acquire read lock
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	// Get user operations in range (inclusive)
	ops := make([]*types.UserOperation, end-begin)
	copy(ops, pool.userOps[begin:end])

	return ops, nil
}

func (pool *Mempool) Clear() {
	// Acquire write lock
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	pool.userOps = make([]*types.UserOperation, 0)
}

func (pool *Mempool) Size() int {
	// Acquire read lock
	pool.mutex.RLock()
	defer pool.mutex.RUnlock()

	return len(pool.userOps)
}
