#!/bin/bash
killall -9 bitcoind
sleep 2
killall -9 elementsd
sleep 2

rm -r ~/btc-datadir ; rm -r ~/ocean-datadir ;
mkdir ~/btc-datadir ; mkdir ~/ocean-datadir ;

printf '%s\n' '#!/bin/sh' 'rpcuser=user' \
    'rpcpassword=pass' \
    'rpcport=18000' \
    'keypool=0' \
    'deprecatedrpc=signrawtransaction' \
    'server=1' \
    'regtest=1' \
    'daemon=1' \
    'txindex=1' > ~/btc-datadir/bitcoin.conf

printf '%s\n' '#!/bin/sh' "rpcuser=user" \
    "rpcpassword=pass" \
    "rpcport=18001" \
    "port=10001" \
    "initialfreecoins=123456789" \
    "regtest=1" \
    "server=1" \
    "daemon=1" \
    "listen=1" \
    "txindex=1" > ~/ocean-datadir/elements.conf

shopt -s expand_aliases

alias btcd="bitcoind -datadir=$HOME/btc-datadir"
alias btcl="bitcoin-cli -datadir=$HOME/btc-datadir"

alias oceand="elementsd -datadir=$HOME/ocean-datadir"
alias oceanl="elements-cli -datadir=$HOME/ocean-datadir"

btcd
sleep 5
oceand
sleep 5

btcl generate 103
sleep 3

btcl sendtoaddress $(btcl getnewaddress) $(btcl getbalance) "" "" true
sleep 3

btcl generate 1
sleep 3
