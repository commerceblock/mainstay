// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package models

import (
	"log"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/mongodb/mongo-go-driver/bson"
)

// Build merkle proof for a specific position in the merkle tree
func buildMerkleProof(position int, tree []*chainhash.Hash) CommitmentMerkleProof {

	// check proof commitment is valid
	numOfCommitments := len(tree)/2 + 1
	if position >= numOfCommitments || tree[position] == nil {
		return CommitmentMerkleProof{}
	}

	// add base commitment in proof
	var proof CommitmentMerkleProof
	proof.ClientPosition = int32(position)
	proof.Commitment = *tree[position]

	// find all intermediarey commitment ops
	// iterate through each tree height determining
	// the commitment that needs to be added to the proof
	// along with the operation type (append or not)
	var ops []CommitmentMerkleProofOp
	offset := 0
	depth := numOfCommitments
	depthPosition := position
	proofIndex := position
	for depth > 1 {
		var op CommitmentMerkleProofOp
		if proofIndex%2 == 0 { // left side
			op.Append = true
			if tree[proofIndex+1] == nil { // if nil append self
				op.Commitment = *tree[proofIndex]
			} else {
				op.Commitment = *tree[proofIndex+1]
			}
		} else { // right side
			op.Append = false
			op.Commitment = *tree[proofIndex-1]
		}
		ops = append(ops, op)

		// go to next tree height and depth size
		// halve initial position to get corresponding one in new depth
		offset += depth
		depth /= 2
		depthPosition /= 2
		proofIndex = offset + (depthPosition % depth)
	}
	proof.Ops = ops
	proof.MerkleRoot = *tree[len(tree)-1]
	return proof
}

// Prove a commitment using the merkle proof provided
func ProveMerkleProof(proof CommitmentMerkleProof) bool {
	hash := proof.Commitment
	log.Printf("client position: %d\n", proof.ClientPosition)
	log.Printf("client commitment: %s\n", hash.String())
	for i := range proof.Ops {
		if proof.Ops[i].Append {
			log.Printf("append: %s\n", proof.Ops[i].Commitment.String())
			hash = *hashLeaves(hash, proof.Ops[i].Commitment)
			log.Printf("result: %s\n", hash.String())
		} else {
			log.Printf("prepend: %s\n", proof.Ops[i].Commitment.String())
			hash = *hashLeaves(proof.Ops[i].Commitment, hash)
			log.Printf("result: %s\n", hash.String())
		}
	}
	log.Printf("merkle root: %s\n", proof.MerkleRoot.String())
	return hash == proof.MerkleRoot
}

// CommitmentMerkleProofOps structure
type CommitmentMerkleProofOp struct {
	Append     bool
	Commitment chainhash.Hash
}

// CommitmentMerkleProofOpsBSON structure for mongoDB
type CommitmentMerkleProofOpBSON struct {
	Append     bool   `bson:"append"`
	Commitment string `bson:"commitment"`
}

// Proof OP field names
const (
	ProofOpAppendName     = "append"
	ProofOpCommitmentName = "commitment"
)

// CommitmentMerkleProof structure
type CommitmentMerkleProof struct {
	MerkleRoot     chainhash.Hash
	ClientPosition int32
	Commitment     chainhash.Hash
	Ops            []CommitmentMerkleProofOp
}

// Implement bson.Marshaler MarshalBSON() method for use with db_mongo interface
func (c CommitmentMerkleProof) MarshalBSON() ([]byte, error) {
	proofBson := CommitmentMerkleProofBSON{MerkleRoot: c.MerkleRoot.String(), ClientPosition: c.ClientPosition, Commitment: c.Commitment.String()}

	var opsBson []CommitmentMerkleProofOpBSON
	for _, op := range c.Ops {
		opsBson = append(opsBson, CommitmentMerkleProofOpBSON{op.Append, op.Commitment.String()})
	}
	proofBson.Ops = opsBson
	return bson.Marshal(proofBson)
}

// Implement bson.Unmarshaler UnmarshalJSON() method for use with db_mongo interface
func (c *CommitmentMerkleProof) UnmarshalBSON(b []byte) error {
	var proofBSON CommitmentMerkleProofBSON
	if err := bson.Unmarshal(b, &proofBSON); err != nil {
		return err
	}
	// TODO : not implemented as not required anywhere at the moment
	return nil
}

// Proof field names
const (
	ProofMerkleRootName     = "merkle_root"
	ProofClientPositionName = "client_position"
	ProofCommitmentName     = "commitment"
	ProofOpsName            = "ops"
)

// CommitmentMerkleProofBSON structure for mongoDB
type CommitmentMerkleProofBSON struct {
	MerkleRoot     string                        `bson:"merkle_root"`
	ClientPosition int32                         `bson:"client_position"`
	Commitment     string                        `bson:"commitment"`
	Ops            []CommitmentMerkleProofOpBSON `bson:"ops"`
}
