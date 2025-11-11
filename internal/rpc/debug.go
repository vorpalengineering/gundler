package rpc

import (
	"fmt"
	"log"

	"github.com/vorpalengineering/gundler/pkg/types"
)

func (rpc *RPCServer) handleDebugMempools() (any, *types.RPCError) {
	// Helper function to get version label from address
	getVersionLabel := func(address string) string {
		switch address {
		case types.EntryPointV06Address.Hex():
			return "MempoolV06"
		case types.EntryPointV07Address.Hex():
			return "MempoolV07"
		case types.EntryPointV08Address.Hex():
			return "MempoolV08"
		default:
			return "MempoolUnknown"
		}
	}

	// Build response with all mempools
	type MempoolInfo struct {
		Label   string                 `json:"label"`
		Address string                 `json:"address"`
		Size    int                    `json:"size"`
		UserOps []*types.UserOperation `json:"userops"`
	}

	mempools := make([]MempoolInfo, 0, len(rpc.mempools))
	for address, mempool := range rpc.mempools {
		mempools = append(mempools, MempoolInfo{
			Label:   getVersionLabel(address),
			Address: address,
			Size:    mempool.Size(),
			UserOps: mempool.GetAll(),
		})
	}

	return mempools, nil
}

func (rpc *RPCServer) handleDebugPause() (any, *types.RPCError) {
	// Check current pause state (check first processor)
	var isPaused bool
	for _, proc := range rpc.processors {
		isPaused = proc.IsPaused()
		break
	}

	// Toggle pause state for all processors
	if isPaused {
		// Unpause all processors
		for _, proc := range rpc.processors {
			proc.Unpause()
		}
		log.Println("All processors unpaused")
	} else {
		// Pause all processors
		for _, proc := range rpc.processors {
			proc.Pause()
		}
		log.Println("All processors paused")
	}

	// Return response with new state
	response := map[string]interface{}{
		"paused": !isPaused,
	}

	return response, nil
}

func (rpc *RPCServer) handleDebugClear() (any, *types.RPCError) {
	// Clear all mempools
	clearedCount := 0
	for _, mempool := range rpc.mempools {
		mempool.Clear()
		log.Printf("Mempool %s cleared", mempool.EntryPoint.Hex())
		clearedCount++
	}

	log.Printf("Cleared %d mempools", clearedCount)

	// Return response
	response := map[string]interface{}{
		"cleared": clearedCount,
		"message": fmt.Sprintf("%d mempools cleared", clearedCount),
	}

	return response, nil
}
