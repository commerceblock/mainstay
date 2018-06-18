// Attest Server Test

package main

import (
    "testing"
    "time"
    "log"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
    "github.com/stretchr/testify/assert"
)

func TestAttestServer(t *testing.T) {
    // TEST INIT
    test := NewTest()
    genesis, _ := test.ocean.GetBlockHash(0)
    latestTx := &Attestation{chainhash.Hash{}, chainhash.Hash{}, true, time.Now()}
    server := NewAttestServer(test.ocean, *latestTx, test.tx0hash, *genesis)
    client := NewAttestClient(test.btc, test.ocean, test.tx0pk, test.tx0hash)

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

    // Test various requests
    req := Request{"BestBlock", ""}
    resp1, _ := server.Respond(req).(BestBlockResponse)
    assert.Equal(t, "", resp1.Error)
    assert.Equal(t, bestblockhash.String(), resp1.BlockHash)

    req = Request{"LatestAttestation", ""}
    resp2, _ := server.Respond(req).(LatestAttestationResponse)
    assert.Equal(t, "", resp2.Error)
    assert.Equal(t, txnew.String(), resp2.TxHash)

    req = Request{"Block", "1"}
    resp3, _ := server.Respond(req).(BlockResponse)
    assert.Equal(t, "", resp3.Error)
    assert.Equal(t, true, resp3.Attested)

    req = Request{"Block", "11"}
    resp4, _ := server.Respond(req).(BlockResponse)
    assert.Equal(t, "", resp4.Error)
    assert.Equal(t, false, resp4.Attested)

    req = Request{"WhenMoon", ""}
    resp5, _ := server.Respond(req).(Response)
    assert.Equal(t, "**AttestServer** Non supported request type", resp5.Error)

    // Test requests for the attested best block and a new generated block
    client.sideClient.Generate(1)
    bestblockhashnew, _ := client.sideClient.GetBestBlockHash()

    req = Request{"Block", bestblockhash.String()}
    resp6, _ := server.Respond(req).(BlockResponse)
    assert.Equal(t, "", resp6.Error)
    assert.Equal(t, true, resp6.Attested)

    req = Request{"Block", bestblockhashnew.String()}
    resp7, _ := server.Respond(req).(BlockResponse)
    assert.Equal(t, "", resp7.Error)
    assert.Equal(t, false, resp7.Attested)

    // Test requests for a tx in the best attested block and a tx in a newly generated block
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

    req = Request{"Transaction", txAttested}
    resp8, _ := server.Respond(req).(TransactionResponse)
    assert.Equal(t, "", resp8.Error)
    assert.Equal(t, true, resp8.Attested)

    req = Request{"Transaction", txNotAttested}
    resp9, _ := server.Respond(req).(TransactionResponse)
    assert.Equal(t, "", resp9.Error)
    assert.Equal(t, false, resp9.Attested)
}
