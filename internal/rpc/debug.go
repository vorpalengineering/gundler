package rpc

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/vorpalengineering/gundler/internal/types"
)

func (rpc *RPCServer) handleDebugMempools(w http.ResponseWriter, r *http.Request) {
	// Restrict to GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	// Return JSON response
	response := map[string]interface{}{
		"mempools": mempools,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (rpc *RPCServer) handleDebugPause(w http.ResponseWriter, r *http.Request) {
	// Restrict to POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	// Return JSON response with new state
	response := map[string]interface{}{
		"paused": !isPaused,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (rpc *RPCServer) handleDebugClear(w http.ResponseWriter, r *http.Request) {
	// Restrict to POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Clear all mempools
	clearedCount := 0
	for _, mempool := range rpc.mempools {
		mempool.Clear()
		log.Printf("Mempool %s cleared", mempool.EntryPoint.Hex())
		clearedCount++
	}

	log.Printf("Cleared %d mempools", clearedCount)

	// Return JSON response
	response := map[string]interface{}{
		"cleared": clearedCount,
		"message": fmt.Sprintf("%d mempools cleared", clearedCount),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
