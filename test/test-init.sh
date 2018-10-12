#!/bin/bash
shopt -s expand_aliases

alias btcd="bitcoind -datadir=$HOME/btc-datadir"
alias btcl="bitcoin-cli -datadir=$HOME/btc-datadir"

btcl stop
sleep 1

rm -r ~/btc-datadir ;
mkdir ~/btc-datadir ;

printf '%s\n' '#!/bin/sh' 'rpcuser=user' \
    'rpcpassword=pass' \
    'rpcport=18443' \
    'keypool=0' \
    'deprecatedrpc=signrawtransaction' \
    'server=1' \
    'regtest=1' \
    'daemon=1' \
    'txindex=1' > ~/btc-datadir/bitcoin.conf

btcd
sleep 5

btcl generate 103
sleep 1

btcl importaddress "2MxBi6eodnuoVCw8McGrf1nuoVhastqoBXB"
btcl sendtoaddress "2MxBi6eodnuoVCw8McGrf1nuoVhastqoBXB" $(btcl getbalance) "" "" true
sleep 1

btcl generate 1
sleep 1
