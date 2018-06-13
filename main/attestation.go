// Attestation information on the transaction generated and the data attested

package main

import (
    "time"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
)

type Attestation struct {
    txid            chainhash.Hash
    clientHash      chainhash.Hash
    confirmed       bool
    latestTime      time.Time
}
