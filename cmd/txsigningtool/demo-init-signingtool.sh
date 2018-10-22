#!/bin/bash
shopt -s expand_aliases

alias btcd="bitcoind -datadir=$HOME/btc-datadir-signingtool"
alias btcl="bitcoin-cli -datadir=$HOME/btc-datadir-signingtool"

btcl stop
sleep 1

rm -r ~/btc-datadir-signingtool
mkdir ~/btc-datadir-signingtool

printf '%s\n' '#!/bin/sh' 'rpcuser=user' \
    'rpcpassword=pass' \
    'regtest.rpcport=18453' \
    'regtest.port=18454' \
    'regtest.addnode=localhost:18444' \
    'keypool=0' \
    'deprecatedrpc=signrawtransaction' \
    'server=1' \
    'regtest=1' \
    'daemon=1' \
    'txindex=1' > ~/btc-datadir-signingtool/bitcoin.conf

btcd
sleep 2

btcl importaddress "2MyC1i1FGy6MZWyMgmZXku4gdWZxWCRa6RL"
btcl importaddress "512103e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b332102f3a78a7bd6cf01c56312e7e828bef74134dfb109e59afd088526212d96518e7552ae" "" true true

btcl listunspent
sleep 10
