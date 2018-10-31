package models

import (
	"errors"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// Attestation structure
// Holds information on the attestation transaction generated
// and the information on the sidechain hash attested
// Attestation is unconfirmed until included in a mainchain block
type Attestation struct {
	Txid       chainhash.Hash
	Tx         wire.MsgTx
	commitment *Commitment
}

// Attestation constructor for defaulting some values
func NewAttestation(txid chainhash.Hash, commitment *Commitment) *Attestation {
	return &Attestation{txid, wire.MsgTx{}, commitment}
}

// Attestation constructor for defaulting all values
func NewAttestationDefault() *Attestation {
	return &Attestation{chainhash.Hash{}, wire.MsgTx{}, (*Commitment)(nil)}
}

// Set commitment
func (a *Attestation) SetCommitment(commitment *Commitment) {
	a.commitment = commitment
}

// Get commitment
func (a Attestation) Commitment() (*Commitment, error) {
	if a.commitment == (*Commitment)(nil) {
		return (*Commitment)(nil), errors.New("Attestation Commitment not defined")
	}
	return a.commitment, nil
}

// Get commitment hash
func (a Attestation) CommitmentHash() chainhash.Hash {
	if a.commitment == (*Commitment)(nil) {
		return chainhash.Hash{}
	}
	return a.commitment.GetCommitmentHash()
}
