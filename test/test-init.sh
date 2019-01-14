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

btcl importaddress "2N6kVS5GVY8jRtQV861Q6NchaaHZsyxSU7D"
btcl importaddress "512103e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b332103820968a1518a1d6edb9ba168402480cd3988b589f1aa2dd0d60c6cead25794f652ae" "" true true
btcl sendtoaddress "2N6kVS5GVY8jRtQV861Q6NchaaHZsyxSU7D" $(btcl getbalance) "" "" true

btcl generate 1
