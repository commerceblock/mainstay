// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	"errors"
	"testing"

	"mainstay/db"
	"mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/assert"
)

// Test AttestServer UpdateLatestAttestation with no latest commitment
func TestAttestServerUpdateLatestAttestation_NoClientCommitments(t *testing.T) {
	// TEST INIT
	dbFake := db.NewDbFake()
	server := NewAttestServer(dbFake)

	respClientCommitment := (*models.Commitment)(nil)
	txid, _ := chainhash.NewHashFromStr("11111111111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latest := models.NewAttestation(*txid, respClientCommitment)
	latest.Confirmed = true

	// Test update latest attestation
	errUpdate := server.UpdateLatestAttestation(*latest)
	assert.Equal(t, errors.New(models.ErrorCommitmentNotDefined), errUpdate)
}

// Test AttestServer UpdateLatestAttestation with 1 latest commitment
func TestAttestServerUpdateLatestAttestation_1ClientCommitments(t *testing.T) {
	// TEST INIT
	dbFake := db.NewDbFake()
	server := NewAttestServer(dbFake)

	// set db latest commitment
	hash0, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitments := []models.ClientCommitment{models.ClientCommitment{*hash0, 0}}
	latestCommitment, _ := models.NewCommitment([]chainhash.Hash{*hash0})
	dbFake.SetClientCommitments(latestCommitments)

	// Test latest attestation request
	respAttestationHash, errAttestation := server.GetLatestAttestationCommitmentHash()
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, chainhash.Hash{}, respAttestationHash)
	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash(true)
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, chainhash.Hash{}, respAttestationHash)
	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash(false)
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, chainhash.Hash{}, respAttestationHash)

	// Generate new attestation and update server
	respClientCommitment, err := server.GetClientCommitment()
	assert.Equal(t, nil, err)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respClientCommitment.GetCommitmentHash())

	txid, _ := chainhash.NewHashFromStr("11111111111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latest := models.NewAttestation(*txid, &respClientCommitment)

	// Test update latest attestation unconfirmed
	errUpdate := server.UpdateLatestAttestation(*latest)
	assert.Equal(t, nil, errUpdate)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash()
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, chainhash.Hash{}, respAttestationHash)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash(true)
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, chainhash.Hash{}, respAttestationHash)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash(false)
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respAttestationHash)

	assert.Equal(t, 0, len(dbFake.AttestationsInfo))

	// Test update latest attestation confirmed
	latest.Confirmed = true
	latest.Info = models.AttestationInfo{
		Txid:      "11111111111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
		Blockhash: "abcde34e881d9a1e6cdc3418b54bb57747106bc75e9e84426661f27f98ada3b7",
		Amount:    int64(1),
		Time:      int64(1542121293)}
	errUpdate = server.UpdateLatestAttestation(*latest)
	assert.Equal(t, nil, errUpdate)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash()
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respAttestationHash)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash(true)
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respAttestationHash)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash(false)
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, chainhash.Hash{}, respAttestationHash)

	assert.Equal(t, 1, len(dbFake.AttestationsInfo))
	assert.Equal(t, models.AttestationInfo{
		Txid:      "11111111111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
		Blockhash: "abcde34e881d9a1e6cdc3418b54bb57747106bc75e9e84426661f27f98ada3b7",
		Amount:    int64(1),
		Time:      int64(1542121293)}, dbFake.AttestationsInfo[0])

	// Test db updated correctly
	assert.Equal(t, *txid, dbFake.Attestations[0].Txid)
	assert.Equal(t, true, dbFake.Attestations[0].Confirmed)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), dbFake.Attestations[0].CommitmentHash())

	assert.Equal(t, latestCommitment.GetCommitmentHash(), dbFake.MerkleCommitments[0].MerkleRoot)
	assert.Equal(t, int32(0), dbFake.MerkleCommitments[0].ClientPosition)
	assert.Equal(t, *hash0, dbFake.MerkleCommitments[0].Commitment)

	assert.Equal(t, latestCommitment.GetCommitmentHash(), dbFake.MerkleProofs[0].MerkleRoot)
	assert.Equal(t, int32(0), dbFake.MerkleProofs[0].ClientPosition)
	assert.Equal(t, *hash0, dbFake.MerkleProofs[0].Commitment)
	assert.Equal(t, true, dbFake.MerkleProofs[0].Ops[0].Append)
	assert.Equal(t, *hash0, dbFake.MerkleProofs[0].Ops[0].Commitment)

	// add an additional unconfirmed attestation
	// set db latest commitment
	hash2, _ := chainhash.NewHashFromStr("baaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitments2 := []models.ClientCommitment{models.ClientCommitment{*hash2, 0}}
	latestCommitment2, _ := models.NewCommitment([]chainhash.Hash{*hash2})
	dbFake.SetClientCommitments(latestCommitments2)

	// Generate new attestation and update server
	respClientCommitment2, err2 := server.GetClientCommitment()
	assert.Equal(t, nil, err2)
	assert.Equal(t, latestCommitment2.GetCommitmentHash(), respClientCommitment2.GetCommitmentHash())

	// update with latest unconfirmed
	txid2, _ := chainhash.NewHashFromStr("23311111111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latest2 := models.NewAttestation(*txid2, &respClientCommitment2)
	errUpdate = server.UpdateLatestAttestation(*latest2)
	assert.Equal(t, nil, errUpdate)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash()
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respAttestationHash)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash(true)
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respAttestationHash)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash(false)
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, latestCommitment2.GetCommitmentHash(), respAttestationHash)
}

// Test AttestServer UpdateLatestAttestation with 3 latest commitment
func TestAttestServerUpdateLatestAttestation_3ClientCommitments(t *testing.T) {
	// TEST INIT
	dbFake := db.NewDbFake()
	server := NewAttestServer(dbFake)

	// set db latest commitment
	hash0, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash1, _ := chainhash.NewHashFromStr("baaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash01, _ := chainhash.NewHashFromStr("f6dab9f1bfb9ba3f33178e040ff78ff79bc875bfb383ce6db28f46b8226ca073")
	hash2, _ := chainhash.NewHashFromStr("caaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash22, _ := chainhash.NewHashFromStr("e0ae56a5a7eec5de827346ea45dd3d834c006d12e333d0d949aa974dda4928ed")
	latestCommitments := []models.ClientCommitment{
		models.ClientCommitment{*hash0, 0},
		models.ClientCommitment{*hash1, 1},
		models.ClientCommitment{*hash2, 2}}
	latestCommitment, _ := models.NewCommitment([]chainhash.Hash{*hash0, *hash1, *hash2})
	dbFake.SetClientCommitments(latestCommitments)

	// Test latest attestation request
	respAttestationHash, errAttestation := server.GetLatestAttestationCommitmentHash()
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, chainhash.Hash{}, respAttestationHash)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash(true)
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, chainhash.Hash{}, respAttestationHash)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash(false)
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, chainhash.Hash{}, respAttestationHash)

	// Generate new attestation and update server
	respClientCommitment, err := server.GetClientCommitment()
	assert.Equal(t, nil, err)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respClientCommitment.GetCommitmentHash())

	txid, _ := chainhash.NewHashFromStr("11111111111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latest := models.NewAttestation(*txid, &respClientCommitment)

	// Test update latest attestation unconfirmed
	errUpdate := server.UpdateLatestAttestation(*latest)
	assert.Equal(t, nil, errUpdate)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash()
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, chainhash.Hash{}, respAttestationHash)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash(true)
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, chainhash.Hash{}, respAttestationHash)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash(false)
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respAttestationHash)

	assert.Equal(t, 0, len(dbFake.AttestationsInfo))

	// Test update latest attestation confirmed
	latest.Confirmed = true
	latest.Info = models.AttestationInfo{
		Txid:      "11111111111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
		Blockhash: "abcde34e881d9a1e6cdc3418b54bb57747106bc75e9e84426661f27f98ada3b7",
		Amount:    int64(1),
		Time:      int64(1542121293)}
	errUpdate = server.UpdateLatestAttestation(*latest)
	assert.Equal(t, nil, errUpdate)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash()
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respAttestationHash)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash(true)
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respAttestationHash)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash(false)
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, chainhash.Hash{}, respAttestationHash)

	assert.Equal(t, 1, len(dbFake.AttestationsInfo))
	assert.Equal(t, models.AttestationInfo{
		Txid:      "11111111111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
		Blockhash: "abcde34e881d9a1e6cdc3418b54bb57747106bc75e9e84426661f27f98ada3b7",
		Amount:    int64(1),
		Time:      int64(1542121293)}, dbFake.AttestationsInfo[0])

	// Test db updated correctly
	assert.Equal(t, *txid, dbFake.Attestations[0].Txid)
	assert.Equal(t, true, dbFake.Attestations[0].Confirmed)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), dbFake.Attestations[0].CommitmentHash())

	assert.Equal(t, latestCommitment.GetCommitmentHash(), dbFake.MerkleCommitments[0].MerkleRoot)
	assert.Equal(t, int32(0), dbFake.MerkleCommitments[0].ClientPosition)
	assert.Equal(t, *hash0, dbFake.MerkleCommitments[0].Commitment)

	assert.Equal(t, latestCommitment.GetCommitmentHash(), dbFake.MerkleCommitments[1].MerkleRoot)
	assert.Equal(t, int32(1), dbFake.MerkleCommitments[1].ClientPosition)
	assert.Equal(t, *hash1, dbFake.MerkleCommitments[1].Commitment)

	assert.Equal(t, latestCommitment.GetCommitmentHash(), dbFake.MerkleCommitments[2].MerkleRoot)
	assert.Equal(t, int32(2), dbFake.MerkleCommitments[2].ClientPosition)
	assert.Equal(t, *hash2, dbFake.MerkleCommitments[2].Commitment)

	assert.Equal(t, latestCommitment.GetCommitmentHash(), dbFake.MerkleProofs[0].MerkleRoot)
	assert.Equal(t, int32(0), dbFake.MerkleProofs[0].ClientPosition)
	assert.Equal(t, *hash0, dbFake.MerkleProofs[0].Commitment)
	assert.Equal(t, true, dbFake.MerkleProofs[0].Ops[0].Append)
	assert.Equal(t, *hash1, dbFake.MerkleProofs[0].Ops[0].Commitment)
	assert.Equal(t, true, dbFake.MerkleProofs[0].Ops[1].Append)
	assert.Equal(t, *hash01, dbFake.MerkleProofs[0].Ops[1].Commitment)

	assert.Equal(t, latestCommitment.GetCommitmentHash(), dbFake.MerkleProofs[1].MerkleRoot)
	assert.Equal(t, int32(1), dbFake.MerkleProofs[1].ClientPosition)
	assert.Equal(t, *hash1, dbFake.MerkleProofs[1].Commitment)
	assert.Equal(t, false, dbFake.MerkleProofs[1].Ops[0].Append)
	assert.Equal(t, *hash0, dbFake.MerkleProofs[1].Ops[0].Commitment)
	assert.Equal(t, true, dbFake.MerkleProofs[1].Ops[1].Append)
	assert.Equal(t, *hash01, dbFake.MerkleProofs[1].Ops[1].Commitment)

	assert.Equal(t, latestCommitment.GetCommitmentHash(), dbFake.MerkleProofs[2].MerkleRoot)
	assert.Equal(t, int32(2), dbFake.MerkleProofs[2].ClientPosition)
	assert.Equal(t, *hash2, dbFake.MerkleProofs[2].Commitment)
	assert.Equal(t, true, dbFake.MerkleProofs[2].Ops[0].Append)
	assert.Equal(t, *hash2, dbFake.MerkleProofs[2].Ops[0].Commitment)
	assert.Equal(t, false, dbFake.MerkleProofs[2].Ops[1].Append)
	assert.Equal(t, *hash22, dbFake.MerkleProofs[2].Ops[1].Commitment)

	// add an additional unconfirmed attestation
	// set db latest commitment
	hashX, _ := chainhash.NewHashFromStr("baaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hashY, _ := chainhash.NewHashFromStr("caaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hashZ, _ := chainhash.NewHashFromStr("daaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitments2 := []models.ClientCommitment{
		models.ClientCommitment{*hashX, 0},
		models.ClientCommitment{*hashY, 1},
		models.ClientCommitment{*hashZ, 2}}
	latestCommitment2, _ := models.NewCommitment([]chainhash.Hash{*hashX, *hashY, *hashZ})
	dbFake.SetClientCommitments(latestCommitments2)

	// Generate new attestation and update server
	respClientCommitment2, err2 := server.GetClientCommitment()
	assert.Equal(t, nil, err2)
	assert.Equal(t, latestCommitment2.GetCommitmentHash(), respClientCommitment2.GetCommitmentHash())

	// update with latest unconfirmed
	txid2, _ := chainhash.NewHashFromStr("23311111111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latest2 := models.NewAttestation(*txid2, &respClientCommitment2)
	errUpdate = server.UpdateLatestAttestation(*latest2)
	assert.Equal(t, nil, errUpdate)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash()
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respAttestationHash)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash(true)
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respAttestationHash)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash(false)
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, latestCommitment2.GetCommitmentHash(), respAttestationHash)
}

// Test AttestServer GetClientCommitment
func TestAttestServerGetClientCommitment(t *testing.T) {
	// TEST INIT
	dbFake := db.NewDbFake()
	server := NewAttestServer(dbFake)

	// check empty latest commitment first
	respClientCommitment, err := server.GetClientCommitment()
	assert.Equal(t, errors.New(models.ErrorCommitmentListEmpty), err)

	// set db latest commitment
	hash0, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash1, _ := chainhash.NewHashFromStr("baaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash2, _ := chainhash.NewHashFromStr("caaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")

	// update server with incorrect latest commitment and test server
	latestCommitments := []models.ClientCommitment{
		models.ClientCommitment{*hash0, 0}, models.ClientCommitment{*hash2, 2}}
	dbFake.SetClientCommitments(latestCommitments)

	respClientCommitment, err = server.GetClientCommitment()
	assert.Equal(t, nil, err)
	latestCommitment, err2 := models.NewCommitment([]chainhash.Hash{*hash0, chainhash.Hash{}, *hash2})
	assert.Equal(t, nil, err2)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respClientCommitment.GetCommitmentHash())

	// update server with incorrect latest commitment and test server
	latestCommitments = []models.ClientCommitment{
		models.ClientCommitment{*hash1, 1}, models.ClientCommitment{*hash2, 2}}
	dbFake.SetClientCommitments(latestCommitments)

	respClientCommitment, err = server.GetClientCommitment()
	assert.Equal(t, nil, err)
	latestCommitment, err2 = models.NewCommitment([]chainhash.Hash{chainhash.Hash{}, *hash1, *hash2})
	assert.Equal(t, nil, err2)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respClientCommitment.GetCommitmentHash())

	// update server with incorrect latest commitment and test server
	latestCommitments = []models.ClientCommitment{models.ClientCommitment{*hash2, 2}}
	dbFake.SetClientCommitments(latestCommitments)

	respClientCommitment, err = server.GetClientCommitment()
	assert.Equal(t, nil, err)
	latestCommitment, err2 = models.NewCommitment([]chainhash.Hash{chainhash.Hash{}, chainhash.Hash{}, *hash2})
	assert.Equal(t, nil, err2)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respClientCommitment.GetCommitmentHash())

	// update server with correct latest commitment and test server
	latestCommitments = []models.ClientCommitment{
		models.ClientCommitment{*hash0, 0},
		models.ClientCommitment{*hash1, 1},
		models.ClientCommitment{*hash2, 2}}
	latestCommitment, err2 = models.NewCommitment([]chainhash.Hash{*hash0, *hash1, *hash2})
	assert.Equal(t, nil, err2)
	dbFake.SetClientCommitments(latestCommitments)

	respClientCommitment, err = server.GetClientCommitment()
	assert.Equal(t, nil, err)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respClientCommitment.GetCommitmentHash())
}

// Test AttestServer GetAttestationCommitment
func TestAttestServerGetAttestationCommitment(t *testing.T) {
	//TEST INIT
	dbFake := db.NewDbFake()
	server := NewAttestServer(dbFake)

	// set db latest commitment
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hashY, _ := chainhash.NewHashFromStr("baaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hashZ, _ := chainhash.NewHashFromStr("caaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")

	// check empty attestation first
	commitment, err := server.GetAttestationCommitment(chainhash.Hash{})
	assert.Equal(t, nil, err)
	assert.Equal(t, chainhash.Hash{}, commitment.GetCommitmentHash())

	commitment, err = server.GetAttestationCommitment(chainhash.Hash{}, true)
	assert.Equal(t, nil, err)
	assert.Equal(t, chainhash.Hash{}, commitment.GetCommitmentHash())

	commitment, err = server.GetAttestationCommitment(chainhash.Hash{}, false)
	assert.Equal(t, errors.New(models.ErrorCommitmentListEmpty), err)

	// update attestation to server
	latestCommitments0 := []models.ClientCommitment{
		models.ClientCommitment{*hashX, 0},
		models.ClientCommitment{*hashY, 1},
		models.ClientCommitment{*hashZ, 2}}
	dbFake.SetClientCommitments(latestCommitments0)
	latestCommitment0, _ := models.NewCommitment([]chainhash.Hash{*hashX, *hashY, *hashZ})

	txid0, _ := chainhash.NewHashFromStr("11111111111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latest0 := models.NewAttestation(*txid0, latestCommitment0)
	latest0.Confirmed = true
	errUpdate := server.UpdateLatestAttestation(*latest0)
	assert.Equal(t, nil, errUpdate)

	// check commitment for new attestation
	commitment, err = server.GetAttestationCommitment(*txid0)
	assert.Equal(t, nil, err)
	assert.Equal(t, latestCommitment0.GetCommitmentHash(), commitment.GetCommitmentHash())

	// check commitment for new attestation
	commitment, err = server.GetAttestationCommitment(*txid0, true)
	assert.Equal(t, nil, err)
	assert.Equal(t, latestCommitment0.GetCommitmentHash(), commitment.GetCommitmentHash())

	// check commitment for new attestation
	commitment, err = server.GetAttestationCommitment(*txid0, false)
	assert.Equal(t, nil, err)
	assert.Equal(t, latestCommitment0.GetCommitmentHash(), commitment.GetCommitmentHash())

	// add another attestation to server
	latestCommitments1 := []models.ClientCommitment{
		models.ClientCommitment{*hashX, 0},
		models.ClientCommitment{*hashY, 1}}
	dbFake.SetClientCommitments(latestCommitments1)
	latestCommitment1, _ := models.NewCommitment([]chainhash.Hash{*hashX, *hashY})

	txid1, _ := chainhash.NewHashFromStr("21111111111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latest1 := models.NewAttestation(*txid1, latestCommitment1)
	latest1.Confirmed = true
	errUpdate = server.UpdateLatestAttestation(*latest1)
	assert.Equal(t, nil, errUpdate)

	// check commitment for new attestation
	commitment, err = server.GetAttestationCommitment(*txid1)
	assert.Equal(t, nil, err)
	assert.Equal(t, latestCommitment1.GetCommitmentHash(), commitment.GetCommitmentHash())

	commitment, err = server.GetAttestationCommitment(*txid1, true)
	assert.Equal(t, nil, err)
	assert.Equal(t, latestCommitment1.GetCommitmentHash(), commitment.GetCommitmentHash())

	commitment, err = server.GetAttestationCommitment(*txid1, false)
	assert.Equal(t, nil, err)
	assert.Equal(t, latestCommitment1.GetCommitmentHash(), commitment.GetCommitmentHash())

	// check commitment for old attestation
	commitment, err = server.GetAttestationCommitment(*txid0)
	assert.Equal(t, nil, err)
	assert.Equal(t, latestCommitment0.GetCommitmentHash(), commitment.GetCommitmentHash())

	commitment, err = server.GetAttestationCommitment(*txid0, true)
	assert.Equal(t, nil, err)
	assert.Equal(t, latestCommitment0.GetCommitmentHash(), commitment.GetCommitmentHash())

	commitment, err = server.GetAttestationCommitment(*txid0, false)
	assert.Equal(t, nil, err)
	assert.Equal(t, latestCommitment0.GetCommitmentHash(), commitment.GetCommitmentHash())

	// check commitment for invalid attestation
	commitment, err = server.GetAttestationCommitment(chainhash.Hash{})
	assert.Equal(t, nil, err)
	assert.Equal(t, chainhash.Hash{}, commitment.GetCommitmentHash())

	commitment, err = server.GetAttestationCommitment(chainhash.Hash{}, true)
	assert.Equal(t, nil, err)
	assert.Equal(t, chainhash.Hash{}, commitment.GetCommitmentHash())

	commitment, err = server.GetAttestationCommitment(chainhash.Hash{}, false)
	assert.Equal(t, errors.New(models.ErrorCommitmentListEmpty), err)
}
