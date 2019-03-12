// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package models

import (
	"errors"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"go.mongodb.org/mongo-driver/bson"
)

// error consts
const (
	ErrorCommitmentNotDefined = "Commitment not defined"
)

// Attestation structure
// Holds information on the attestation transaction generated
// and the information on the sidechain hash attested
// Attestation is unconfirmed until included in a mainchain block
type Attestation struct {
	Txid       chainhash.Hash
	Tx         wire.MsgTx
	Confirmed  bool
	Info       AttestationInfo
	commitment *Commitment
}

// Attestation constructor for defaulting some values
func NewAttestation(txid chainhash.Hash, commitment *Commitment) *Attestation {
	return &Attestation{txid, wire.MsgTx{}, false, AttestationInfo{}, commitment}
}

// Attestation constructor for defaulting all values
func NewAttestationDefault() *Attestation {
	return &Attestation{chainhash.Hash{}, wire.MsgTx{}, false, AttestationInfo{}, (*Commitment)(nil)}
}

// Update info with details from wallet transaction
func (a *Attestation) UpdateInfo(tx *btcjson.GetTransactionResult) {
	amount := int64(0)
	if len(a.Tx.TxOut) > 0 {
		amount = a.Tx.TxOut[0].Value
	}
	a.Info = AttestationInfo{
		Txid:      a.Txid.String(),
		Blockhash: tx.BlockHash,
		Amount:    amount,
		Time:      tx.Time,
	}
}

// Set commitment
func (a *Attestation) SetCommitment(commitment *Commitment) {
	a.commitment = commitment
}

// Get commitment
func (a Attestation) Commitment() (*Commitment, error) {
	if a.commitment == (*Commitment)(nil) {
		return (*Commitment)(nil), errors.New(ErrorCommitmentNotDefined)
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
	attestationTime := time.Now()
	if a.Info.Time != 0 { // check if tx time set
		attestationTime = time.Unix(a.Info.Time, 0)
	}
	attestationBSON := AttestationBSON{a.Txid.String(), a.CommitmentHash().String(), a.Confirmed, attestationTime}
	return bson.Marshal(attestationBSON)
}

// Implement bson.Unmarshaler UnmarshalJSON() method for use with db_mongo interface
func (a *Attestation) UnmarshalBSON(b []byte) error {
	var attestationBSON AttestationBSON
	if err := bson.Unmarshal(b, &attestationBSON); err != nil {
		return err
	}
	txidHash, errHash := chainhash.NewHashFromStr(attestationBSON.Txid)
	if errHash != nil {
		return errHash
	}
	a.Txid = *txidHash
	a.Confirmed = attestationBSON.Confirmed
	// THIS IS INCOMPLETE
	// in order to get a full Attestation model
	// we still need to Umarshal the commitment
	// model and set through SetCommitment()
	return nil
}

// Attestation field names
const (
	AttestationTxidName       = "txid"
	AttestationMerkleRootName = "merkle_root"
	AttestationConfirmedName  = "confirmed"
	AttestationInsertedAtName = "inserted_at"
)

// AttestationBSON structure for mongoDb
type AttestationBSON struct {
	Txid       string    `bson:"txid"`
	MerkleRoot string    `bson:"merkle_root"`
	Confirmed  bool      `bson:"confirmed"`
	InsertedAt time.Time `bson:"inserted_at"`
}
