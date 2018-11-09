# mainstay
The mainstay repository is an application that implements the [Mainstay protocol](https://www.commerceblock.com/wp-content/uploads/2018/03/commerceblock-mainstay-whitepaper.pdf) designed by [CommerceBlock](https://www.commerceblock.com). It is a Go daemon performing attestations of the [Ocean](https://github.com/commerceblock/ocean) network to Bitcoin. It also includes a Confirmation tool that can be run in parallel to the Ocean/Bitcoin networks and confirm these attestations.

## Prerequisites
* Go (https://github.com/golang)
* Bitcoin (https://github.com/bitcoin/bitcoin)

## Instructions

### Attestation Service

- Install go and the attestation service by following `scripts/build.sh`

- Regtest Demo
    - Start the ocean-demo `python ocean-demo/PROTOCOLS/demo.py`
    - Run `mainstay -regtest`

- Testnet Mode

    - Download and run a full Bitcoin Node on testnet mode, fully indexed and in blocksonly mode. `signrawtransaction` should also be included as a `deprecatedrpc` option. Add the connection details (actual value or ENV variable) to this node in `conf/conf.json`

    - Fund this wallet node, send all the funds to a single address and store the `TX_HASH` and `PRIVKEY` of this transaction.

    - The `TX_HASH` should be included in the genesis block of the Ocean network using the conf option `attestationhash`. Initiate the Ocean testnet network and add connection details (actual value or ENV variable) to one of it's nodes in `conf/conf.json`.

    - Run `mainstay -tx TX_HASH`

- Unit Testing
    - `/$GOPATH/src/mainstay/run-tests.sh`

### Confirmation Tool

The confirmation tool `cmd/confirmationtool` can be used to confirm all the attestations of the Ocean network to Bitcoin and wait for any new attestations that will be happening.

Running this tool will require a full Bitcoin testnet node and a full Ocean node. Connection details for these should be included in `cmd/confirmationtool/conf.json`. To run this tool you need to first fetch the `TX_HASH` from the `attestationhash` field in the Ocean genesis block and then run:

`go run cmd/confirmationtool/confirmationtool.go -tx TX_HASH`

This will initially take some time to sync up all the attestations that have been committed so far and then will wait for any new attestations. Logging is displayed for each attestation and for full details the `-detailed` flag can be used.
