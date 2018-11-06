package server

import (
	"mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Db Interface
type Db interface {
	saveAttestation(models.Attestation) error
	saveCommitment(models.Commitment) error
	getLatestAttestedCommitmentHash() (chainhash.Hash, error)
	getLatestCommitment() (models.Commitment, error)
	getAttestationCommitment(chainhash.Hash) (models.Commitment, error)
}
