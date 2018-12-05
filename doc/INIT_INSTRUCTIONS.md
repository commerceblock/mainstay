# General instructions for initialising Mainstay

## Initial attestation

### Generate 2 addresses
$ bitcoin-cli -datadir=testnetbtc-datadir/ getnewaddress
2NFBB5okotyGFLmceXK7q18ufuv11NmefUJ

$ bitcoin-cli -datadir=testnetbtc-datadir/ getnewaddress
2NE8WKRRuj53udVsuyj5GbVfyUNN6ZSE4ia

### Generate multisig 1 of 2 address
$ bitcoin-cli -datadir=testnetbtc-datadir/ addmultisigaddress 1 "[\"2NFBB5okotyGFLmceXK7q18ufuv11NmefUJ\",\"2NE8WKRRuj53udVsuyj5GbVfyUNN6ZSE4ia\"]" "" legacy

{
  "address": "2N3CJAMXMmBaT1bAHc1LHwXWVRVKHHxdYyj",
  "redeemScript": "51210381324c14a482646e9ad7cf82372021e5ecb9a7e1b67ee168dddf1e97dafe40af210376c091faaeb6bb3b74e0568db5dd499746d99437758a5cb1e60ab38f02e279c352ae"
}

### Dump priv keys
$ bitcoin-cli -datadir=testnetbtc-datadir/ dumpprivkey 2NFBB5okotyGFLmceXK7q18ufuv11NmefUJ
cVYM5QbqdvXn4NEUCM5jgC4tTw7BieTNmty4fJgqGtPLWX14KXuA

$ bitcoin-cli -datadir=testnetbtc-datadir/ dumpprivkey 2NE8WKRRuj53udVsuyj5GbVfyUNN6ZSE4ia
cVZ6pShjpV37tzXr7GJ5s2oqdgB3DMrV9xNgTrwhhMyQ2n79YJFL

### Import generated multisig address
$ bitcoin-cli -datadir=testnetbtc-datadir/ importaddress 2N3CJAMXMmBaT1bAHc1LHwXWVRVKHHxdYyj "" false

### Send funds to generated multisig address
$ bitcoin-cli -datadir=testnetbtc-datadir/ getbalance
0.07926900

bitcoin-cli -datadir=testnetbtc-datadir/ sendtoaddress 2N3CJAMXMmBaT1bAHc1LHwXWVRVKHHxdYyj 0.07926900 "" "" true

87e56bda501ba6a022f12e178e9f1ac03fb2c07f04e1dfa62ac9e1d83cd840e1

bitcoin-cli -datadir=testnetbtc-datadir/ sendrawtransaction 02000000000101bbd869ef95b3280aad7e6d8c77582d1d7a3d0dc60fc3c3c0228df6931c31561b000000002322002055dec27024dac08f4a4c0738c4834315e6c99004b02348b2f9df960926f44c4bfeffffff0158e878000000000017a9146d237e71ec246acfcc80b249e0e835b9bfe2175687030047304402200c972818b73932c6f48d86f9b9a2c1a67d42b6b798280e51e101145d247630ac022037984534b3e06d38eaecdc849c06d25b35d98330d96b95af8106897188b540050147512102e1ee4e5801efc577a8a9fac006a5908af7dfd37b03bd6bba830d1d3cb7a1ba7821027e73fcf0a3d86eaad56cae92524d4eeac42ec0e83af75c10b2d171f43f42325c52aed4051600
87e56bda501ba6a022f12e178e9f1ac03fb2c07f04e1dfa62ac9e1d83cd840e1

## Topup information

### Generate 2 addresses
$ bitcoin-cli -datadir=/Users/nikolaos/testnetbtc-datadir2/ getnewaddress
2MtEZ7J8ZXoieL7iHyUQw91TZpLEcVQTAyK

$ bitcoin-cli -datadir=/Users/nikolaos/testnetbtc-datadir2/ getnewaddress
2MwvCUjtecBAFcc7SWhEu8NyT1bLsCRtN6J

### Generate multisig 1 of 2 address
$ bitcoin-cli -datadir=/Users/nikolaos/testnetbtc-datadir2/ addmultisigaddress 1 "[\"2MtEZ7J8ZXoieL7iHyUQw91TZpLEcVQTAyK\",\"2MwvCUjtecBAFcc7SWhEu8NyT1bLsCRtN6J\"]" "" legacy

{
  "address": "2NBYFyyMpPeLCb67bykLBHMByu1dSRGsim1",
  "redeemScript": "512102a2411030da6082ac32d0166fc19f03e264c6c2a138f83a29120d0b59969670792103d11753d31309988c323142a0171e5b2319a8651479835afa4ab8ecb6442141b952ae"
}

### Dump priv keys
$ bitcoin-cli -datadir=/Users/nikolaos/testnetbtc-datadir2/ dumpprivkey 2MtEZ7J8ZXoieL7iHyUQw91TZpLEcVQTAyK
cPLfW9BRRJjZNwNHwrz6B5XEmsTHRFsHYYFRTAQChULT5nUn8FkW

$ bitcoin-cli -datadir=/Users/nikolaos/testnetbtc-datadir2/ dumpprivkey 2MwvCUjtecBAFcc7SWhEu8NyT1bLsCRtN6J
cTgsB8DjF2vjhFrtCPopvknmNWP6CTQeb1Sd9zvXxRF5qHp9V4ct

## Running the service
`go build && go install && mainstay`

## Running the signing tools

- signer 1

`go run $GOPATH/src/mainstay/cmd/txsigningtool/txsigningtool.go -pk cVZ6pShjpV37tzXr7GJ5s2oqdgB3DMrV9xNgTrwhhMyQ2n79YJFL -pkTopup cPLfW9BRRJjZNwNHwrz6B5XEmsTHRFsHYYFRTAQChULT5nUn8FkW -host *:5001`

- signer 2

`go run $GOPATH/src/mainstay/cmd/txsigningtool/txsigningtool.go -pk cVYM5QbqdvXn4NEUCM5jgC4tTw7BieTNmty4fJgqGtPLWX14KXuA -pkTopup cTgsB8DjF2vjhFrtCPopvknmNWP6CTQeb1Sd9zvXxRF5qHp9V4ct -host *:5002`
