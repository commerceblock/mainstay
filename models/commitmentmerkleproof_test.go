// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package models

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/assert"
)

// Test build merkle proof and verify for 5 commitment tree
func TestMerkleProof_5Commitments(t *testing.T) {
	hash0, _ := chainhash.NewHashFromStr("1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash1, _ := chainhash.NewHashFromStr("2a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash2, _ := chainhash.NewHashFromStr("3a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash3, _ := chainhash.NewHashFromStr("4a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash4, _ := chainhash.NewHashFromStr("5a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")

	hash01 := *hashLeaves(*hash0, *hash1)
	hash23 := *hashLeaves(*hash2, *hash3)
	hash44 := *hashLeaves(*hash4, *hash4)
	hash0123 := *hashLeaves(hash01, hash23)
	hash4444 := *hashLeaves(hash44, hash44)
	hashMerkleRoot := *hashLeaves(hash0123, hash4444)

	// build merkle tree
	commitments := []chainhash.Hash{*hash0, *hash1, *hash2, *hash3, *hash4}
	merkleTree := buildMerkleTree(commitments)

	// test proofs for different commitments
	proof0 := buildMerkleProof(0, merkleTree)
	assert.Equal(t, *hash0, proof0.Commitment)
	assert.Equal(t, hashMerkleRoot, proof0.MerkleRoot)
	assert.Equal(t, 3, len(proof0.Ops))
	assert.Equal(t, true, proof0.Ops[0].Append)
	assert.Equal(t, *hash1, proof0.Ops[0].Commitment)
	assert.Equal(t, true, proof0.Ops[1].Append)
	assert.Equal(t, hash23, proof0.Ops[1].Commitment)
	assert.Equal(t, true, proof0.Ops[2].Append)
	assert.Equal(t, hash4444, proof0.Ops[2].Commitment)

	proof1 := buildMerkleProof(1, merkleTree)
	assert.Equal(t, *hash1, proof1.Commitment)
	assert.Equal(t, hashMerkleRoot, proof1.MerkleRoot)
	assert.Equal(t, 3, len(proof1.Ops))
	assert.Equal(t, false, proof1.Ops[0].Append)
	assert.Equal(t, *hash0, proof1.Ops[0].Commitment)
	assert.Equal(t, true, proof1.Ops[1].Append)
	assert.Equal(t, hash23, proof1.Ops[1].Commitment)
	assert.Equal(t, true, proof1.Ops[2].Append)
	assert.Equal(t, hash4444, proof1.Ops[2].Commitment)

	proof2 := buildMerkleProof(2, merkleTree)
	assert.Equal(t, *hash2, proof2.Commitment)
	assert.Equal(t, hashMerkleRoot, proof2.MerkleRoot)
	assert.Equal(t, 3, len(proof2.Ops))
	assert.Equal(t, true, proof2.Ops[0].Append)
	assert.Equal(t, *hash3, proof2.Ops[0].Commitment)
	assert.Equal(t, false, proof2.Ops[1].Append)
	assert.Equal(t, hash01, proof2.Ops[1].Commitment)
	assert.Equal(t, true, proof2.Ops[2].Append)
	assert.Equal(t, hash4444, proof2.Ops[2].Commitment)

	proof3 := buildMerkleProof(3, merkleTree)
	assert.Equal(t, *hash3, proof3.Commitment)
	assert.Equal(t, hashMerkleRoot, proof3.MerkleRoot)
	assert.Equal(t, 3, len(proof3.Ops))
	assert.Equal(t, false, proof3.Ops[0].Append)
	assert.Equal(t, *hash2, proof3.Ops[0].Commitment)
	assert.Equal(t, false, proof3.Ops[1].Append)
	assert.Equal(t, hash01, proof3.Ops[1].Commitment)
	assert.Equal(t, true, proof3.Ops[2].Append)
	assert.Equal(t, hash4444, proof3.Ops[2].Commitment)

	proof4 := buildMerkleProof(4, merkleTree)
	assert.Equal(t, *hash4, proof4.Commitment)
	assert.Equal(t, hashMerkleRoot, proof4.MerkleRoot)
	assert.Equal(t, 3, len(proof4.Ops))
	assert.Equal(t, true, proof4.Ops[0].Append)
	assert.Equal(t, *hash4, proof4.Ops[0].Commitment)
	assert.Equal(t, true, proof4.Ops[1].Append)
	assert.Equal(t, hash44, proof4.Ops[1].Commitment)
	assert.Equal(t, false, proof4.Ops[2].Append)
	assert.Equal(t, hash0123, proof4.Ops[2].Commitment)

	// test empty proofs
	proof5 := buildMerkleProof(5, merkleTree)
	assert.Equal(t, CommitmentMerkleProof{}, proof5)
	proof6 := buildMerkleProof(6, merkleTree)
	assert.Equal(t, CommitmentMerkleProof{}, proof6)
	proof7 := buildMerkleProof(7, merkleTree)
	assert.Equal(t, CommitmentMerkleProof{}, proof7)

	// verify that CommitmentMerkleTree arrives to the same result
	commitmentMerkleTree := CommitmentMerkleTree{}
	commitmentMerkleTree.commitments = commitments
	commitmentMerkleTree.updateTreeStore()
	proofs := []CommitmentMerkleProof{proof0, proof1, proof2, proof3, proof4}
	assert.Equal(t, commitmentMerkleTree.getMerkleProofs(), proofs)
}

// Test build merkle proof and verify for 4 commitment tree
func TestMerkleProof_4Commitments(t *testing.T) {
	hash0, _ := chainhash.NewHashFromStr("1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash1, _ := chainhash.NewHashFromStr("2a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash2, _ := chainhash.NewHashFromStr("3a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash3, _ := chainhash.NewHashFromStr("4a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")

	hash01 := *hashLeaves(*hash0, *hash1)
	hash23 := *hashLeaves(*hash2, *hash3)
	hashMerkleRoot := *hashLeaves(hash01, hash23)

	// build merkle tree
	commitments := []chainhash.Hash{*hash0, *hash1, *hash2, *hash3}
	merkleTree := buildMerkleTree(commitments)

	// test proofs for different commitments
	proof0 := buildMerkleProof(0, merkleTree)
	assert.Equal(t, *hash0, proof0.Commitment)
	assert.Equal(t, hashMerkleRoot, proof0.MerkleRoot)
	assert.Equal(t, 2, len(proof0.Ops))
	assert.Equal(t, true, proof0.Ops[0].Append)
	assert.Equal(t, *hash1, proof0.Ops[0].Commitment)
	assert.Equal(t, true, proof0.Ops[1].Append)
	assert.Equal(t, hash23, proof0.Ops[1].Commitment)

	proof1 := buildMerkleProof(1, merkleTree)
	assert.Equal(t, *hash1, proof1.Commitment)
	assert.Equal(t, hashMerkleRoot, proof1.MerkleRoot)
	assert.Equal(t, 2, len(proof1.Ops))
	assert.Equal(t, false, proof1.Ops[0].Append)
	assert.Equal(t, *hash0, proof1.Ops[0].Commitment)
	assert.Equal(t, true, proof1.Ops[1].Append)
	assert.Equal(t, hash23, proof1.Ops[1].Commitment)

	proof2 := buildMerkleProof(2, merkleTree)
	assert.Equal(t, *hash2, proof2.Commitment)
	assert.Equal(t, hashMerkleRoot, proof2.MerkleRoot)
	assert.Equal(t, 2, len(proof2.Ops))
	assert.Equal(t, true, proof2.Ops[0].Append)
	assert.Equal(t, *hash3, proof2.Ops[0].Commitment)
	assert.Equal(t, false, proof2.Ops[1].Append)
	assert.Equal(t, hash01, proof2.Ops[1].Commitment)

	proof3 := buildMerkleProof(3, merkleTree)
	assert.Equal(t, *hash3, proof3.Commitment)
	assert.Equal(t, hashMerkleRoot, proof3.MerkleRoot)
	assert.Equal(t, 2, len(proof3.Ops))
	assert.Equal(t, false, proof3.Ops[0].Append)
	assert.Equal(t, *hash2, proof3.Ops[0].Commitment)
	assert.Equal(t, false, proof3.Ops[1].Append)
	assert.Equal(t, hash01, proof3.Ops[1].Commitment)

	// test empty proofs
	proof4 := buildMerkleProof(4, merkleTree)
	assert.Equal(t, CommitmentMerkleProof{}, proof4)
	proof5 := buildMerkleProof(5, merkleTree)
	assert.Equal(t, CommitmentMerkleProof{}, proof5)
	proof6 := buildMerkleProof(6, merkleTree)
	assert.Equal(t, CommitmentMerkleProof{}, proof6)
	proof7 := buildMerkleProof(7, merkleTree)
	assert.Equal(t, CommitmentMerkleProof{}, proof7)

	// verify that CommitmentMerkleTree arrives to the same result
	commitmentMerkleTree := CommitmentMerkleTree{}
	commitmentMerkleTree.commitments = commitments
	commitmentMerkleTree.updateTreeStore()
	proofs := []CommitmentMerkleProof{proof0, proof1, proof2, proof3}
	assert.Equal(t, commitmentMerkleTree.getMerkleProofs(), proofs)
}

// Test build merkle proof and verify for 3 commitment tree
func TestMerkleProof_3Commitments(t *testing.T) {
	hash0, _ := chainhash.NewHashFromStr("1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash1, _ := chainhash.NewHashFromStr("2a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash2, _ := chainhash.NewHashFromStr("3a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")

	hash01 := *hashLeaves(*hash0, *hash1)
	hash22 := *hashLeaves(*hash2, *hash2)
	hashMerkleRoot := *hashLeaves(hash01, hash22)

	// build merkle tree
	commitments := []chainhash.Hash{*hash0, *hash1, *hash2}
	merkleTree := buildMerkleTree(commitments)

	// test proofs for different commitments
	proof0 := buildMerkleProof(0, merkleTree)
	assert.Equal(t, *hash0, proof0.Commitment)
	assert.Equal(t, hashMerkleRoot, proof0.MerkleRoot)
	assert.Equal(t, 2, len(proof0.Ops))
	assert.Equal(t, true, proof0.Ops[0].Append)
	assert.Equal(t, *hash1, proof0.Ops[0].Commitment)
	assert.Equal(t, true, proof0.Ops[1].Append)
	assert.Equal(t, hash22, proof0.Ops[1].Commitment)

	proof1 := buildMerkleProof(1, merkleTree)
	assert.Equal(t, *hash1, proof1.Commitment)
	assert.Equal(t, hashMerkleRoot, proof1.MerkleRoot)
	assert.Equal(t, 2, len(proof1.Ops))
	assert.Equal(t, false, proof1.Ops[0].Append)
	assert.Equal(t, *hash0, proof1.Ops[0].Commitment)
	assert.Equal(t, true, proof1.Ops[1].Append)
	assert.Equal(t, hash22, proof1.Ops[1].Commitment)

	proof2 := buildMerkleProof(2, merkleTree)
	assert.Equal(t, *hash2, proof2.Commitment)
	assert.Equal(t, hashMerkleRoot, proof2.MerkleRoot)
	assert.Equal(t, 2, len(proof2.Ops))
	assert.Equal(t, true, proof2.Ops[0].Append)
	assert.Equal(t, *hash2, proof2.Ops[0].Commitment)
	assert.Equal(t, false, proof2.Ops[1].Append)
	assert.Equal(t, hash01, proof2.Ops[1].Commitment)

	// test empty proofs
	proof3 := buildMerkleProof(3, merkleTree)
	assert.Equal(t, CommitmentMerkleProof{}, proof3)
	proof9 := buildMerkleProof(9, merkleTree)
	assert.Equal(t, CommitmentMerkleProof{}, proof9)

	// verify that CommitmentMerkleTree arrives to the same result
	commitmentMerkleTree := CommitmentMerkleTree{}
	commitmentMerkleTree.commitments = commitments
	commitmentMerkleTree.updateTreeStore()
	proofs := []CommitmentMerkleProof{proof0, proof1, proof2}
	assert.Equal(t, commitmentMerkleTree.getMerkleProofs(), proofs)
}

// Test build merkle proof and verify for 1 commitment tree
func TestMerkleProof_1Commitments(t *testing.T) {
	hash0, _ := chainhash.NewHashFromStr("1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")

	hashMerkleRoot := *hashLeaves(*hash0, *hash0)

	// build merkle tree
	commitments := []chainhash.Hash{*hash0}
	merkleTree := buildMerkleTree(commitments)

	// test proofs for different commitments
	proof0 := buildMerkleProof(0, merkleTree)
	assert.Equal(t, *hash0, proof0.Commitment)
	assert.Equal(t, hashMerkleRoot, proof0.MerkleRoot)
	assert.Equal(t, 1, len(proof0.Ops))
	assert.Equal(t, true, proof0.Ops[0].Append)
	assert.Equal(t, *hash0, proof0.Ops[0].Commitment)

	proof1 := buildMerkleProof(1, merkleTree)
	assert.Equal(t, CommitmentMerkleProof{}, proof1)

	// test empty proofs
	proof4 := buildMerkleProof(4, merkleTree)
	assert.Equal(t, CommitmentMerkleProof{}, proof4)

	// verify that CommitmentMerkleTree arrives to the same result
	commitmentMerkleTree := CommitmentMerkleTree{}
	commitmentMerkleTree.commitments = commitments
	commitmentMerkleTree.updateTreeStore()
	proofs := []CommitmentMerkleProof{proof0}
	assert.Equal(t, commitmentMerkleTree.getMerkleProofs(), proofs)
}

// Test prove commitment given merkle proof
func TestMerkleProof_ProveCommitment(t *testing.T) {
	hash0, _ := chainhash.NewHashFromStr("1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash1, _ := chainhash.NewHashFromStr("2a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash2, _ := chainhash.NewHashFromStr("3a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash3, _ := chainhash.NewHashFromStr("4a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash4, _ := chainhash.NewHashFromStr("5a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")

	// build merkle tree
	commitments := []chainhash.Hash{*hash0, *hash1, *hash2, *hash3, *hash4}
	merkleTree := buildMerkleTree(commitments)

	// test proving merkle proof with complete ops and partial ops list
	proof0 := buildMerkleProof(0, merkleTree)
	assert.Equal(t, true, ProveMerkleProof(proof0))
	proof0.Ops = proof0.Ops[1:]
	assert.Equal(t, false, ProveMerkleProof(proof0))

	proof1 := buildMerkleProof(1, merkleTree)
	assert.Equal(t, true, ProveMerkleProof(proof1))
	proof0.Ops = proof0.Ops[1:]
	assert.Equal(t, false, ProveMerkleProof(proof0))

	proof2 := buildMerkleProof(2, merkleTree)
	assert.Equal(t, true, ProveMerkleProof(proof2))
	proof2.Ops = proof2.Ops[1:]
	assert.Equal(t, false, ProveMerkleProof(proof2))

	proof3 := buildMerkleProof(3, merkleTree)
	assert.Equal(t, true, ProveMerkleProof(proof3))
	proof3.Ops = proof3.Ops[1:]
	assert.Equal(t, false, ProveMerkleProof(proof3))

	proof4 := buildMerkleProof(4, merkleTree)
	assert.Equal(t, true, ProveMerkleProof(proof4))
	proof4.Ops = proof0.Ops[1:]
	assert.Equal(t, false, ProveMerkleProof(proof4))
}

// Test build merkle proof and verify for 3 commitment tree
func TestMerkleProof_BSON(t *testing.T) {
	hash0, _ := chainhash.NewHashFromStr("1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash1, _ := chainhash.NewHashFromStr("2a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash2, _ := chainhash.NewHashFromStr("3a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")

	// build merkle tree
	commitments := []chainhash.Hash{*hash0, *hash1, *hash2}
	commitmentMerkleTree := CommitmentMerkleTree{}
	commitmentMerkleTree.commitments = commitments
	commitmentMerkleTree.updateTreeStore()

	proofs := commitmentMerkleTree.getMerkleProofs()
	proof0 := proofs[0]

	// test marshal proof model
	bytes, errBytes := proof0.MarshalBSON()
	assert.Equal(t, []byte{0x8b, 0x1, 0x0, 0x0, 0x2, 0x6d, 0x65, 0x72, 0x6b, 0x6c, 0x65, 0x5f, 0x72, 0x6f, 0x6f, 0x74, 0x0, 0x41, 0x0, 0x0, 0x0, 0x62, 0x62, 0x30, 0x38, 0x38, 0x63, 0x31, 0x30, 0x36, 0x62, 0x33, 0x33, 0x37, 0x39, 0x62, 0x36, 0x34, 0x32, 0x34, 0x33, 0x63, 0x31, 0x61, 0x34, 0x39, 0x31, 0x35, 0x66, 0x37, 0x32, 0x61, 0x38, 0x34, 0x37, 0x64, 0x34, 0x35, 0x63, 0x37, 0x35, 0x31, 0x33, 0x62, 0x31, 0x35, 0x32, 0x63, 0x61, 0x64, 0x35, 0x38, 0x33, 0x65, 0x62, 0x33, 0x63, 0x30, 0x61, 0x31, 0x30, 0x36, 0x33, 0x63, 0x32, 0x0, 0x10, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x6d, 0x65, 0x6e, 0x74, 0x0, 0x41, 0x0, 0x0, 0x0, 0x31, 0x61, 0x33, 0x39, 0x65, 0x33, 0x34, 0x65, 0x38, 0x38, 0x31, 0x64, 0x39, 0x61, 0x31, 0x65, 0x36, 0x63, 0x64, 0x63, 0x33, 0x34, 0x31, 0x38, 0x62, 0x35, 0x34, 0x61, 0x61, 0x35, 0x37, 0x37, 0x34, 0x37, 0x31, 0x30, 0x36, 0x62, 0x63, 0x37, 0x35, 0x65, 0x39, 0x65, 0x38, 0x34, 0x34, 0x32, 0x36, 0x36, 0x36, 0x31, 0x66, 0x32, 0x37, 0x66, 0x39, 0x38, 0x61, 0x64, 0x61, 0x33, 0x62, 0x37, 0x0, 0x4, 0x6f, 0x70, 0x73, 0x0, 0xc9, 0x0, 0x0, 0x0, 0x3, 0x30, 0x0, 0x5f, 0x0, 0x0, 0x0, 0x8, 0x61, 0x70, 0x70, 0x65, 0x6e, 0x64, 0x0, 0x1, 0x2, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x6d, 0x65, 0x6e, 0x74, 0x0, 0x41, 0x0, 0x0, 0x0, 0x32, 0x61, 0x33, 0x39, 0x65, 0x33, 0x34, 0x65, 0x38, 0x38, 0x31, 0x64, 0x39, 0x61, 0x31, 0x65, 0x36, 0x63, 0x64, 0x63, 0x33, 0x34, 0x31, 0x38, 0x62, 0x35, 0x34, 0x61, 0x61, 0x35, 0x37, 0x37, 0x34, 0x37, 0x31, 0x30, 0x36, 0x62, 0x63, 0x37, 0x35, 0x65, 0x39, 0x65, 0x38, 0x34, 0x34, 0x32, 0x36, 0x36, 0x36, 0x31, 0x66, 0x32, 0x37, 0x66, 0x39, 0x38, 0x61, 0x64, 0x61, 0x33, 0x62, 0x37, 0x0, 0x0, 0x3, 0x31, 0x0, 0x5f, 0x0, 0x0, 0x0, 0x8, 0x61, 0x70, 0x70, 0x65, 0x6e, 0x64, 0x0, 0x1, 0x2, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x6d, 0x65, 0x6e, 0x74, 0x0, 0x41, 0x0, 0x0, 0x0, 0x31, 0x61, 0x64, 0x37, 0x32, 0x63, 0x63, 0x32, 0x38, 0x38, 0x37, 0x65, 0x62, 0x34, 0x30, 0x32, 0x64, 0x34, 0x35, 0x34, 0x62, 0x31, 0x39, 0x39, 0x32, 0x64, 0x62, 0x30, 0x61, 0x64, 0x61, 0x36, 0x32, 0x30, 0x61, 0x66, 0x66, 0x64, 0x62, 0x37, 0x37, 0x61, 0x34, 0x36, 0x38, 0x34, 0x35, 0x36, 0x31, 0x37, 0x35, 0x32, 0x33, 0x65, 0x38, 0x34, 0x36, 0x62, 0x62, 0x62, 0x63, 0x39, 0x33, 0x35, 0x0, 0x0, 0x0, 0x0}, bytes)
	assert.Equal(t, nil, errBytes)

	// test proof model to document
	doc, docErr := GetDocumentFromModel(proof0)
	assert.Equal(t, nil, docErr)
	assert.Equal(t, proof0.MerkleRoot.String(), doc.Lookup(ProofMerkleRootName).StringValue())
	assert.Equal(t, proof0.ClientPosition, doc.Lookup(ProofClientPositionName).Int32())
	assert.Equal(t, proof0.Commitment.String(), doc.Lookup(ProofCommitmentName).StringValue())

	for pos := range proof0.Ops {
		arrVal := doc.Lookup(ProofOpsName).Array()[uint(pos)]
		docOp, docOpErr := GetDocumentFromModel(arrVal)
		assert.Equal(t, nil, docOpErr)
		assert.Equal(t, proof0.Ops[pos].Append, docOp.Lookup(ProofOpAppendName).Boolean())
		assert.Equal(t, proof0.Ops[pos].Commitment.String(), docOp.Lookup(ProofOpCommitmentName).StringValue())
	}
}
