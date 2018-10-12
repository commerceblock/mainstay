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

mkdir ~/ocean-datadir/terms-and-conditions
printf '%s\n' 'These are the terms and conditions' \
       'Approve to use the CBT network' > ~/ocean-datadir/terms-and-conditions/latest.txt

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
    "initialfreecoins=2100000000000000" \
    "signblockscript=52210394caaf4ccf40a216c2e1ad012d517b25b3d524c1249a55501b63a74b93adddd32103db95cad2e9506a926e22a7fa1294dc455fae403a39dc2a35ef9cb37813b9ba4a2102e5d227e18a196575a771be0993a639732bebdff6d6c3f0bd20568c21fb41730953ae" \
    "con_mandatorycoinbase=52210394caaf4ccf40a216c2e1ad012d517b25b3d524c1249a55501b63a74b93adddd32103db95cad2e9506a926e22a7fa1294dc455fae403a39dc2a35ef9cb37813b9ba4a2102e5d227e18a196575a771be0993a639732bebdff6d6c3f0bd20568c21fb41730953ae" \
    "issuecontrolscript=5121024da1a90bce74c44b37914a9c3e4cfc52953b90693c155cd4091d6c593b1624d151ae" \
    "initialfreecoinsdestination=5121024da1a90bce74c44b37914a9c3e4cfc52953b90693c155cd4091d6c593b1624d151ae"\
    "regtest=0" \
    "server=1" \
    "daemon=1" \
    "listen=1" \
    "txindex=1" \
    "disablewallet=1" > ~/ocean-datadir/elements.conf

btcd
sleep 2
oceand
sleep 2

btcl generate 103
sleep 1

#multisig=$(btcl createmultisig 1 '''["'''$pub1'''", "'''$pub2'''"]''')
#multisigaddr=$(echo $multisig | jq --raw-output '.address')

btcl importaddress "2MxBi6eodnuoVCw8McGrf1nuoVhastqoBXB"
btcl sendtoaddress "2MxBi6eodnuoVCw8McGrf1nuoVhastqoBXB" $(btcl getbalance) "" "" true
sleep 1

btcl generate 1
sleep 1

btcl listunspent
sleep 10
