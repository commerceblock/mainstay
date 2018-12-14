// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package models

import (
	"errors"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/assert"
)

// Test Commitment high level interface
func TestCommitment(t *testing.T) {
	hash0, _ := chainhash.NewHashFromStr("1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash1, _ := chainhash.NewHashFromStr("2a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash2, _ := chainhash.NewHashFromStr("3a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	root, _ := chainhash.NewHashFromStr("bb088c106b3379b64243c1a4915f72a847d45c7513b152cad583eb3c0a1063c2")
	commitments := []chainhash.Hash{*hash0, *hash1, *hash2}

	merkleTree := buildMerkleTree(commitments)
	proof0 := buildMerkleProof(0, merkleTree)
	proof1 := buildMerkleProof(1, merkleTree)
	proof2 := buildMerkleProof(2, merkleTree)
	proofs := []CommitmentMerkleProof{proof0, proof1, proof2}

	_, errCommitmentEmpty := NewCommitment([]chainhash.Hash{})
	assert.Equal(t, errors.New(ErrorCommitmentListEmpty), errCommitmentEmpty)

	commitment, errCommitment := NewCommitment(commitments)
	assert.Equal(t, nil, errCommitment)

	merkleCommitments := commitment.GetMerkleCommitments()
	for pos := range merkleCommitments {
		assert.Equal(t, commitments[pos], merkleCommitments[pos].Commitment)
		assert.Equal(t, pos, int(merkleCommitments[pos].ClientPosition))
		assert.Equal(t, commitment.GetCommitmentHash(), merkleCommitments[pos].MerkleRoot)
	}

	commitmentHash := commitment.GetCommitmentHash()
	assert.Equal(t, *root, commitmentHash)

	merkleProofs := commitment.GetMerkleProofs()
	assert.Equal(t, proofs, merkleProofs)
}

// Test Commitment BSON interface
func TestCommitmentBSON(t *testing.T) {
	hash0, _ := chainhash.NewHashFromStr("1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash1, _ := chainhash.NewHashFromStr("2a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash2, _ := chainhash.NewHashFromStr("3a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	root, _ := chainhash.NewHashFromStr("bb088c106b3379b64243c1a4915f72a847d45c7513b152cad583eb3c0a1063c2")
	commitments := []chainhash.Hash{*hash0, *hash1, *hash2}
	commitment, _ := NewCommitment(commitments)

	merkleCommitments := commitment.GetMerkleCommitments()

	// test marshal commitment model
	commitment0 := merkleCommitments[0]
	assert.Equal(t, *root, commitment0.MerkleRoot)

	bytes, errBytes := commitment0.MarshalBSON()
	assert.Equal(t, []uint8([]byte{0xbd, 0x0, 0x0, 0x0, 0x2, 0x6d, 0x65, 0x72, 0x6b, 0x6c, 0x65, 0x5f, 0x72, 0x6f, 0x6f, 0x74, 0x0, 0x41, 0x0, 0x0, 0x0, 0x62, 0x62, 0x30, 0x38, 0x38, 0x63, 0x31, 0x30, 0x36, 0x62, 0x33, 0x33, 0x37, 0x39, 0x62, 0x36, 0x34, 0x32, 0x34, 0x33, 0x63, 0x31, 0x61, 0x34, 0x39, 0x31, 0x35, 0x66, 0x37, 0x32, 0x61, 0x38, 0x34, 0x37, 0x64, 0x34, 0x35, 0x63, 0x37, 0x35, 0x31, 0x33, 0x62, 0x31, 0x35, 0x32, 0x63, 0x61, 0x64, 0x35, 0x38, 0x33, 0x65, 0x62, 0x33, 0x63, 0x30, 0x61, 0x31, 0x30, 0x36, 0x33, 0x63, 0x32, 0x0, 0x10, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x6d, 0x65, 0x6e, 0x74, 0x0, 0x41, 0x0, 0x0, 0x0, 0x31, 0x61, 0x33, 0x39, 0x65, 0x33, 0x34, 0x65, 0x38, 0x38, 0x31, 0x64, 0x39, 0x61, 0x31, 0x65, 0x36, 0x63, 0x64, 0x63, 0x33, 0x34, 0x31, 0x38, 0x62, 0x35, 0x34, 0x61, 0x61, 0x35, 0x37, 0x37, 0x34, 0x37, 0x31, 0x30, 0x36, 0x62, 0x63, 0x37, 0x35, 0x65, 0x39, 0x65, 0x38, 0x34, 0x34, 0x32, 0x36, 0x36, 0x36, 0x31, 0x66, 0x32, 0x37, 0x66, 0x39, 0x38, 0x61, 0x64, 0x61, 0x33, 0x62, 0x37, 0x0, 0x0}), bytes)
	assert.Equal(t, nil, errBytes)

	// test unmarshal commitment model and verify reverse works
	testCommitment0 := &CommitmentMerkleCommitment{}
	testCommitment0.UnmarshalBSON(bytes)
	assert.Equal(t, commitment0.MerkleRoot, testCommitment0.MerkleRoot)
	assert.Equal(t, commitment0.ClientPosition, testCommitment0.ClientPosition)
	assert.Equal(t, commitment0.Commitment, testCommitment0.Commitment)

	// test commitment model to document
	doc, docErr := GetDocumentFromModel(commitment0)
	assert.Equal(t, nil, docErr)
	assert.Equal(t, commitment0.MerkleRoot.String(), doc.Lookup(CommitmentMerkleRootName).StringValue())
	assert.Equal(t, commitment0.ClientPosition, doc.Lookup(CommitmentClientPositionName).Int32())
	assert.Equal(t, commitment0.Commitment.String(), doc.Lookup(CommitmentCommitmentName).StringValue())

	// test reverse document to commitment model
	testtestCommitment0 := &CommitmentMerkleCommitment{}
	docErr = GetModelFromDocument(doc, testtestCommitment0)
	assert.Equal(t, nil, docErr)
	assert.Equal(t, commitment0.MerkleRoot, testtestCommitment0.MerkleRoot)
	assert.Equal(t, commitment0.ClientPosition, testtestCommitment0.ClientPosition)
	assert.Equal(t, commitment0.Commitment, testtestCommitment0.Commitment)
}
