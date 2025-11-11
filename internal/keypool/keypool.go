package keypool

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type PooledKey struct {
	PrivateKey *ecdsa.PrivateKey
	Address    common.Address
	InFlight   bool
}

type KeyPool struct {
	keys      []*PooledKey
	ethClient *ethclient.Client
	chainID   *big.Int
	mutex     sync.Mutex
	cond      *sync.Cond
}

func NewKeyPool(privateKeyStrings []string, ethClient *ethclient.Client, chainID *big.Int) (*KeyPool, error) {
	if len(privateKeyStrings) == 0 {
		return nil, fmt.Errorf("no private keys provided")
	}

	keys := make([]*PooledKey, 0, len(privateKeyStrings))

	for i, pkStr := range privateKeyStrings {
		// Remove 0x prefix if present
		pkStr = strings.TrimPrefix(strings.TrimSpace(pkStr), "0x")

		// Parse private key
		privateKey, err := crypto.HexToECDSA(pkStr)
		if err != nil {
			return nil, fmt.Errorf("invalid private key at index %d: %w", i, err)
		}

		// Derive address from private key
		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("failed to cast public key to ECDSA at index %d", i)
		}
		address := crypto.PubkeyToAddress(*publicKeyECDSA)

		keys = append(keys, &PooledKey{
			PrivateKey: privateKey,
			Address:    address,
			InFlight:   false,
		})

		log.Printf("Loaded key %d: Address: %s", i+1, address.Hex())
	}

	pool := &KeyPool{
		keys:      keys,
		ethClient: ethClient,
		chainID:   chainID,
	}
	pool.cond = sync.NewCond(&pool.mutex)

	log.Printf("KeyPool initialized with %d keys", len(keys))

	return pool, nil
}

func (kp *KeyPool) SubmitTransaction(ctx context.Context, tx *ethtypes.Transaction) (common.Hash, common.Address, error) {
	// Get next available key (blocks if all keys are in-flight)
	key, err := kp.getNextAvailableKey(ctx)
	if err != nil {
		return common.Hash{}, common.Address{}, fmt.Errorf("failed to get available key: %w", err)
	}

	// Mark key as in-flight
	kp.markKeyInFlight(key.Address)

	// Sign transaction
	signedTx, err := ethtypes.SignTx(tx, ethtypes.NewEIP155Signer(kp.chainID), key.PrivateKey)
	if err != nil {
		kp.ReleaseKey(key.Address) // Release key on error
		return common.Hash{}, common.Address{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Submit transaction
	err = kp.ethClient.SendTransaction(ctx, signedTx)
	if err != nil {
		kp.ReleaseKey(key.Address) // Release key on error
		return common.Hash{}, common.Address{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	txHash := signedTx.Hash()
	log.Printf("Transaction submitted: %s from key: %s", txHash.Hex(), key.Address.Hex())

	return txHash, key.Address, nil
}

func (kp *KeyPool) ReleaseKey(address common.Address) {
	kp.mutex.Lock()
	defer kp.mutex.Unlock()

	for _, key := range kp.keys {
		if key.Address == address {
			if key.InFlight {
				key.InFlight = false
				log.Printf("Key released: %s", address.Hex())
				kp.cond.Signal() // Wake up one waiting goroutine
			}
			return
		}
	}
}

func (kp *KeyPool) GetKeyCount() int {
	return len(kp.keys)
}

func (kp *KeyPool) getNextAvailableKey(ctx context.Context) (*PooledKey, error) {
	kp.mutex.Lock()
	defer kp.mutex.Unlock()

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Find first available key
		for _, key := range kp.keys {
			if !key.InFlight {
				return key, nil
			}
		}

		// All keys are in-flight, wait for signal
		log.Println("All keys are in-flight, waiting for available key...")
		kp.cond.Wait()
	}
}

func (kp *KeyPool) markKeyInFlight(address common.Address) {
	kp.mutex.Lock()
	defer kp.mutex.Unlock()

	for _, key := range kp.keys {
		if key.Address == address {
			key.InFlight = true
			log.Printf("Key marked as in-flight: %s", address.Hex())
			return
		}
	}
}
