package server

import (
	"mainstay/clients"
	"mainstay/models"
	"mainstay/test"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/assert"
)

// Attest Server Test
func TestServer(t *testing.T) {
	// TEST INIT
	test := test.NewTest(false, false)
	testConfig := test.Config
	var sideClientFake *clients.SidechainClientFake
	sideClientFake = testConfig.OceanClient().(*clients.SidechainClientFake)

	server := NewServer(nil, nil, testConfig)

	// Generate blocks in side chain
	sideClientFake.Generate(10)

	// Generate single attestation transaction
	server.updateCommitment()
	sidehash := server.latestCommitment
	txid, _ := chainhash.NewHashFromStr("11111111111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")

	// Update latest in server
	latest := models.NewAttestation(*txid, sidehash)
	server.updateLatest(*latest)
	assert.Equal(t, *txid, server.latestAttestation.Txid)
	assert.Equal(t, sidehash, server.latestAttestation.AttestedHash)
	assert.Equal(t, int32(10), server.latestHeight)

	bestblockhash, _ := server.sideClient.GetBestBlockHash()

	// Test various models.Requests
	req := models.RequestGet_s{"BestBlock", ""}
	resp1, _ := server.respond(req).(models.BestBlockResponse)
	assert.Equal(t, "", resp1.Error)
	assert.Equal(t, bestblockhash.String(), resp1.BlockHash)

	req = models.RequestGet_s{"BestBlockHeight", ""}
	resp1b, _ := server.respond(req).(models.BestBlockHeightResponse)
	assert.Equal(t, "", resp1b.Error)
	assert.Equal(t, int32(10), resp1b.BlockHeight)

	req = models.RequestGet_s{"LatestAttestation", ""}
	resp2, _ := server.respond(req).(models.LatestAttestationResponse)
	assert.Equal(t, "", resp2.Error)
	assert.Equal(t, txid.String(), resp2.TxHash)

	req = models.RequestGet_s{"VerifyBlock", "1"}
	resp3, _ := server.respond(req).(models.VerifyBlockResponse)
	assert.Equal(t, "", resp3.Error)
	assert.Equal(t, true, resp3.Attested)

	req = models.RequestGet_s{"VerifyBlock", "11"}
	resp4, _ := server.respond(req).(models.VerifyBlockResponse)
	assert.Equal(t, "", resp4.Error)
	assert.Equal(t, false, resp4.Attested)

	req = models.RequestGet_s{"WhenMoon", ""}
	resp5, _ := server.respond(req).(models.Response)
	assert.Equal(t, "**Server** Non supported request type", resp5.Error)

	// Test models.Requests for the attested best block and a new generated block
	sideClientFake.Generate(1)
	server.updateCommitment()
	bestblockhashnew, _ := server.sideClient.GetBestBlockHash()
	assert.Equal(t, *bestblockhashnew, server.latestCommitment)

	req = models.RequestGet_s{"VerifyBlock", bestblockhash.String()}
	resp6, _ := server.respond(req).(models.VerifyBlockResponse)
	assert.Equal(t, "", resp6.Error)
	assert.Equal(t, true, resp6.Attested)

	req = models.RequestGet_s{"VerifyBlock", bestblockhashnew.String()}
	resp7, _ := server.respond(req).(models.VerifyBlockResponse)
	assert.Equal(t, "", resp7.Error)
	assert.Equal(t, false, resp7.Attested)
}
