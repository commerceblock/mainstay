package server

import (
	"errors"
	"testing"

	"mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/assert"
)

// Test Server UpdateLatestAttestation with no latest commitment
func TestServerUpdateLatestAttestation_NoLatestCommitments(t *testing.T) {
	// TEST INIT
	dbFake := NewDbFake()
	server := NewServer(dbFake)

	respLatestCommitment := (*models.Commitment)(nil)
	txid, _ := chainhash.NewHashFromStr("11111111111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latest := models.NewAttestation(*txid, respLatestCommitment)
	latest.Confirmed = true

	// Test update latest attestation
	errUpdate := server.UpdateLatestAttestation(*latest)
	assert.Equal(t, errors.New("Attestation Commitment not defined"), errUpdate)

}

// Test Server UpdateLatestAttestation with 1 latest commitment
func TestServerUpdateLatestAttestation_1LatestCommitments(t *testing.T) {
	// TEST INIT
	dbFake := NewDbFake()
	server := NewServer(dbFake)

	// set db latest commitment
	hash0, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitments := []models.LatestCommitment{models.LatestCommitment{*hash0, 0}}
	latestCommitment, _ := models.NewCommitment([]chainhash.Hash{*hash0})
	dbFake.setLatestCommitments(latestCommitments)

	// Test latest attestation request
	respAttestationHash, errAttestation := server.GetLatestAttestationCommitmentHash()
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, chainhash.Hash{}, respAttestationHash)

	// Generate new attestation and update server
	respLatestCommitment, err := server.GetLatestCommitment()
	assert.Equal(t, nil, err)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respLatestCommitment.GetCommitmentHash())

	txid, _ := chainhash.NewHashFromStr("11111111111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latest := models.NewAttestation(*txid, &respLatestCommitment)
	latest.Confirmed = true

	// Test update latest attestation
	errUpdate := server.UpdateLatestAttestation(*latest)
	assert.Equal(t, nil, errUpdate)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash()
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respAttestationHash)

	// Test db updated correctly
	assert.Equal(t, *txid, dbFake.attestations[0].Txid)
	assert.Equal(t, true, dbFake.attestations[0].Confirmed)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), dbFake.attestations[0].CommitmentHash())

	assert.Equal(t, latestCommitment.GetCommitmentHash(), dbFake.merkleCommitments[0].MerkleRoot)
	assert.Equal(t, int32(0), dbFake.merkleCommitments[0].ClientPosition)
	assert.Equal(t, *hash0, dbFake.merkleCommitments[0].Commitment)

	assert.Equal(t, latestCommitment.GetCommitmentHash(), dbFake.merkleProofs[0].MerkleRoot)
	assert.Equal(t, int32(0), dbFake.merkleProofs[0].ClientPosition)
	assert.Equal(t, *hash0, dbFake.merkleProofs[0].Commitment)
	assert.Equal(t, true, dbFake.merkleProofs[0].Ops[0].Append)
	assert.Equal(t, *hash0, dbFake.merkleProofs[0].Ops[0].Commitment)
}

// Test Server UpdateLatestAttestation with 3 latest commitment
func TestServerUpdateLatestAttestation_3LatestCommitments(t *testing.T) {
	// TEST INIT
	dbFake := NewDbFake()
	server := NewServer(dbFake)

	// set db latest commitment
	hash0, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash1, _ := chainhash.NewHashFromStr("baaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash01, _ := chainhash.NewHashFromStr("f6dab9f1bfb9ba3f33178e040ff78ff79bc875bfb383ce6db28f46b8226ca073")
	hash2, _ := chainhash.NewHashFromStr("caaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash22, _ := chainhash.NewHashFromStr("e0ae56a5a7eec5de827346ea45dd3d834c006d12e333d0d949aa974dda4928ed")
	latestCommitments := []models.LatestCommitment{
		models.LatestCommitment{*hash0, 0},
		models.LatestCommitment{*hash1, 1},
		models.LatestCommitment{*hash2, 2}}
	latestCommitment, _ := models.NewCommitment([]chainhash.Hash{*hash0, *hash1, *hash2})
	dbFake.setLatestCommitments(latestCommitments)

	// Test latest attestation request
	respAttestationHash, errAttestation := server.GetLatestAttestationCommitmentHash()
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, chainhash.Hash{}, respAttestationHash)

	// Generate new attestation and update server
	respLatestCommitment, err := server.GetLatestCommitment()
	assert.Equal(t, nil, err)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respLatestCommitment.GetCommitmentHash())

	txid, _ := chainhash.NewHashFromStr("11111111111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latest := models.NewAttestation(*txid, &respLatestCommitment)
	latest.Confirmed = true

	// Test update latest attestation
	errUpdate := server.UpdateLatestAttestation(*latest)
	assert.Equal(t, nil, errUpdate)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash()
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respAttestationHash)

	// Test db updated correctly
	assert.Equal(t, *txid, dbFake.attestations[0].Txid)
	assert.Equal(t, true, dbFake.attestations[0].Confirmed)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), dbFake.attestations[0].CommitmentHash())

	assert.Equal(t, latestCommitment.GetCommitmentHash(), dbFake.merkleCommitments[0].MerkleRoot)
	assert.Equal(t, int32(0), dbFake.merkleCommitments[0].ClientPosition)
	assert.Equal(t, *hash0, dbFake.merkleCommitments[0].Commitment)

	assert.Equal(t, latestCommitment.GetCommitmentHash(), dbFake.merkleCommitments[1].MerkleRoot)
	assert.Equal(t, int32(1), dbFake.merkleCommitments[1].ClientPosition)
	assert.Equal(t, *hash1, dbFake.merkleCommitments[1].Commitment)

	assert.Equal(t, latestCommitment.GetCommitmentHash(), dbFake.merkleCommitments[2].MerkleRoot)
	assert.Equal(t, int32(2), dbFake.merkleCommitments[2].ClientPosition)
	assert.Equal(t, *hash2, dbFake.merkleCommitments[2].Commitment)

	assert.Equal(t, latestCommitment.GetCommitmentHash(), dbFake.merkleProofs[0].MerkleRoot)
	assert.Equal(t, int32(0), dbFake.merkleProofs[0].ClientPosition)
	assert.Equal(t, *hash0, dbFake.merkleProofs[0].Commitment)
	assert.Equal(t, true, dbFake.merkleProofs[0].Ops[0].Append)
	assert.Equal(t, *hash1, dbFake.merkleProofs[0].Ops[0].Commitment)
	assert.Equal(t, true, dbFake.merkleProofs[0].Ops[1].Append)
	assert.Equal(t, *hash01, dbFake.merkleProofs[0].Ops[1].Commitment)

	assert.Equal(t, latestCommitment.GetCommitmentHash(), dbFake.merkleProofs[1].MerkleRoot)
	assert.Equal(t, int32(1), dbFake.merkleProofs[1].ClientPosition)
	assert.Equal(t, *hash1, dbFake.merkleProofs[1].Commitment)
	assert.Equal(t, false, dbFake.merkleProofs[1].Ops[0].Append)
	assert.Equal(t, *hash0, dbFake.merkleProofs[1].Ops[0].Commitment)
	assert.Equal(t, true, dbFake.merkleProofs[1].Ops[1].Append)
	assert.Equal(t, *hash01, dbFake.merkleProofs[1].Ops[1].Commitment)

	assert.Equal(t, latestCommitment.GetCommitmentHash(), dbFake.merkleProofs[2].MerkleRoot)
	assert.Equal(t, int32(2), dbFake.merkleProofs[2].ClientPosition)
	assert.Equal(t, *hash2, dbFake.merkleProofs[2].Commitment)
	assert.Equal(t, true, dbFake.merkleProofs[2].Ops[0].Append)
	assert.Equal(t, *hash2, dbFake.merkleProofs[2].Ops[0].Commitment)
	assert.Equal(t, false, dbFake.merkleProofs[2].Ops[1].Append)
	assert.Equal(t, *hash22, dbFake.merkleProofs[2].Ops[1].Commitment)
}

// Test Server GetLatestCommitment
func TestServerGetLatestCommitment(t *testing.T) {
	// TEST INIT
	dbFake := NewDbFake()
	server := NewServer(dbFake)

	// set db latest commitment
	hash0, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash1, _ := chainhash.NewHashFromStr("baaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hash2, _ := chainhash.NewHashFromStr("caaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")

	// update server with incorrect latest commitment and test server
	latestCommitments := []models.LatestCommitment{
		models.LatestCommitment{*hash0, 0}, models.LatestCommitment{*hash2, 2}}
	dbFake.setLatestCommitments(latestCommitments)

	respLatestCommitment, err := server.GetLatestCommitment()
	assert.Equal(t, errors.New("Latest commitment missing in position 1\n"), err)
	assert.Equal(t, chainhash.Hash{}, respLatestCommitment.GetCommitmentHash())

	// update server with incorrect latest commitment and test server
	latestCommitments = []models.LatestCommitment{
		models.LatestCommitment{*hash0, 1}, models.LatestCommitment{*hash2, 2}}
	dbFake.setLatestCommitments(latestCommitments)

	respLatestCommitment, err = server.GetLatestCommitment()
	assert.Equal(t, errors.New("Latest commitment missing in position 0\n"), err)
	assert.Equal(t, chainhash.Hash{}, respLatestCommitment.GetCommitmentHash())

	// update server with incorrect latest commitment and test server
	latestCommitments = []models.LatestCommitment{models.LatestCommitment{*hash2, 2}}
	dbFake.setLatestCommitments(latestCommitments)

	respLatestCommitment, err = server.GetLatestCommitment()
	assert.Equal(t, errors.New("Latest commitment missing in position 0\n"), err)
	assert.Equal(t, chainhash.Hash{}, respLatestCommitment.GetCommitmentHash())

	// update server with correct latest commitment and test server
	latestCommitments = []models.LatestCommitment{
		models.LatestCommitment{*hash0, 0},
		models.LatestCommitment{*hash1, 1},
		models.LatestCommitment{*hash2, 2}}
	latestCommitment, err := models.NewCommitment([]chainhash.Hash{*hash0, *hash1, *hash2})
	assert.Equal(t, nil, err)
	dbFake.setLatestCommitments(latestCommitments)

	respLatestCommitment, err = server.GetLatestCommitment()
	assert.Equal(t, nil, err)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respLatestCommitment.GetCommitmentHash())
}
