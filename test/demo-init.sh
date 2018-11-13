#!/bin/bash
shopt -s expand_aliases

alias btcd="bitcoind -datadir=$HOME/btc-datadir"
alias btcl="bitcoin-cli -datadir=$HOME/btc-datadir"

btcl stop
sleep 1

rm -r ~/btc-datadir
mkdir ~/btc-datadir

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
sleep 2

btcl generate 103
sleep 1

#multisig=$(btcl createmultisig 1 '''["'''$pub1'''", "'''$pub2'''"]''')
#multisigaddr=$(echo $multisig | jq --raw-output '.address')

btcl importaddress "2MxBi6eodnuoVCw8McGrf1nuoVhastqoBXB"
btcl importaddress "512103e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b332102f3a78a7bd6cf01c56312e7e828bef74134dfb109e59afd088526212d96518e7552ae" "" true true
btcl sendtoaddress "2MxBi6eodnuoVCw8McGrf1nuoVhastqoBXB" $(btcl getbalance) "" "" true
sleep 1

btcl generate 1
sleep 1

btcl stop
sleep 1

printf '%s\n' '#!/bin/sh' 'rpcuser=user' \
    'rpcpassword=pass' \
    'rpcport=18443' \
    'keypool=0' \
    'deprecatedrpc=signrawtransaction' \
    'blocksonly=1' \
    'server=1' \
    'regtest=1' \
    'daemon=1' \
    'txindex=1' > ~/btc-datadir/bitcoin.conf

btcd
sleep 2

btcl listunspent
sleep 10
