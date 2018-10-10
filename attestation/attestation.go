package attestation

import (
    "time"

    "github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Attestation state type
type AttestationState int

// Attestation states
const (
    ASTATE_NEW_ATTESTATION         AttestationState = 0
    ASTATE_UNCONFIRMED             AttestationState = 1
    ASTATE_CONFIRMED               AttestationState = 2
    ASTATE_AWAITING_PUBKEYS        AttestationState = 3
    ASTATE_AWAITING_SIGS           AttestationState = 4
    ASTATE_FAILURE                 AttestationState = 10
)

// Attestation structure
// Holds information on the attestation transaction generated
// and the information on the sidechain hash attested
// Attestation is unconfirmed until included in a mainchain block
type Attestation struct {
    txid            chainhash.Hash
    attestedHash    chainhash.Hash
    state           AttestationState
    latestTime      time.Time
}
