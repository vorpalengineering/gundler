package mempool

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vorpalengineering/gundler/pkg/types"
)

type Mempool struct {
	mutex         sync.RWMutex
	userOps       []*types.UserOperation
	userOpsByHash map[common.Hash]*types.UserOperation
	EntryPoint    common.Address
	ChainID       *big.Int
	// userOpsBySender map[common.Address]*types.UserOperation
}

func NewMempool(entryPoint common.Address, chainID *big.Int) *Mempool {
	return &Mempool{
		userOps:       make([]*types.UserOperation, 0),
		userOpsByHash: make(map[common.Hash]*types.UserOperation, 0),
		EntryPoint:    entryPoint,
		ChainID:       chainID,
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

	// Check for duplicates
	userOpHash := userOp.Hash(pool.EntryPoint, pool.ChainID)
	_, exists := pool.userOpsByHash[userOpHash]
	if exists {
		return fmt.Errorf("duplicate userOp: %v", userOpHash)
	}

	// TODO: Check pending userOps from sender

	// Append userop to array
	pool.userOps = append(pool.userOps, userOp)
	pool.userOpsByHash[userOpHash] = userOp

	return nil
}

func (pool *Mempool) RemoveByIndex(index int) error {
	// Acquire write lock
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	// Validate index
	if index < 0 || index >= len(pool.userOps) {
		return fmt.Errorf("invalid index: %s", index)
	}

	// Remove userOp by hash
	userOpHash := pool.userOps[index].Hash(pool.EntryPoint, pool.ChainID)
	_, exists := pool.userOpsByHash[userOpHash]
	if exists {
		delete(pool.userOpsByHash, userOpHash)
	}

	// Remove userOp at index
	pool.userOps = append(pool.userOps[:index], pool.userOps[index+1:]...)

	return nil
}

func (pool *Mempool) RemoveByIndexRange(begin int, end int) error {
	// Acquire write lock
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	// Validate range bounds
	if begin < 0 || end > len(pool.userOps) || begin > end {
		return fmt.Errorf("invalid range bounds: begin=%d, end=%d, length=%d", begin, end, len(pool.userOps))
	}

	// Remove userOps from hash map
	for i := begin; i < end; i++ {
		userOpHash := pool.userOps[i].Hash(pool.EntryPoint, pool.ChainID)
		delete(pool.userOpsByHash, userOpHash)
	}

	// Remove userOps from slice
	pool.userOps = append(pool.userOps[:begin], pool.userOps[end:]...)

	return nil
}

func (pool *Mempool) GetByIndex(index int) (*types.UserOperation, error) {
	// Acquire read lock
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	// Validate index
	if index < 0 || index >= len(pool.userOps) {
		return nil, fmt.Errorf("index out of range")
	}

	return pool.userOps[index], nil
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
	// Acquire read lock
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	// Validate range bounds
	if begin > end || end > len(pool.userOps) {
		return nil, fmt.Errorf("invalid range bounds")
	}

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
