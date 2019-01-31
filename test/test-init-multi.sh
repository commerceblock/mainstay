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

btcl importaddress "2N53Hkuyz8gSM2swdAvt7yqzxH8vVCKxgvK"
btcl importaddress "522103dfc3e2f3d0a3ebf9265d30a87764206d2ee0198820eee200eee4fb3f18eaac43210375f474311ba6248dc7ea1d4044114ee8e8c9cad3974ce2ae5a44dfaa285f3f372103cf016cd19049437c1cfa241bcf1baac58e22c71cae2dc06cb15259ee2f61bb2b53ae" "" true true
btcl sendtoaddress "2N53Hkuyz8gSM2swdAvt7yqzxH8vVCKxgvK" $(btcl getbalance) "" "" true

btcl generate 1
