# mainstay
The mainstay repository is an application that implements the [Mainstay protocol](https://www.commerceblock.com/wp-content/uploads/2018/03/commerceblock-mainstay-whitepaper.pdf) designed by [CommerceBlock](https://www.commerceblock.com). It is a Go daemon performing attestations of the [Ocean](https://github.com/commerceblock/ocean) network and other client commitments to Bitcoin. It also includes a Confirmation tool that can be run in parallel to the Bitcoin network in order confirm the attestations and prove that the commitments were included in the attestation.

## Prerequisites
* Go (https://github.com/golang)
* Bitcoin (https://github.com/bitcoin/bitcoin)
* Zmq (https://github.com/zeromq/libzmq)

## Instructions

### Attestation Service

- Install go and the attestation service by following `scripts/build.sh`

- Regtest Demo
    - Start the ocean-demo `python ocean-demo/PROTOCOLS/demo.py`
    - Run `mainstay -regtest`

- Testnet Mode

    - Download and run a full Bitcoin Node on testnet mode, fully indexed and in blocksonly mode. `signrawtransaction` should also be included as a `deprecatedrpc` option. Add the connection details (actual value or ENV variable) to this node in `conf/conf.json`

    - Fund this wallet node, send all the funds to a single (m of n sig) P2SH address and store the `TX_HASH`, `PRIVKEY_x` and `REDEEM_SCRIPT` of this transaction, where `x in [0, n-1]`.

    - The `TX_HASH` should be included in the genesis block of the Ocean network using the conf option `attestationhash`. Initiate the Ocean testnet network and add connection details (actual value or ENV variable) to one of it's nodes in `conf/conf.json`.

    - Run `mainstay -tx TX_HASH -pk PRIVKEY_0 -script REDEEM_SCRIPT` for the main client and `go run cmd/txsigningtool/txsigningtool.go -tx TX_HASH -pk PRIVKEY_x -script REDEEMSCRIPT` for each remaining (n-1) transaction signer of the m-of-n multisig P2SH address

- Unit Testing
    - `/$GOPATH/src/mainstay/run-tests.sh`

### Tools

#### Transaction Signing Tool

The transaction signing tool `cmd/txsigningtool` is a tool used by each signer of the attestation multisig to sign attestation transactions. The tool subscribes to the main attestation service in order to receive confirmed attestation hashes for tweaking priv key before signing, new commitment hashes to verify new destination address and new bitcoin attestation transactions to sign and publish signature for attestation service to receive.

The tool maintains a connection to a Bitcoin Node for transaction validation and signing.

#### Client Signup Tool

The client signup tool `cmd/clientsignuptool` is used by the maintainer of the attestation service to sign up new clients. The client will provide a P2PKH address that will be required to verify the signature of new commitments sent to the attestation API. The tool assigns a new position to the client in the CMR and provides an auth_token for authorizing API POST requests of new commitments submitted by the client.

The tool requires a connection to the DB used by the service with appropriate permissions.

For auth-token generation only, token generator tool `cmd/tokengeneratortool` can be used.

#### Client Confirmation Tool

The confirmation tool `cmd/confirmationtool` can be used to confirm all the attestations of a client Ocean-type network to Bitcoin and wait for any new attestations that will be happening.

Running this tool will require a full Bitcoin testnet node and a full Ocean node. Connection details for these should be included in `cmd/confirmationtool/conf.json`. To run this tool you need to first fetch the `TX_HASH` from the `attestationhash` field in the Ocean genesis block and then run:

`go run cmd/confirmationtool/confirmationtool.go -tx TX_HASH`

This will initially take some time to sync up all the attestations that have been committed so far and then will wait for any new attestations. Logging is displayed for each attestation and for full details the `-detailed` flag can be used.
