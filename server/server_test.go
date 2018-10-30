package server

import (
	"mainstay/clients"
	"mainstay/models"
	"mainstay/test"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/assert"
)

// Test Server responses to requests
func TestServer(t *testing.T) {
	// TEST INIT
	test := test.NewTest(false, false)
	sideClientFake := test.OceanClient.(*clients.SidechainClientFake)
	server := NewServer(test.Config, sideClientFake)

	// Generate blocks in side chain and update server latest
	sideClientFake.Generate(10)
	bestblockhash, _ := sideClientFake.GetBestBlockHash()
	server.latestCommitment = *bestblockhash

	// Test latest commitment request
	respCommitment, errCommitment := server.GetLatestCommitment()
	assert.Equal(t, nil, errCommitment)
	assert.Equal(t, *bestblockhash, respCommitment)

	// Test latest attestation request
	respAttestation, errAttestation := server.GetLatestAttestation()
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, chainhash.Hash{}, respAttestation.Txid)
	assert.Equal(t, chainhash.Hash{}, respAttestation.AttestedHash)

	// Generate new attestation and update server
	txid, _ := chainhash.NewHashFromStr("11111111111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latest := models.NewAttestation(*txid, *bestblockhash)

	// Test update latest attestation
	errUpdate := server.UpdateLatestAttestation(*latest)
	assert.Equal(t, nil, errUpdate)
	assert.Equal(t, *txid, server.latestAttestation.Txid)
	assert.Equal(t, *bestblockhash, server.latestAttestation.AttestedHash)

	// Test latest attestation again after update
	respAttestation, errAttestation = server.GetLatestAttestation()
	assert.Equal(t, nil, errAttestation)
	assert.Equal(t, *txid, respAttestation.Txid)
	assert.Equal(t, *bestblockhash, respAttestation.AttestedHash)
}
