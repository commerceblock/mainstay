# Mock data

Proof object (from `https://mainstay.xyz/api/v1/commitment/latestproof?position=0`)

```
{"response":
  {  "txid":"b891111d35ffc72709140b7bd2a82fde20deca53831f42a96704dede42c793d2",
    "commitment":"ef185d0085995021811d95cca6670756fb05959e2af6f02793d3f67327f29107",
    "merkle_root":"8d0ad2782d8f6e3f63c6f9611841c239630b55061d558abcc6bac53349edac70",
    "ops":[{"append":true,"commitment":"0000000000000000000000000000000000000000000000000000000000000000"}]
  },
  "timestamp":1698071913011,"allowance":{"cost":7086667}}
```

`BASE_PUBKEY: 031dd94c5262454986a2f0a6c557d2cbe41ec5a8131c588b9367c9310125a8a7dc`
`CHAINCODE: 0a090f710e47968aee906804f211cf10cde9a11e14908ca0f78cc55dd190ceaa`

For the specified attestation transaction (TxID: `b891111d35ffc72709140b7bd2a82fde20deca53831f42a96704dede42c793d2`)

The `bitcoind` RPC command `getrawtransaction b891111d35ffc72709140b7bd2a82fde20deca53831f42a96704dede42c793d2 true` (verbose output) returns:

```
{
    "result": {
        "txid": "b891111d35ffc72709140b7bd2a82fde20deca53831f42a96704dede42c793d2",
        "hash": "b891111d35ffc72709140b7bd2a82fde20deca53831f42a96704dede42c793d2",
        "version": 2,
        "size": 194,
        "vsize": 194,
        "weight": 776,
        "locktime": 0,
        "vin": [
            {
                "txid": "047352f01e5e3f8adc04a797311dde3917f274e55ceafb78edc39ff5d87d16c5",
                "vout": 0,
                "scriptSig": {
                    "asm": "0 30440220049d3138f841b63e96725cb9e86a53a92cd1d9e1b0740f5d4cd2ae0bcab684bf0220208d555c7e24e4c01cf67dfa9161091533e9efd6d1602bb53a49f7195c16b037[ALL] 5121036bd7943325ed9c9e1a44d98a8b5759c4bf4807df4312810ed5fc09dfb967811951ae",
                    "hex": "004730440220049d3138f841b63e96725cb9e86a53a92cd1d9e1b0740f5d4cd2ae0bcab684bf0220208d555c7e24e4c01cf67dfa9161091533e9efd6d1602bb53a49f7195c16b03701255121036bd7943325ed9c9e1a44d98a8b5759c4bf4807df4312810ed5fc09dfb967811951ae"
                },
                "sequence": 4294967293
            }
        ],
        "vout": [
            {
                "value": 0.01040868,
                "n": 0,
                "scriptPubKey": {
                    "asm": "OP_HASH160 29d13058087ddf2d48de404376fdcb5c4abff4bc OP_EQUAL",
                    "desc": "addr(35W8E71bdDhQw4ZC7uUZvXG3qhyWVYxfMB)#4rtfrxzg",
                    "hex": "a91429d13058087ddf2d48de404376fdcb5c4abff4bc87",
                    "address": "35W8E71bdDhQw4ZC7uUZvXG3qhyWVYxfMB",
                    "type": "scripthash"
                }
            }
        ],
        "hex": "0200000001c5167dd8f59fc3ed78fbea5ce574f21739de1d3197a704dc8a3f5e1ef0527304000000006f004730440220049d3138f841b63e96725cb9e86a53a92cd1d9e1b0740f5d4cd2ae0bcab684bf0220208d555c7e24e4c01cf67dfa9161091533e9efd6d1602bb53a49f7195c16b03701255121036bd7943325ed9c9e1a44d98a8b5759c4bf4807df4312810ed5fc09dfb967811951aefdffffff01e4e10f000000000017a91429d13058087ddf2d48de404376fdcb5c4abff4bc8700000000",
        "blockhash": "000000000000000000036cb20420528cf0f00abb3a5716d80b5c87146b764d47",
        "confirmations": 12894,
        "time": 1690540748,
        "blocktime": 1690540748
    },
    "error": null,
    "id": null
}
```

The `bitcoind` RPC command `getrawtransaction b891111d35ffc72709140b7bd2a82fde20deca53831f42a96704dede42c793d2 false` (non-verbose output) returns:

```
{
    "result": "0200000001c5167dd8f59fc3ed78fbea5ce574f21739de1d3197a704dc8a3f5e1ef0527304000000006f004730440220049d3138f841b63e96725cb9e86a53a92cd1d9e1b0740f5d4cd2ae0bcab684bf0220208d555c7e24e4c01cf67dfa9161091533e9efd6d1602bb53a49f7195c16b03701255121036bd7943325ed9c9e1a44d98a8b5759c4bf4807df4312810ed5fc09dfb967811951aefdffffff01e4e10f000000000017a91429d13058087ddf2d48de404376fdcb5c4abff4bc8700000000",
    "error": null,
    "id": null
}
```

The `merkle_root` of the proof object gives the follwoing BIP175 derivation path:

```
28844
60745
13253
47814
48266
21789
1621
2915
14786
16664
25081
50787
16238
36653
30930
2701
```


