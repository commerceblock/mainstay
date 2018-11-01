package models

import (
	"errors"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Commitment structure
type Commitment struct {
	tree CommitmentMerkleTree
}

// Return new Commitment instance
func NewCommitment(commitments []chainhash.Hash) (*Commitment, error) {
	// check length
	if len(commitments) == 0 {
		return nil, errors.New("List of commitments is empty")
	}
	commitmentTree := NewCommitmentMerkleTree(commitments)
	return &Commitment{commitmentTree}, nil
}

// TODO:
func (c Commitment) GetMerkleProofs() []interface{} {
	return nil
}

// Get merkle commitments for Commitment
func (c Commitment) GetMerkleCommitments() []chainhash.Hash {
	return c.tree.getMerkleCommitments()
}

// Get merkle root hash for Commitment
func (c Commitment) GetCommitmentHash() chainhash.Hash {
	return c.tree.getMerkleRoot()
}
