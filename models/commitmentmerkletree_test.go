// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package models

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/assert"
)

// Test nextPow(n) function
func TestNextPow(t *testing.T) {
	assert.Equal(t, 0, nextPow(-1))
	assert.Equal(t, 0, nextPow(0))
	assert.Equal(t, 2, nextPow(1))
	assert.Equal(t, 2, nextPow(2))
	assert.Equal(t, 4, nextPow(3))
	assert.Equal(t, 8, nextPow(5))
}

// Test hashLeaves(hash, hash) function
func TestHashleaves(t *testing.T) {
	hash0, _ := chainhash.NewHashFromStr("1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash1, _ := chainhash.NewHashFromStr("2a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")

	assert.Equal(t, "2b6689ee13e50cb4d79392fdd8ac71aa451823ae521964e069aad8810369ef5a", hashLeaves(*hash0, *hash1).String())
}

// Test build merkle tree for 5 commitment tree
func TestMerkleTree_5Commitments(t *testing.T) {
	hash0, _ := chainhash.NewHashFromStr("1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash1, _ := chainhash.NewHashFromStr("2a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash2, _ := chainhash.NewHashFromStr("3a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash3, _ := chainhash.NewHashFromStr("4a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash4, _ := chainhash.NewHashFromStr("5a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")

	// test partial merkle tree with 5 hashes
	partialCommitments5 := []chainhash.Hash{*hash0, *hash1, *hash2, *hash3, *hash4}
	partialMerkleTree5 := buildMerkleTree(partialCommitments5)

	assert.Equal(t, 15, len(partialMerkleTree5))
	assert.Equal(t, hash0, partialMerkleTree5[0])
	assert.Equal(t, hash1, partialMerkleTree5[1])
	assert.Equal(t, hash2, partialMerkleTree5[2])
	assert.Equal(t, hash3, partialMerkleTree5[3])
	assert.Equal(t, hash4, partialMerkleTree5[4])
	assert.Equal(t, (*chainhash.Hash)(nil), partialMerkleTree5[5])
	assert.Equal(t, (*chainhash.Hash)(nil), partialMerkleTree5[6])
	assert.Equal(t, (*chainhash.Hash)(nil), partialMerkleTree5[7])
	assert.Equal(t, hashLeaves(*hash0, *hash1), partialMerkleTree5[8])
	assert.Equal(t, hashLeaves(*hash2, *hash3), partialMerkleTree5[9])
	assert.Equal(t, hashLeaves(*hash4, *hash4), partialMerkleTree5[10])
	assert.Equal(t, (*chainhash.Hash)(nil), partialMerkleTree5[11])
	assert.Equal(t, hashLeaves(*hashLeaves(*hash0, *hash1), *hashLeaves(*hash2, *hash3)), partialMerkleTree5[12])
	assert.Equal(t, hashLeaves(*hashLeaves(*hash4, *hash4), *hashLeaves(*hash4, *hash4)), partialMerkleTree5[13])
	assert.Equal(t, hashLeaves(*hashLeaves(*hashLeaves(*hash0, *hash1), *hashLeaves(*hash2, *hash3)),
		*hashLeaves(*hashLeaves(*hash4, *hash4), *hashLeaves(*hash4, *hash4))),
		partialMerkleTree5[14])

	// verify that CommitmentMerkleTree arrives to the same result
	partialCommitmentMerkleTree5 := CommitmentMerkleTree{}
	partialCommitmentMerkleTree5.commitments = partialCommitments5
	partialCommitmentMerkleTree5.updateTreeStore()
	assert.Equal(t, partialCommitmentMerkleTree5.getMerkleRoot(), *partialMerkleTree5[14])
}

// Test build merkle tree for 4 commitment tree
func TestMerkleTree_4Commitments(t *testing.T) {
	hash0, _ := chainhash.NewHashFromStr("1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash1, _ := chainhash.NewHashFromStr("2a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash2, _ := chainhash.NewHashFromStr("3a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash3, _ := chainhash.NewHashFromStr("4a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")

	// test full merkle tree with all 4 hashes
	commitments := []chainhash.Hash{*hash0, *hash1, *hash2, *hash3}
	merkleTree := buildMerkleTree(commitments)
	assert.Equal(t, 7, len(merkleTree))
	assert.Equal(t, hash0, merkleTree[0])
	assert.Equal(t, hash1, merkleTree[1])
	assert.Equal(t, hash2, merkleTree[2])
	assert.Equal(t, hash3, merkleTree[3])
	assert.Equal(t, hashLeaves(*hash0, *hash1), merkleTree[4])
	assert.Equal(t, hashLeaves(*hash2, *hash3), merkleTree[5])
	assert.Equal(t, hashLeaves(*hashLeaves(*hash0, *hash1), *hashLeaves(*hash2, *hash3)), merkleTree[6])

	// verify that CommitmentMerkleTree arrives to the same result
	commitmentMerkleTree := CommitmentMerkleTree{}
	commitmentMerkleTree.commitments = commitments
	commitmentMerkleTree.updateTreeStore()
	assert.Equal(t, commitmentMerkleTree.getMerkleRoot(), *merkleTree[6])
}

// Test build merkle tree for 3 commitment tree
func TestMerkleTree_3Commitments(t *testing.T) {
	hash0, _ := chainhash.NewHashFromStr("1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash1, _ := chainhash.NewHashFromStr("2a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash2, _ := chainhash.NewHashFromStr("3a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")

	// test partial merkle tree with 3 hashes
	partialCommitments := []chainhash.Hash{*hash0, *hash1, *hash2}
	partialMerkleTree := buildMerkleTree(partialCommitments)

	assert.Equal(t, 7, len(partialMerkleTree))
	assert.Equal(t, hash0, partialMerkleTree[0])
	assert.Equal(t, hash1, partialMerkleTree[1])
	assert.Equal(t, hash2, partialMerkleTree[2])
	assert.Equal(t, (*chainhash.Hash)(nil), partialMerkleTree[3])
	assert.Equal(t, hashLeaves(*hash0, *hash1), partialMerkleTree[4])
	assert.Equal(t, hashLeaves(*hash2, *hash2), partialMerkleTree[5])
	assert.Equal(t, hashLeaves(*hashLeaves(*hash0, *hash1), *hashLeaves(*hash2, *hash2)), partialMerkleTree[6])

	// verify that CommitmentMerkleTree arrives to the same result
	partialCommitmentMerkleTree := CommitmentMerkleTree{}
	partialCommitmentMerkleTree.commitments = partialCommitments
	partialCommitmentMerkleTree.updateTreeStore()
	assert.Equal(t, partialCommitmentMerkleTree.getMerkleRoot(), *partialMerkleTree[6])
}

// Test build merkle tree for 1 commitment tree
func TestMerkleTree_1Commitments(t *testing.T) {
	hash0, _ := chainhash.NewHashFromStr("1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")

	// test partial merkle tree with 1 hash
	partialCommitments := []chainhash.Hash{*hash0}
	partialMerkleTree := buildMerkleTree(partialCommitments)

	assert.Equal(t, 3, len(partialMerkleTree))
	assert.Equal(t, hash0, partialMerkleTree[0])
	assert.Equal(t, (*chainhash.Hash)(nil), partialMerkleTree[1])
	assert.Equal(t, hashLeaves(*hash0, *hash0), partialMerkleTree[2])

	// verify that CommitmentMerkleTree arrives to the same result
	partialCommitmentMerkleTree := CommitmentMerkleTree{}
	partialCommitmentMerkleTree.commitments = partialCommitments
	partialCommitmentMerkleTree.updateTreeStore()
	assert.Equal(t, partialCommitmentMerkleTree.getMerkleRoot(), *partialMerkleTree[2])
}
