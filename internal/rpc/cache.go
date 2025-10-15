package rpc

import (
	"math/big"
	"sync"
)

type Cache struct {
	mutex   sync.RWMutex
	chainID *big.Int
}

func NewCache() *Cache {
	return &Cache{}
}

func (cache *Cache) SetChainID(chainID *big.Int) {
	// Acquire write lock
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.chainID = chainID
}

func (cache *Cache) GetChainID() *big.Int {
	// Acquire read lock
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	return cache.chainID
}
