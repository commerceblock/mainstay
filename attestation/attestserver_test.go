package attestation

import (
	"log"
	"mainstay/clients"
	"mainstay/models"
	"mainstay/test"
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/assert"
)

// Attest Server Test
func TestAttestServer(t *testing.T) {
	// TEST INIT
	test := test.NewTest(false, false)
	var sideClientFake *clients.SidechainClientFake
	sideClientFake = test.GetSidechainClient().(*clients.SidechainClientFake)

	genesis, _ := sideClientFake.GetBlockHash(0)
	latestTx := &Attestation{chainhash.Hash{}, chainhash.Hash{}, true, time.Now()}
	server := NewAttestServer(sideClientFake, *latestTx, test.Tx0hash, *genesis)
	client := NewAttestClient(test.Btc, sideClientFake, test.BtcConfig, test.Tx0hash)

	// Generate blocks in side chain
	sideClientFake.Generate(10)

	// Generate single attestation transaction
	_, unspent := client.findLastUnspent()
	sidehash, addr := client.getNextAttestationAddr()
	txnew := client.sendAttestation(addr, unspent, true)
	client.mainClient.Generate(1)

	// Update latest in server
	latest := &Attestation{txnew, sidehash, true, time.Now()}
	server.UpdateLatest(*latest)
	assert.Equal(t, true, server.latest.confirmed)
	assert.Equal(t, txnew, server.latest.txid)
	assert.Equal(t, sidehash, server.latest.attestedHash)
	assert.Equal(t, int32(10), server.latestHeight)

	bestblockhash, _ := client.sideClient.GetBestBlockHash()

	// Test various models.Requests
	req := models.Request{"BestBlock", ""}
	resp1, _ := server.Respond(req).(models.BestBlockResponse)
	assert.Equal(t, "", resp1.Error)
	assert.Equal(t, bestblockhash.String(), resp1.BlockHash)

	req = models.Request{"BestBlockHeight", ""}
	resp1b, _ := server.Respond(req).(models.BestBlockHeightResponse)
	assert.Equal(t, "", resp1b.Error)
	assert.Equal(t, int32(10), resp1b.BlockHeight)

	req = models.Request{"LatestAttestation", ""}
	resp2, _ := server.Respond(req).(models.LatestAttestationResponse)
	assert.Equal(t, "", resp2.Error)
	assert.Equal(t, txnew.String(), resp2.TxHash)

	req = models.Request{"Block", "1"}
	resp3, _ := server.Respond(req).(models.BlockResponse)
	assert.Equal(t, "", resp3.Error)
	assert.Equal(t, true, resp3.Attested)

	req = models.Request{"Block", "11"}
	resp4, _ := server.Respond(req).(models.BlockResponse)
	assert.Equal(t, "", resp4.Error)
	assert.Equal(t, false, resp4.Attested)

	req = models.Request{"WhenMoon", ""}
	resp5, _ := server.Respond(req).(models.Response)
	assert.Equal(t, "**AttestServer** Non supported request type", resp5.Error)

	// Test models.Requests for the attested best block and a new generated block
	sideClientFake.Generate(1)
	bestblockhashnew, _ := client.sideClient.GetBestBlockHash()

	req = models.Request{"Block", bestblockhash.String()}
	resp6, _ := server.Respond(req).(models.BlockResponse)
	assert.Equal(t, "", resp6.Error)
	assert.Equal(t, true, resp6.Attested)

	req = models.Request{"Block", bestblockhashnew.String()}
	resp7, _ := server.Respond(req).(models.BlockResponse)
	assert.Equal(t, "", resp7.Error)
	assert.Equal(t, false, resp7.Attested)

	// Test models.Requests for a tx in the best attested block and a tx in a newly generated block
	blocktxs, err1 := sideClientFake.GetBlockTxs(bestblockhash)
	if err1 != nil {
		log.Fatal(err1)
	}
	txAttested := blocktxs[0]
	blocktxsnew, err2 := sideClientFake.GetBlockTxs(bestblockhashnew)
	if err2 != nil {
		log.Fatal(err2)
	}
	txNotAttested := blocktxsnew[0]

	req = models.Request{"Transaction", txAttested}
	resp8, _ := server.Respond(req).(models.TransactionResponse)
	assert.Equal(t, "", resp8.Error)
	assert.Equal(t, true, resp8.Attested)

	req = models.Request{"Transaction", txNotAttested}
	resp9, _ := server.Respond(req).(models.TransactionResponse)
	assert.Equal(t, "", resp9.Error)
	assert.Equal(t, false, resp9.Attested)
}
