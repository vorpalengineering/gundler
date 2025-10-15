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
go run cmd/main.go --rpc https://rpc.testnet.telos.net --chain-id 41
```

### Flags

| Flag | Name | Default |
| :------- | :------: | -------: |
| --rpc     | RPC URL   | None    |
| --chain-id   | Chain ID   | None   |

### Curl Commands

```bash
# Healthcheck
curl http://localhost:8080/health
```