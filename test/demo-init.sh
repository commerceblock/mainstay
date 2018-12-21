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

btcl importaddress "2N6kVS5GVY8jRtQV861Q6NchaaHZsyxSU7D"
btcl importaddress "512103e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b332103820968a1518a1d6edb9ba168402480cd3988b589f1aa2dd0d60c6cead25794f652ae" "" true true
btcl sendtoaddress "2N6kVS5GVY8jRtQV861Q6NchaaHZsyxSU7D" $(btcl getbalance) "" "" true
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
