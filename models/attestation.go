package models

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// Attestation structure
// Holds information on the attestation transaction generated
// and the information on the sidechain hash attested
// Attestation is unconfirmed until included in a mainchain block
type Attestation struct {
	Txid         chainhash.Hash
	AttestedHash chainhash.Hash
	Tx           wire.MsgTx
}

// Attestation constructor for defaulting some values
func NewAttestation(txid chainhash.Hash, hash chainhash.Hash) *Attestation {
	return &Attestation{txid, hash, wire.MsgTx{}}
}

// Attestation constructor for defaulting all values
func NewAttestationDefault() *Attestation {
	return &Attestation{chainhash.Hash{}, chainhash.Hash{}, wire.MsgTx{}}
}
