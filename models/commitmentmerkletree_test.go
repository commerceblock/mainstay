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
	hashRoot := *hashLeaves(hash0123, hash4444)

	// build merkle tree
	commitments := []chainhash.Hash{*hash0, *hash1, *hash2, *hash3, *hash4}
	merkleTree := buildMerkleTree(commitments)

	// test proofs for different commitments
	proof0 := buildMerkleProof(0, merkleTree)
	assert.Equal(t, *hash0, proof0.Commitment)
	assert.Equal(t, hashRoot, proof0.Root)
	assert.Equal(t, 3, len(proof0.Ops))
	assert.Equal(t, true, proof0.Ops[0].Append)
	assert.Equal(t, *hash1, proof0.Ops[0].Commitment)
	assert.Equal(t, true, proof0.Ops[1].Append)
	assert.Equal(t, hash23, proof0.Ops[1].Commitment)
	assert.Equal(t, true, proof0.Ops[2].Append)
	assert.Equal(t, hash4444, proof0.Ops[2].Commitment)

	proof1 := buildMerkleProof(1, merkleTree)
	assert.Equal(t, *hash1, proof1.Commitment)
	assert.Equal(t, hashRoot, proof1.Root)
	assert.Equal(t, 3, len(proof1.Ops))
	assert.Equal(t, false, proof1.Ops[0].Append)
	assert.Equal(t, *hash0, proof1.Ops[0].Commitment)
	assert.Equal(t, true, proof1.Ops[1].Append)
	assert.Equal(t, hash23, proof1.Ops[1].Commitment)
	assert.Equal(t, true, proof1.Ops[2].Append)
	assert.Equal(t, hash4444, proof1.Ops[2].Commitment)

	proof2 := buildMerkleProof(2, merkleTree)
	assert.Equal(t, *hash2, proof2.Commitment)
	assert.Equal(t, hashRoot, proof2.Root)
	assert.Equal(t, 3, len(proof2.Ops))
	assert.Equal(t, true, proof2.Ops[0].Append)
	assert.Equal(t, *hash3, proof2.Ops[0].Commitment)
	assert.Equal(t, false, proof2.Ops[1].Append)
	assert.Equal(t, hash01, proof2.Ops[1].Commitment)
	assert.Equal(t, true, proof2.Ops[2].Append)
	assert.Equal(t, hash4444, proof2.Ops[2].Commitment)

	proof3 := buildMerkleProof(3, merkleTree)
	assert.Equal(t, *hash3, proof3.Commitment)
	assert.Equal(t, hashRoot, proof3.Root)
	assert.Equal(t, 3, len(proof3.Ops))
	assert.Equal(t, false, proof3.Ops[0].Append)
	assert.Equal(t, *hash2, proof3.Ops[0].Commitment)
	assert.Equal(t, false, proof3.Ops[1].Append)
	assert.Equal(t, hash01, proof3.Ops[1].Commitment)
	assert.Equal(t, true, proof3.Ops[2].Append)
	assert.Equal(t, hash4444, proof3.Ops[2].Commitment)

	proof4 := buildMerkleProof(4, merkleTree)
	assert.Equal(t, *hash4, proof4.Commitment)
	assert.Equal(t, hashRoot, proof4.Root)
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
}

// Test build merkle proof and verify for 4 commitment tree
func TestMerkleProof_4Commitments(t *testing.T) {
	hash0, _ := chainhash.NewHashFromStr("1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash1, _ := chainhash.NewHashFromStr("2a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash2, _ := chainhash.NewHashFromStr("3a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash3, _ := chainhash.NewHashFromStr("4a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")

	hash01 := *hashLeaves(*hash0, *hash1)
	hash23 := *hashLeaves(*hash2, *hash3)
	hashRoot := *hashLeaves(hash01, hash23)

	// build merkle tree
	commitments := []chainhash.Hash{*hash0, *hash1, *hash2, *hash3}
	merkleTree := buildMerkleTree(commitments)

	// test proofs for different commitments
	proof0 := buildMerkleProof(0, merkleTree)
	assert.Equal(t, *hash0, proof0.Commitment)
	assert.Equal(t, hashRoot, proof0.Root)
	assert.Equal(t, 2, len(proof0.Ops))
	assert.Equal(t, true, proof0.Ops[0].Append)
	assert.Equal(t, *hash1, proof0.Ops[0].Commitment)
	assert.Equal(t, true, proof0.Ops[1].Append)
	assert.Equal(t, hash23, proof0.Ops[1].Commitment)

	proof1 := buildMerkleProof(1, merkleTree)
	assert.Equal(t, *hash1, proof1.Commitment)
	assert.Equal(t, hashRoot, proof1.Root)
	assert.Equal(t, 2, len(proof1.Ops))
	assert.Equal(t, false, proof1.Ops[0].Append)
	assert.Equal(t, *hash0, proof1.Ops[0].Commitment)
	assert.Equal(t, true, proof1.Ops[1].Append)
	assert.Equal(t, hash23, proof1.Ops[1].Commitment)

	proof2 := buildMerkleProof(2, merkleTree)
	assert.Equal(t, *hash2, proof2.Commitment)
	assert.Equal(t, hashRoot, proof2.Root)
	assert.Equal(t, 2, len(proof2.Ops))
	assert.Equal(t, true, proof2.Ops[0].Append)
	assert.Equal(t, *hash3, proof2.Ops[0].Commitment)
	assert.Equal(t, false, proof2.Ops[1].Append)
	assert.Equal(t, hash01, proof2.Ops[1].Commitment)

	proof3 := buildMerkleProof(3, merkleTree)
	assert.Equal(t, *hash3, proof3.Commitment)
	assert.Equal(t, hashRoot, proof3.Root)
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
}

// Test build merkle proof and verify for 3 commitment tree
func TestMerkleProof_3Commitments(t *testing.T) {
	hash0, _ := chainhash.NewHashFromStr("1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash1, _ := chainhash.NewHashFromStr("2a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash2, _ := chainhash.NewHashFromStr("3a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")

	hash01 := *hashLeaves(*hash0, *hash1)
	hash22 := *hashLeaves(*hash2, *hash2)
	hashRoot := *hashLeaves(hash01, hash22)

	// build merkle tree
	commitments := []chainhash.Hash{*hash0, *hash1, *hash2}
	merkleTree := buildMerkleTree(commitments)

	// test proofs for different commitments
	proof0 := buildMerkleProof(0, merkleTree)
	assert.Equal(t, *hash0, proof0.Commitment)
	assert.Equal(t, hashRoot, proof0.Root)
	assert.Equal(t, 2, len(proof0.Ops))
	assert.Equal(t, true, proof0.Ops[0].Append)
	assert.Equal(t, *hash1, proof0.Ops[0].Commitment)
	assert.Equal(t, true, proof0.Ops[1].Append)
	assert.Equal(t, hash22, proof0.Ops[1].Commitment)

	proof1 := buildMerkleProof(1, merkleTree)
	assert.Equal(t, *hash1, proof1.Commitment)
	assert.Equal(t, hashRoot, proof1.Root)
	assert.Equal(t, 2, len(proof1.Ops))
	assert.Equal(t, false, proof1.Ops[0].Append)
	assert.Equal(t, *hash0, proof1.Ops[0].Commitment)
	assert.Equal(t, true, proof1.Ops[1].Append)
	assert.Equal(t, hash22, proof1.Ops[1].Commitment)

	proof2 := buildMerkleProof(2, merkleTree)
	assert.Equal(t, *hash2, proof2.Commitment)
	assert.Equal(t, hashRoot, proof2.Root)
	assert.Equal(t, 2, len(proof2.Ops))
	assert.Equal(t, true, proof2.Ops[0].Append)
	assert.Equal(t, *hash2, proof2.Ops[0].Commitment)
	assert.Equal(t, false, proof2.Ops[1].Append)
	assert.Equal(t, hash01, proof2.Ops[1].Commitment)

	proof3 := buildMerkleProof(3, merkleTree)
	assert.Equal(t, CommitmentMerkleProof{}, proof3)
	proof9 := buildMerkleProof(9, merkleTree)
	assert.Equal(t, CommitmentMerkleProof{}, proof9)
}

// Test build merkle proof and verify for 1 commitment tree
func TestMerkleProof_1Commitments(t *testing.T) {
	hash0, _ := chainhash.NewHashFromStr("1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")

	hashRoot := *hashLeaves(*hash0, *hash0)

	// build merkle tree
	commitments := []chainhash.Hash{*hash0}
	merkleTree := buildMerkleTree(commitments)

	// test proofs for different commitments
	proof0 := buildMerkleProof(0, merkleTree)
	assert.Equal(t, *hash0, proof0.Commitment)
	assert.Equal(t, hashRoot, proof0.Root)
	assert.Equal(t, 1, len(proof0.Ops))
	assert.Equal(t, true, proof0.Ops[0].Append)
	assert.Equal(t, *hash0, proof0.Ops[0].Commitment)

	proof1 := buildMerkleProof(1, merkleTree)
	assert.Equal(t, CommitmentMerkleProof{}, proof1)
	proof4 := buildMerkleProof(4, merkleTree)
	assert.Equal(t, CommitmentMerkleProof{}, proof4)
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
	assert.Equal(t, true, proveMerkleProof(proof0))
	proof0.Ops = proof0.Ops[1:]
	assert.Equal(t, false, proveMerkleProof(proof0))

	proof1 := buildMerkleProof(1, merkleTree)
	assert.Equal(t, true, proveMerkleProof(proof1))
	proof0.Ops = proof0.Ops[1:]
	assert.Equal(t, false, proveMerkleProof(proof0))

	proof2 := buildMerkleProof(2, merkleTree)
	assert.Equal(t, true, proveMerkleProof(proof2))
	proof2.Ops = proof2.Ops[1:]
	assert.Equal(t, false, proveMerkleProof(proof2))

	proof3 := buildMerkleProof(3, merkleTree)
	assert.Equal(t, true, proveMerkleProof(proof3))
	proof3.Ops = proof3.Ops[1:]
	assert.Equal(t, false, proveMerkleProof(proof3))

	proof4 := buildMerkleProof(4, merkleTree)
	assert.Equal(t, true, proveMerkleProof(proof4))
	proof4.Ops = proof0.Ops[1:]
	assert.Equal(t, false, proveMerkleProof(proof4))
}
