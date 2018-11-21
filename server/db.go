package server

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
	saveAttestation(models.Attestation) error
	saveAttestationInfo(models.AttestationInfo) error
	saveMerkleCommitments(commitments []models.CommitmentMerkleCommitment) error
	saveMerkleProofs(proofs []models.CommitmentMerkleProof) error

	// util methods
	getAttestationCount(...bool) (int64, error)
	getAttestationMerkleRoot(chainhash.Hash) (string, error)

	// get methods required by server
	getLatestAttestationMerkleRoot(bool) (string, error)
	getClientCommitments() ([]models.ClientCommitment, error)
	getAttestationMerkleCommitments(chainhash.Hash) ([]models.CommitmentMerkleCommitment, error)
}
