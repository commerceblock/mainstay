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

btcl importaddress "2N8AAQy6SH5HGoAtzwr5xp4LTicqJ3fic8d"
btcl importaddress "512103e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b3321037361a2dba6a9e82faaf5465c36937adba283c878c506000b8479894c6f9cbae752ae" "" true true
btcl sendtoaddress "2N8AAQy6SH5HGoAtzwr5xp4LTicqJ3fic8d" $(btcl getbalance) "" "" true

btcl generate 1
