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
    "addnode=localhost:18801" \
    "addnode=localhost:18803" \
    "initialfreecoins=50000000000000" \
    "signblockscript=522103d517f6e9affa60000a08d478970e6bbfa45d63b1967ed1e066dd46b802edb2a62102afc18e8a7ff988ca1ae7b659cb09a79852d301c2283e18cba1faf7a0b020b1a22102edd8080e31f05c68cf68a97782ac97744e86ba19dfd3ba68e597f10868ee5bc453ae" \
    "regtest=0" \
    "server=1" \
    "daemon=1" \
    "listen=1" \
    "txindex=1" \
    "disablewallet=1" > ~/ocean-datadir/elements.conf

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
