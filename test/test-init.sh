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

btcl importaddress "2N74sgEvpJRwBZqjYUEXwPfvuoLZnRaF1xJ"
btcl importaddress "512103e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b33210325bf82856a8fdcc7a2c08a933343d2c6332c4c252974d6b09b6232ea4080462652ae" "" true true
btcl sendtoaddress "2N74sgEvpJRwBZqjYUEXwPfvuoLZnRaF1xJ" $(btcl getbalance) "" "" true

btcl generate 1
