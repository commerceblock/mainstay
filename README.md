# Mainstay
The mainstay repository is an application that implements the [Mainstay protocol](https://www.commerceblock.com/wp-content/uploads/2018/03/commerceblock-mainstay-whitepaper.pdf) designed by [CommerceBlock](https://www.commerceblock.com). It consists of a Go daemon that performs attestations of the [Ocean](https://github.com/commerceblock/ocean) network along with client commitments to Bitcoin in the form of a commitment merkle tree.

Mainstay is accompanied by a Confirmation tool that can be run in parallel with the Bitcoin network to confirm attestations and prove the commitment inclusion in Mainstay attestations.

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
    - Regtest mode
        - Run service: `mainstay -regtest`
        - Run signer: `go run $GOPATH/src/mainstay/cmd/txsigningtool/txsigningtool.go -regtest`
        - Insert commitments to "ClientCommitment" database collection in order to generate new attestations
    - Testnet/Mainnet mode
        - Download and run a full Bitcoin Node on testnet mode, fully indexed and in blocksonly mode.

        - Fund this wallet node, send all the funds to a single (`m of n` sig) P2SH address and store the `TX_HASH`, `PRIVKEY_x` and `REDEEM_SCRIPT` of this transaction, where `x in [0, n-1]`.

            (In the case of an Ocean-type network the `TX_HASH` should be included in the genesis block using the config option `attestationhash`)

        - Follow the same procedure to generate a single (`m of n` sig) P2SH address used to topup the service and store the `TOPUP_ADDRESS`, `TOPUP_PRIVKEY_x` and `TOPUP_SCRIPT`.

        - Run the mainstay attestation service by:

            `mainstay`

            Command line parameters should be set in `.conf` file

        - Run transaction signers of the m-of-n multisig P2SH addresses for `x in [0, n-1]` by:

            `go run $GOPATH/src/mainstay/cmd/txsigningtool/txsigningtool.go -pk PRIVKEY_x -pkTopup TOPUP_PRIVKEY_x -host SIGNER_HOST`

            Command line parameters should be set in the corresponding signer `.conf` file


- Unit Testing
    - `/$GOPATH/src/mainstay/run-tests.sh`

## Tools

Along with the Mainstay daemon there is various tools offered serving utilities for both Mainstay operators and clients of Mainstay. These tools and their functionality are briefly summarized below:

- Client Confirmation Tool

The confirmation tool `cmd/confirmationtool` can be used to confirm all the attestations of a client Ocean-type network to Bitcoin and wait for any new attestations that will be happening.

- Commitment Tool

The commitment tool `cmd/commitmenttool` can be used to send hash commitments to the Mainstay API.

- Transaction Signing Tool

The transaction signing tool `cmd/txsigningtool` is a dummy testing tool for signing multisig attestations.

- Client Signup Tool

The client signup tool `cmd/clientsignuptool` can be used to sign up new clients to the Mainstay service.

- Token Generator Tool

The token generator tool `cmd/tokengeneratortool` can be used to generate unique authorization tokens for client signup.

- Multisig Tool

The multisig tool `cmd/multisigtool` can be used to generate multisig scripts and P2SH addresses for Mainstay configuration.

For more information go to [tool guidelines](/cmd/README.md).

For example use cases go to [docs](/docs).

