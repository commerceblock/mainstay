package models

import (
	"errors"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/mongodb/mongo-go-driver/bson"
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

// Get merkle proofs for Commitment
func (c Commitment) GetMerkleProofs() []CommitmentMerkleProof {
	return c.tree.getMerkleProofs()
}

// Get merkle commitments for Commitment
func (c Commitment) GetMerkleCommitments() []CommitmentMerkleCommitment {
	var commitments []CommitmentMerkleCommitment
	for pos, commitment := range c.tree.getMerkleCommitments() {
		commitments = append(commitments, CommitmentMerkleCommitment{c.GetCommitmentHash(), int32(pos), commitment})
	}
	return commitments
}

// Get merkle root hash for Commitment
func (c Commitment) GetCommitmentHash() chainhash.Hash {
	return c.tree.getMerkleRoot()
}

// struct for db CommitmentMerkleCommitment
type CommitmentMerkleCommitment struct {
	MerkleRoot     chainhash.Hash
	ClientPosition int32
	Commitment     chainhash.Hash
}

// Implement bson.Marshaler MarshalBSON() method for use with db_mongo interface
func (c CommitmentMerkleCommitment) MarshalBSON() ([]byte, error) {
	commitmentBSON := CommitmentMerkleCommitmentBSON{c.MerkleRoot.String(), c.ClientPosition, c.Commitment.String()}
	return bson.Marshal(commitmentBSON)

}

//CommitmentMerkleCommitmentBSON structure for mongoDB
type CommitmentMerkleCommitmentBSON struct {
	MerkleRoot     string `bson:"merkle_root"`
	ClientPosition int32  `bson:"client_position"`
	Commitment     string `bson:"commitment"`
}
