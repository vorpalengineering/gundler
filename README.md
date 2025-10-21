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

```bash
go run cmd/main.go --rpc https://rpc.testnet.telos.net --beneficiary 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
```

### Flags

| Flag | Name | Default |
| :------- | :------: | -------: |
| --rpc | RPC URL | None  |
| --port | Port | 3000 |

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