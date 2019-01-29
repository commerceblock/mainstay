# Tools

## Transaction Signing Tool

The transaction signing tool can be used by each signer of the mainstay multisig to sign transactions.

`go run $GOPATH/src/mainstay/cmd/txsigningtool/txsigningtool.go -pk PRIVKEY -pkTopup TOPUP_PRIVKEY -host SIGNER_HOST`

where:

- `PRIVKEY`: private key of address initial funds were paid to
- `TOPUP_PRIVKEY`: private key of the topup address
- `SIGNER_HOST`: host address that the signer is publishing at and for the mainstay service to subscribe to

The tool subscribes to the mainstay service in order to receive confirmed attestation hashes and new bitcoin attestation transaction pre-images. These transactions are signed and broadcast back to the mainstay service.

To do the signing ECDSA libraries are used and and no Bitcoin node connection is required.

The live release of Mainstay will be instead using an HSM interface. Thus this tool is for testing purposes only.

## Client Signup Tool

The client signup tool can be used to sign up new clients to the mainstay service.

`go run $GOPATH/src/mainstay/cmd/clientsignuptool/clientsignuptool.go`

Connectivity to the mainstay db instance is required. Config can be set in `cmd/clientsignuptool/conf.json`.

The client will need to provide an ECDSA public key. The corresponding private key will be used by the client to sign the commitment send to the mainstay API. The signature is then verified by the API using the public key provided.

The tool assigns a new position to the client in the commitment merkle tree and also provides a unique auth_token for authorizing API POST requests submitted by the client. For random auth-token generation only, token generator tool `cmd/tokengeneratortool` can be used.

## Token Generator Tool

The token generator tool can be used to generate unique authorization tokens for client signup.

`go run $GOPATH/src/mainstay/cmd/tokengeneratortool/tokengeneratortool.go`

## Client Confirmation Tool

The confirmation tool can be used to confirm all the attestations of a client Ocean-type network to Bitcoin and wait for any new attestations that will be happening.

Running this tool will require a full Bitcoin testnet node and a full Ocean node. Connection details for these should be included in `cmd/confirmationtool/conf.json`.

The `API_HOST` field should be set to the mainstay URL. This can be updated in `cmd/confirmationtool/confirmationtool.go`.

To run this tool you need to first fetch the `TX_HASH` from the `attestationhash` field in the Ocean genesis block, as well as the publicly available `REDEEM_SCRIPT` of the attestation service multisig. The tool can also be started with any other `TX_HASH` attestation found in the mainstay website. A client should use his designated `CLIENT_POSITION` that was assigned during signup and run the tool using:

`go run cmd/confirmationtool/confirmationtool.go -tx TX_HASH -script REDEEM_SCRIPT -position CLIENT_POSITION -apiHost https://mainstay.xyz`

This will initially take some time to sync up all the attestations that have been committed so far and then will wait for any new attestations. Logging is displayed for each attestation and for full details the `-detailed` flag can be used.

## Commitment Tool

The commitment tool can be used to send hash commitments to the Mainstay API.

The tool functions in three different modes:

- Init mode to generate ECDSA keys
- One time commitment mode
- Recurrent commitment of Ocean blockhashes mode

Various command line arguments need to be provided:

- `-apiHost`: host address of Mainstay API (default: https://testnet.mainstay.xyz)
- `-init`: init mode to generate ECDSA pubkey/privkey (default: false)
- `-ocean`: ocean mode to use recurrent commitment mode (default: false)
- `-delay`: delay in minutes between sending commitments in ocean mode (default: 60)
- `-position`: client position on commitment merkle tree
- `-authtoken`: client authorization token generated on registration
- `-privkey`: Client private key, if signature has not been generated using a different source
- `-signature`: Client signature for commitment, for one time mode only
- `-commitment`: Commitment to be sent to API, for one time mode only

Ocean connectivity details need to be provided in the `cmd/commitmenttool/conf.json` file if Ocean mode is selected.

To run use the following along with the command line arguments, e.g.:

`go run $GOPATH/src/mainstay/cmd/commitmentool/commitmenttool.go -position 5 -authtoken abcbd-eddde-fllqqwoe -commitment 73902d2a365fff2724e26d975148124268ec6a84991016683817ea2c973b199b -signature MEUCIQCuuFmkoYnwRo5CsR7jE3m6MODsQMMLfCL4Vb5ILPYKCAIgCeh1AJD70L0s6TRr5dyoQAwdLbuBwUsgrYTux9XtXn0=`

## Multisig Tool

The multisig tool can be used to generate multisig scripts and P2SH addresses for Mainstay configuration.

Two modes:

- Regtest mode (multisig/P2SH generation for regtest and unit tests)
- Main mode

Command line arguments:

- `-chain`: set bitcoin chain configuration to regtest/testnet/mainnet (defaults to mainnet)
- `-nKeys`: num of keys (main mode)
- `-nSigs`: num of sigs (main mode)
- `-keys`: list of comma separated pub keys in hex format (main mode)
- `-keysX`: list of comma separated pub key X coordinates (main mode if -keys not set)
- `-keysY`: list of comma separated pub key Y coordinates (main mode if -keys not set)

The multisig generated can be used as the Mainstay `initScript` config option.

The P2SH address generated can be used to pay funds to initiate Mainstay.

Examples on how to run:


- `go run $GOPATH/src/mainstay/cmd/multisigtool/multisigtool.go -chain=mainnet -nKeys=2 -nSigs=1 -keysX=17073944010873801765385810419928396464299027769026919728232198509972577863206,80413053216156218546514694130398099327511867032326801302280634421130221500147 -keysY=475813022329769762590164284448176075334749443379722569322944728779216384721,11222700187475866687235948284541357909717856537392660494591205788179681685365`
- `go run $GOPATH/src/mainstay/cmd/multisigtool/multisigtool.go -chain=testnet -nKeys=2 -nSigs=1 -keys=03e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b33,03e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b33`
- `go run $GOPATH/src/mainstay/cmd/multisigtool/multisigtool.go -chain=regtest`
