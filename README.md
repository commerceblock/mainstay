# mainstay
The mainstay repository is an application that implements the [Mainstay protocol](https://www.commerceblock.com/wp-content/uploads/2018/03/commerceblock-mainstay-whitepaper.pdf) designed by [CommerceBlock](https://www.commerceblock.com). It is a Go daemon performing attestations of the [Ocean](https://github.com/commerceblock/ocean) network and other client commitments to Bitcoin. It also includes a Confirmation tool that can be run in parallel to the Bitcoin network in order confirm the attestations and prove that the commitments were included in the attestation.

# Prerequisites
* Go (https://github.com/golang)
* Bitcoin (https://github.com/bitcoin/bitcoin)
* Zmq (https://github.com/zeromq/libzmq)

# Instructions

## Attestation Service

- Install Go and the attestation service by following `scripts/build.sh`

- Setup up database collections and roles using `scripts/db-init.js`

- Setup `conf.json` file under `/config` by following [config guidelines](/config/README.md)

- Run service
    - Regtest
        - Run service: `mainstay -regtest`
        - Run signer: `go run $GOPATH/src/mainstay/cmd/txsigningtool/txsigningtool.go -regtest`
        - Insert commitments to "ClientCommitment" database collection in order to generate new attestations
    - Testnet
        - Download and run a full Bitcoin Node on testnet mode, fully indexed and in blocksonly mode.

        - Fund this wallet node, send all the funds to a single (m of n sig) P2SH address and store the `TX_HASH`, `PRIVKEY_x` and `REDEEM_SCRIPT` of this transaction, where `x in [0, n-1]`.

        - Follow the same procedure to generate a single (m of n sig) P2SH address used to topup the service and store the `TOPUP_ADDRESS`, `TOPUP_PRIVKEY_x` and `TOPUP_SCRIPT`.

        - The `TX_HASH` should be included in the genesis block of the client Ocean-type network using the config option `attestationhash`.

        - Run the mainstay attestation service by:

            `mainstay`

            Command line parameters should be set in `.conf` file

        - Run transaction signers of the m-of-n multisig P2SH addresses for `x in [0, n-1]` by:

            `go run $GOPATH/src/mainstay/cmd/txsigningtool/txsigningtool.go -pk PRIVKEY_x -pkTopup TOPUP_PRIVKEY_x -host SIGNER_HOST`

            Command line parameters should be set in the corresponding signer `.conf` file


- Unit Testing
    - `/$GOPATH/src/mainstay/run-tests.sh`

## Tools

### Transaction Signing Tool

The transaction signing tool `cmd/txsigningtool` is a tool used by each signer of the attestation multisig to sign attestation transactions. The tool subscribes to the main attestation service in order to receive confirmed attestation hashes for tweaking priv key before signing, new commitment hashes to verify new destination address and new bitcoin attestation transaction pre-images to sign and publish signature for attestation service to receive.

The tool uses ECDSA libraries to do the signing and does not require connection to a Bitcoin node.

### Client Signup Tool

The client signup tool `cmd/clientsignuptool` is used by the maintainer of the attestation service to sign up new clients. The client will provide a P2PKH address that will be required to verify the signature of new commitments sent to the attestation API. The tool assigns a new position to the client in the CMR and provides an auth_token for authorizing API POST requests of new commitments submitted by the client.

The tool requires a connection to the DB used by the service with appropriate permissions.

For auth-token generation only, token generator tool `cmd/tokengeneratortool` can be used.

### Client Confirmation Tool

The confirmation tool `cmd/confirmationtool` can be used to confirm all the attestations of a client Ocean-type network to Bitcoin and wait for any new attestations that will be happening.

Running this tool will require a full Bitcoin testnet node and a full Ocean node. Connection details for these should be included in `cmd/confirmationtool/conf.json`.

The `API_HOST` field should be set to the mainstay URL. This can be updated in `cmd/confirmationtool/confirmationtool.go`.

To run this tool you need to first fetch the `TX_HASH` from the `attestationhash` field in the Ocean genesis block, as well as the publicly available `REDEEM_SCRIPT` of the attestation service multisig. The tool can also be started with any other `TX_HASH` attestation found in the mainstay website. A client should use his designated `CLIENT_POSITION` that was assigned during signup and run the tool using:

`go run cmd/confirmationtool/confirmationtool.go -tx TX_HASH -script REDEEM_SCRIPT -position CLIENT_POSITION -apiHost https://mainstay.xyz`

This will initially take some time to sync up all the attestations that have been committed so far and then will wait for any new attestations. Logging is displayed for each attestation and for full details the `-detailed` flag can be used.
