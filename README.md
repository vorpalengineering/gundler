# Gundler

## Universal ERC4337 Go Bundler

Gundler is...

* High Performance
* Low Footprint
* Minimal Dependency
* ERC4337 Spec Compliant

### RPC Methods

```
eth_chainId
eth_supportedEntryPoints
eth_sendUserOperation
eth_getUserOperationReceipt
```

### Setup

1. Create a config file (or copy from the example):
```bash
cp example.config.json config.json
```

2. Edit the config file with your settings:
```json
{
  "ethereum_rpc": "https://rpc.testnet.telos.net",
  "port": 3000,
  "beneficiary": "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
}
```

3. Run gundler:
```bash
go run cmd/main.go
```

Or specify a custom config file path:
```bash
go run cmd/main.go --config /path/to/your/config.json
```

### Flags

| Flag | Description | Default |
| :------- | :------: | -------: |
| --config | Path to JSON config file | ./config.json |

### Config File Format

The config file must be a JSON file with the following fields:

| Field | Type | Required | Description |
| :------- | :------: | :-------: | :------- |
| ethereum_rpc | string | Yes | Ethereum RPC URL |
| port | number | No | Port to run the server on (default: 3000) |
| beneficiary | string | Yes | Beneficiary address |

### Curl Commands

```bash
# Healthcheck
curl http://localhost:3000/health

# eth_chainId Method
curl -X POST http://localhost:3000 \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}'

# eth_supportedEntryPoints Method
curl -X POST http://localhost:3000 \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","method":"eth_supportedEntryPoints","params":[],"id":2}'

# eth_sendUserOperation Method
curl -X POST http://localhost:3000 \
    -H "Content-Type: application/json" \
    -d '{
      "jsonrpc": "2.0",
      "method": "eth_sendUserOperation",
      "params": [
        {
          "sender": "0x1234567890123456789012345678901234567890",
          "nonce": "0x1",
          "factory": "0x0000000000000000000000000000000000000000",
          "factoryData": "0x",
          "callData": "0xabcdef",
          "callGasLimit": "0x10000",
          "verificationGasLimit": "0x20000",
          "preVerificationGas": "0x5000",
          "maxFeePerGas": "0x3b9aca00",
          "maxPriorityFeePerGas": "0x3b9aca00",
          "paymaster": "0x0000000000000000000000000000000000000000",
          "paymasterVerificationGasLimit": "0x0",
          "paymasterPostOpGasLimit": "0x0",
          "paymasterData": "0x",
          "signature": "0x1234"
        },
        "0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789"
      ],
      "id": 1
    }'
```