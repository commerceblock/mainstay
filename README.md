# ocean-attestation
Federated blockchain attestation daemon

## Prerequisites
* Golang
* Bitcoin
* Ocean

## Instructions

- Install prerequisites following `scripts/build.sh`

- Regtest Demo
    - Start the ocean-demo `python ocean-demo/PROTOCOLS/demo.py`
    - Run `ocean-attestation -regtest`

- Main Mode
    - Create a BTC wallet, fund it and send all the funds to a single address
    - Take the `TX_HASH` and the `PK` of the address generated, for the initial transaction
    - Create a `conf.json` file based on the template provided and move it to `conf/conf.json`
    - Run `ocean-attestation -tx TX_HASH -pk PK`

- Unit Testing
    - `/$GOPATH/src/ocean-attestation/run-tests.sh`
