// Attest Server Test

package attestation

import (
    "testing"
    "time"
    "log"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
    "github.com/stretchr/testify/assert"
    "ocean-attestation/models"
    "ocean-attestation/test"
)

func TestAttestServer(t *testing.T) {
    // TEST INIT
    test := test.NewTest()
    genesis, _ := test.Ocean.GetBlockHash(0)
    latestTx := &Attestation{chainhash.Hash{}, chainhash.Hash{}, true, time.Now()}
    server := NewAttestServer(test.Ocean, *latestTx, test.Tx0hash, *genesis)
    client := NewAttestClient(test.Btc, test.Ocean, test.BtcConfig, test.Tx0pk, test.Tx0hash)

    // Generate blocks in side chain
    client.sideClient.Generate(10)

    // Generate single attestation transaction
    _, unspent := client.findLastUnspent()
    sidehash, addr := client.getNextAttestationAddr()
    txnew := client.sendAttestation(addr, unspent)
    client.mainClient.Generate(1)

    // Update latest in server
    latest := &Attestation{txnew, sidehash, true, time.Now()}
    server.UpdateLatest(*latest)
    assert.Equal(t, true, server.latest.confirmed)
    assert.Equal(t, txnew, server.latest.txid)
    assert.Equal(t, sidehash, server.latest.clientHash)
    assert.Equal(t, int32(10), server.latestHeight)

    bestblockhash, _ := client.sideClient.GetBestBlockHash()

    // Test various models.Requests
    req := models.Request{"BestBlock", ""}
    resp1, _ := server.Respond(req).(models.BestBlockResponse)
    assert.Equal(t, "", resp1.Error)
    assert.Equal(t, bestblockhash.String(), resp1.BlockHash)

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
    client.sideClient.Generate(1)
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
    block, err1 := client.sideClient.GetBlockVerbose(bestblockhash)
    if err1 != nil {
        log.Fatal(err1)
    }
    txAttested := block.Tx[0]
    blocknew, err2 := client.sideClient.GetBlockVerbose(bestblockhashnew)
    if err2 != nil {
        log.Fatal(err2)
    }
    txNotAttested := blocknew.Tx[0]

    req = models.Request{"Transaction", txAttested}
    resp8, _ := server.Respond(req).(models.TransactionResponse)
    assert.Equal(t, "", resp8.Error)
    assert.Equal(t, true, resp8.Attested)

    req = models.Request{"Transaction", txNotAttested}
    resp9, _ := server.Respond(req).(models.TransactionResponse)
    assert.Equal(t, "", resp9.Error)
    assert.Equal(t, false, resp9.Attested)
}
