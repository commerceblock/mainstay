#!/bin/bash
shopt -s expand_aliases

alias btcd="bitcoind -datadir=/tmp/btc-datadir"
alias btcl="bitcoin-cli -datadir=/tmp/btc-datadir"

btcl stop
sleep 0.5

rm -r /tmp/btc-datadir ;
mkdir /tmp/btc-datadir ;

printf '%s\n' '#!/bin/sh' 'rpcuser=user' \
    'rpcpassword=pass' \
    'rpcport=18443' \
    'keypool=0' \
    'deprecatedrpc=signrawtransaction' \
    'server=1' \
    'regtest=1' \
    'daemon=1' \
    'txindex=1' > /tmp/btc-datadir/bitcoin.conf

btcd
sleep 0.5

btcl generate 103
sleep 0.5

btcl importaddress "bcrt1q7h6ue5w39ramd4ux6gtxh6swnrefpcfgt7vl64"
btcl sendtoaddress "bcrt1q7h6ue5w39ramd4ux6gtxh6swnrefpcfgt7vl64" $(btcl getbalance) "" "" true

btcl generate 1
