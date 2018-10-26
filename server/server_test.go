package server

import (
	"mainstay/clients"
	"mainstay/models"
	"mainstay/requestapi"
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

	// Test various requestapi.Requests
	req := requestapi.BaseRequest{}
	req.SetRequestType("BestBlock")
	resp1, _ := server.respond(req).(requestapi.BestBlockResponse)
	assert.Equal(t, "", resp1.ResponseError())
	assert.Equal(t, bestblockhash.String(), resp1.BlockHash)

	req.SetRequestType("BestBlockHeight")
	resp1b, _ := server.respond(req).(requestapi.BestBlockHeightResponse)
	assert.Equal(t, "", resp1b.ResponseError())
	assert.Equal(t, int32(10), resp1b.BlockHeight)

	req.SetRequestType("LatestAttestation")
	resp2, _ := server.respond(req).(requestapi.LatestAttestationResponse)
	assert.Equal(t, "", resp2.ResponseError())
	assert.Equal(t, txid.String(), resp2.TxHash)

	reqVerify := requestapi.ServerVerifyBlockRequest{}
	reqVerify.SetRequestType("VerifyBlock")
	reqVerify.Id = "1"
	resp3, _ := server.respond(reqVerify).(requestapi.VerifyBlockResponse)
	assert.Equal(t, "", resp3.ResponseError())
	assert.Equal(t, true, resp3.Attested)

	reqVerify.Id = "11"
	resp4, _ := server.respond(reqVerify).(requestapi.VerifyBlockResponse)
	assert.Equal(t, "", resp4.ResponseError())
	assert.Equal(t, false, resp4.Attested)

	req.SetRequestType("WhenMoon")
	resp5, _ := server.respond(req).(requestapi.Response)
	assert.Equal(t, "**Server** Non supported request type WhenMoon", resp5.ResponseError())

	// Test requestapi.Requests for the attested best block and a new generated block
	sideClientFake.Generate(1)
	server.updateCommitment()
	bestblockhashnew, _ := server.sideClient.GetBestBlockHash()
	assert.Equal(t, *bestblockhashnew, server.latestCommitment)

	reqVerify.Id = bestblockhash.String()
	resp6, _ := server.respond(reqVerify).(requestapi.VerifyBlockResponse)
	assert.Equal(t, "", resp6.ResponseError())
	assert.Equal(t, true, resp6.Attested)

	reqVerify.Id = bestblockhashnew.String()
	resp7, _ := server.respond(reqVerify).(requestapi.VerifyBlockResponse)
	assert.Equal(t, "", resp7.ResponseError())
	assert.Equal(t, false, resp7.Attested)
}
