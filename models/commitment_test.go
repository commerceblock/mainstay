package models

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/assert"
)

func TestCommitment(t *testing.T) {
	hash1, _ := chainhash.NewHashFromStr("1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash2, _ := chainhash.NewHashFromStr("2a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash3, _ := chainhash.NewHashFromStr("3a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")

	_, errCommitmentEmpty := NewCommitment([]chainhash.Hash{})
	assert.Equal(t, "List of commitments is empty", errCommitmentEmpty.Error())

	commitments := []chainhash.Hash{*hash1, *hash2, *hash3}
	commitment, errCommitment := NewCommitment(commitments)
	assert.Equal(t, nil, errCommitment)

	merkleCommitments := commitment.GetMerkleCommitments()
	assert.Equal(t, merkleCommitments, commitments)

	commitmentHash := commitment.GetCommitmentHash()
	assert.Equal(t, "bb088c106b3379b64243c1a4915f72a847d45c7513b152cad583eb3c0a1063c2", commitmentHash.String())

	merkleProofs := commitment.GetMerkleProofs()
	assert.Equal(t, ([]interface{})(nil), merkleProofs)
}
