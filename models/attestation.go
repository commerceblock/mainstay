package models

import (
	"errors"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/mongodb/mongo-go-driver/bson"
)

// Attestation structure
// Holds information on the attestation transaction generated
// and the information on the sidechain hash attested
// Attestation is unconfirmed until included in a mainchain block
type Attestation struct {
	Txid       chainhash.Hash
	Tx         wire.MsgTx
	Confirmed  bool
	commitment *Commitment
}

// Attestation constructor for defaulting some values
func NewAttestation(txid chainhash.Hash, commitment *Commitment) *Attestation {
	return &Attestation{txid, wire.MsgTx{}, false, commitment}
}

// Attestation constructor for defaulting all values
func NewAttestationDefault() *Attestation {
	return &Attestation{chainhash.Hash{}, wire.MsgTx{}, false, (*Commitment)(nil)}
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

// Implement bson.Marshaler MarshalBSON() method for use with db_mongo interface
func (a Attestation) MarshalBSON() ([]byte, error) {
	attestationBSON := AttestationBSON{a.Txid.String(), a.CommitmentHash().String(), a.Confirmed, time.Now()}
	return bson.Marshal(attestationBSON)
}

// Attestation field names
const (
	ATTESTATION_TXID_NAME        = "txid"
	ATTESTATION_MERKLE_ROOT_NAME = "merkle_root"
	ATTESTATION_CONFIRMED_NAME   = "confirmed"
	ATTESTATION_INSERTED_AT_NAME = "inserted_at"
)

// AttestationBSON structure for mongoDb
type AttestationBSON struct {
	Txid       string    `bson:"txid"`
	MerkleRoot string    `bson:"merkle_root"`
	Confirmed  bool      `bson:"confirmed"`
	InsertedAt time.Time `bson:"inserted_at"`
}
