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
go run cmd/main.go --rpc https://rpc.testnet.telos.net
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
```