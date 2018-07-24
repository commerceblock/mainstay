package attestation

import (
    "time"

    "github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Attestation structure
// Holds information on the attestation transaction generated
// and the information on the sidechain hash attested
// Attestation is unconfirmed until included in a mainchain block
type Attestation struct {
    txid            chainhash.Hash
    attestedHash    chainhash.Hash
    confirmed       bool
    latestTime      time.Time
}
