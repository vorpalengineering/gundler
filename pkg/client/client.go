package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/vorpalengineering/gundler/pkg/types"
)

type GundlerClientConfig struct {
	ServerURL string
}

type GundlerClient struct {
	httpClient *http.Client
	serverURL  string
	nextID     atomic.Uint64
}

func NewGundlerClient(config GundlerClientConfig) *GundlerClient {
	return &GundlerClient{
		httpClient: &http.Client{},
		serverURL:  config.ServerURL,
	}
}

func (gc *GundlerClient) ChainID(ctx context.Context) (*big.Int, error) {
	resp, err := gc.call(ctx, "eth_chainId", []interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to call eth_chainId: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("RPC error: code=%d, message=%s", resp.Error.Code, resp.Error.Message)
	}

	// Parse hex string result
	hexStr, ok := resp.Result.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", resp.Result)
	}

	// Remove 0x prefix if present
	hexStr = strings.TrimPrefix(hexStr, "0x")

	// Parse hex string to big.Int
	chainID := new(big.Int)
	if _, ok := chainID.SetString(hexStr, 16); !ok {
		return nil, fmt.Errorf("failed to parse chain ID hex string: %s", hexStr)
	}

	return chainID, nil
}

func (gc *GundlerClient) call(ctx context.Context, method string, params interface{}) (*types.RPCResponse, error) {
	// Marshal params to JSON
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}

	// Create RPC request
	req := types.RPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  paramsJSON,
		ID:      gc.nextID.Add(1),
	}

	// Marshal request to JSON
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", gc.serverURL, bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	httpResp, err := gc.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer httpResp.Body.Close()

	// Check HTTP status
	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("unexpected HTTP status: %d, body: %s", httpResp.StatusCode, string(body))
	}

	// Decode response
	var rpcResp types.RPCResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&rpcResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &rpcResp, nil
}
