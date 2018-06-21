# ocean-attestation
Federated blockchain attestation daemon

## Prerequisites
* Golang
* Bitcoin
* Ocean

## Instructions

- Install prerequisites following `scripts/build.sh`

- Run in test mode using
    `ocean-attestation -test`

- Run by providing initial transaction and private key
    `ocean-attestation -tx TX_HASH -pk PK`
