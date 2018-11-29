// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package models

import (
	"errors"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/mongodb/mongo-go-driver/bson"
)

// error consts
const (
	ERROR_COMMITMENT_LIST_EMPTY = "List of commitments is empty"
)

// Commitment structure
type Commitment struct {
	tree CommitmentMerkleTree
}

// Return new Commitment instance
func NewCommitment(commitments []chainhash.Hash) (*Commitment, error) {
	// check length
	if len(commitments) == 0 {
		return nil, errors.New(ERROR_COMMITMENT_LIST_EMPTY)
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

// Implement bson.Unmarshaler UnmarshalJSON() method for use with db_mongo interface
func (c *CommitmentMerkleCommitment) UnmarshalBSON(b []byte) error {
	var commitmentBSON CommitmentMerkleCommitmentBSON
	if err := bson.Unmarshal(b, &commitmentBSON); err != nil {
		return err
	}
	rootHash, errHash := chainhash.NewHashFromStr(commitmentBSON.MerkleRoot)
	if errHash != nil {
		return errHash
	}
	commitHash, errHash := chainhash.NewHashFromStr(commitmentBSON.Commitment)
	if errHash != nil {
		return errHash
	}

	c.MerkleRoot = *rootHash
	c.ClientPosition = commitmentBSON.ClientPosition
	c.Commitment = *commitHash
	return nil
}

// Commitment field names
const (
	COMMITMENT_MERKLE_ROOT_NAME     = "merkle_root"
	COMMITMENT_CLIENT_POSITION_NAME = "client_position"
	COMMITMENT_COMMITMENT_NAME      = "commitment"
)

//CommitmentMerkleCommitmentBSON structure for mongoDB
type CommitmentMerkleCommitmentBSON struct {
	MerkleRoot     string `bson:"merkle_root"`
	ClientPosition int32  `bson:"client_position"`
	Commitment     string `bson:"commitment"`
}
