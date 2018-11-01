package models

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
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
	proof.Root = *tree[len(tree)-1]
	return proof
}

// Prove a commitment using the merkle proof provided
func proveMerkleProof(proof CommitmentMerkleProof) bool {
	hash := proof.Commitment
	for i := range proof.Ops {
		if proof.Ops[i].Append {
			hash = *hashLeaves(hash, proof.Ops[i].Commitment)
		} else {
			hash = *hashLeaves(proof.Ops[i].Commitment, hash)
		}
	}
	return hash == proof.Root
}

// CommitmentMerkleProofOps structure
type CommitmentMerkleProofOp struct {
	Append     bool
	Commitment chainhash.Hash
}

// CommitmentMerkleProof structure
type CommitmentMerkleProof struct {
	Commitment chainhash.Hash
	Ops        []CommitmentMerkleProofOp
	Root       chainhash.Hash
}
