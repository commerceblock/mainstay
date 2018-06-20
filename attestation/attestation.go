// Attestation information on the transaction generated and the data attested

package attestation

import (
    "time"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
)

type Attestation struct {
    txid            chainhash.Hash
    attestedHash    chainhash.Hash
    confirmed       bool
    latestTime      time.Time
}
