package mempool

import (
	"sync"

	"github.com/vorpalengineering/gundler/internal/types"
)

type Mempool struct {
	mutex   sync.RWMutex
	userOps []*types.UserOperation
	// userOpHashes map[common.Hash]int // UserOperation Hash => mempool index
}

func New() *Mempool {
	return &Mempool{
		userOps: make([]*types.UserOperation, 0),
	}
}

func (pool *Mempool) Add(userOp *types.UserOperation) error {
	// Acquire write lock
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	// TODO: Validate user operation
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
