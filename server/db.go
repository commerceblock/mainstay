package server

import (
	"mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Db Interface
type Db interface {
	saveAttestation(models.Attestation) error
	saveAttestationInfo(models.AttestationInfo) error
	saveMerkleCommitments(commitments []models.CommitmentMerkleCommitment) error
	saveMerkleProofs(proofs []models.CommitmentMerkleProof) error

	getAttestationCount(...bool) (int64, error)
	getAttestationMerkleRoot(chainhash.Hash) (string, error)

	getLatestAttestationMerkleRoot(bool) (string, error)
	getClientCommitments() ([]models.ClientCommitment, error)
	getAttestationMerkleCommitments(chainhash.Hash) ([]models.CommitmentMerkleCommitment, error)
}
