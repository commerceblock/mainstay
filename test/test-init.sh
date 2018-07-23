#!/bin/bash
shopt -s expand_aliases

alias btcd="bitcoind -datadir=$HOME/btc-datadir"
alias btcl="bitcoin-cli -datadir=$HOME/btc-datadir"

alias oceand="elementsd -datadir=$HOME/ocean-datadir"
alias oceanl="elements-cli -datadir=$HOME/ocean-datadir"

btcl stop
sleep 1

oceanl stop
sleep 1

rm -r ~/btc-datadir ; rm -r ~/ocean-datadir ;
mkdir ~/btc-datadir ; mkdir ~/ocean-datadir ;

printf '%s\n' '#!/bin/sh' 'rpcuser=user' \
    'rpcpassword=pass' \
    'rpcport=18443' \
    'keypool=0' \
    'deprecatedrpc=signrawtransaction' \
    'server=1' \
    'regtest=1' \
    'daemon=1' \
    'txindex=1' > ~/btc-datadir/bitcoin.conf

printf '%s\n' '#!/bin/sh' "rpcuser=bitcoinrpc" \
    "rpcpassword=acc1e7a299bc49449912e235b54dbce5" \
    "rpcport=18010" \
    "port=18011" \
    "initialfreecoins=50000000000000" \
    "regtest=1" \
    "server=1" \
    "daemon=1" \
    "listen=1" \
    "txindex=1" > ~/ocean-datadir/elements.conf

btcd
sleep 5
oceand
sleep 5

btcl generate 103
sleep 1

btcl sendtoaddress $(btcl getnewaddress) $(btcl getbalance) "" "" true
sleep 1

btcl generate 1
sleep 1

btcl listunspent
sleep 10
