package models

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/assert"
)

// Test Commitment high level interface
func TestCommitment(t *testing.T) {
	hash0, _ := chainhash.NewHashFromStr("1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash1, _ := chainhash.NewHashFromStr("2a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash2, _ := chainhash.NewHashFromStr("3a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	commitments := []chainhash.Hash{*hash0, *hash1, *hash2}

	merkleTree := buildMerkleTree(commitments)
	proof0 := buildMerkleProof(0, merkleTree)
	proof1 := buildMerkleProof(1, merkleTree)
	proof2 := buildMerkleProof(2, merkleTree)
	proofs := []CommitmentMerkleProof{proof0, proof1, proof2}

	_, errCommitmentEmpty := NewCommitment([]chainhash.Hash{})
	assert.Equal(t, "List of commitments is empty", errCommitmentEmpty.Error())

	commitment, errCommitment := NewCommitment(commitments)
	assert.Equal(t, nil, errCommitment)

	merkleCommitments := commitment.GetMerkleCommitments()
	assert.Equal(t, merkleCommitments, commitments)

	commitmentHash := commitment.GetCommitmentHash()
	assert.Equal(t, "bb088c106b3379b64243c1a4915f72a847d45c7513b152cad583eb3c0a1063c2", commitmentHash.String())

	merkleProofs := commitment.GetMerkleProofs()
	assert.Equal(t, proofs, merkleProofs)
}
