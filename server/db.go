package server

import (
	"mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Db Interface
type Db interface {
	saveAttestation(models.Attestation) error
	saveMerkleCommitments(commitments []models.CommitmentMerkleCommitment) error
	saveMerkleProofs(proofs []models.CommitmentMerkleProof) error

	getLatestAttestationMerkleRoot() (string, error)
	getClientCommitments() ([]models.ClientCommitment, error)
	getAttestationMerkleCommitments(chainhash.Hash) ([]models.CommitmentMerkleCommitment, error)
}
