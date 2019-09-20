#!/bin/bash


if [ -f /run/secrets/ocean_pass ]; then
    export MAINSTAY_OCEAN_PASS="$(cat /run/secrets/ocean_pass)"
fi

if [ -f /run/secrets/commitment_priv ]; then
    export COMMITMENT_PRIV="$(cat /run/secrets/commitment_priv)"
fi


export HOST_1="*:5001"
export HOST_2="*:5002"

if [[ "$1" == "mainstay" ]]; then
    echo "Running attestation"
    mainstay
elif [[ "$1" == "signer1" ]]; then
    echo "Running signer 1"
    go run $GOPATH/src/mainstay/cmd/txsigningtool/txsigningtool.go -pk $PRIV_1 -pkTopup $PRIV_TOPUP_1 -host $HOST_1 -hostMain $HOST_MAIN
elif [[ "$1" == "signer2" ]]; then
    echo "Running signer 2"
    go run $GOPATH/src/mainstay/cmd/txsigningtool/txsigningtool.go -pk $PRIV_2 -pkTopup $PRIV_TOPUP_2 -host $HOST_2 -hostMain $HOST_MAIN
elif [[ "$1" == "ocean_commitment" ]]; then
    echo "Running commitment tool for Ocean"
    go run $GOPATH/src/mainstay/cmd/commitmenttool/commitmenttool.go -ocean -privkey $COMMITMENT_PRIV -position $COMMITMENT_POS -authtoken $COMMITMENT_AUTH -apiHost $MAINSTAY_HOST
else
  $@
fi
