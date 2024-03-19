// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package db

import (
	"mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Db Interface
// Any struct implementing this Db interface needs to have a definition for all the methods below
// These methods are required for saving attestation information to the database
// as well as fetching information on previous commitments required for signing
type Db interface {
	// store methods
	SaveAttestation(models.Attestation) error
	SaveAttestationInfo(models.AttestationInfo) error
	SaveMerkleCommitments(commitments []models.CommitmentMerkleCommitment) error
	SaveMerkleProofs(proofs []models.CommitmentMerkleProof) error

	// util methods
	getAttestationCount(...bool) (int64, error)
	getAttestationMerkleRoot(chainhash.Hash) (string, error)

	// get methods required by server
	GetUnconfirmedAttestations() ([]models.Attestation, error)
	GetLatestAttestationMerkleRoot(bool) (string, error)
	GetClientCommitments() ([]models.ClientCommitment, error)
	GetAttestationMerkleCommitments(chainhash.Hash) ([]models.CommitmentMerkleCommitment, error)
}
