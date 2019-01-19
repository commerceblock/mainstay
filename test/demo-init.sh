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

# btcl importaddress "2N8AAQy6SH5HGoAtzwr5xp4LTicqJ3fic8d"
# btcl importaddress "512103e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b3321037361a2dba6a9e82faaf5465c36937adba283c878c506000b8479894c6f9cbae752ae" "" true true
btcl sendtoaddress "2N8AAQy6SH5HGoAtzwr5xp4LTicqJ3fic8d" $(btcl getbalance) "" "" true
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
