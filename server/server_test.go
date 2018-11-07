package server

import (
	"testing"

	"mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/assert"
)

// Test Server responses to requests
func TestServer(t *testing.T) {
	// TEST INIT
	dbFake := NewDbFake()
	server := NewServer(dbFake)

	// Generate blocks in side chain and update server latest
	latestCommitment, errCommitment := dbFake.getLatestCommitment()
	assert.Equal(t, nil, errCommitment)

	// Test latest commitment request
	respCommitment, errCommitment := server.GetLatestCommitment()
	assert.Equal(t, nil, errCommitment)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respCommitment.GetCommitmentHash())

	// Test latest attestation request
	respAttestationHash, errAttestation := server.GetLatestAttestationCommitmentHash()
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, chainhash.Hash{}, respAttestationHash)

	// Generate new attestation and update server
	txid, _ := chainhash.NewHashFromStr("11111111111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latest := models.NewAttestation(*txid, &latestCommitment)
	latest.Confirmed = true

	// Test update latest attestation
	errUpdate := server.UpdateLatestAttestation(*latest)
	assert.Equal(t, nil, errUpdate)

	respAttestationHash, errAttestation = server.GetLatestAttestationCommitmentHash()
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), respAttestationHash)
}
