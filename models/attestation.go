package models

import (
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// Attestation state type
type AttestationState int

// Attestation states
const (
	ASTATE_NEW_ATTESTATION    AttestationState = 0
	ASTATE_UNCONFIRMED        AttestationState = 1
	ASTATE_CONFIRMED          AttestationState = 2
	ASTATE_COLLECTING_PUBKEYS AttestationState = 3
	ASTATE_COLLECTING_SIGS    AttestationState = 4
	ASTATE_FAILURE            AttestationState = 10
	ASTATE_INIT               AttestationState = 100
)

// Attestation structure
// Holds information on the attestation transaction generated
// and the information on the sidechain hash attested
// Attestation is unconfirmed until included in a mainchain block
type Attestation struct {
	Txid         chainhash.Hash
	AttestedHash chainhash.Hash
	State        AttestationState
	LatestTime   time.Time
	Tx           wire.MsgTx
	Txunspent    btcjson.ListUnspentResult
	RedeemScript string
}

// Attestation constructor for defaulting some values
func NewAttestation(txid chainhash.Hash, hash chainhash.Hash, state AttestationState) *Attestation {
	return &Attestation{txid, hash, state, time.Now(), wire.MsgTx{}, btcjson.ListUnspentResult{}, ""}
}

// Attestation constructor for defaulting all values
func NewAttestationDefault() *Attestation {
	return &Attestation{chainhash.Hash{}, chainhash.Hash{}, ASTATE_INIT, time.Now(), wire.MsgTx{}, btcjson.ListUnspentResult{}, ""}
}
