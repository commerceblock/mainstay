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
    assert.Equal(t, server.latest.confirmed, true)
    assert.Equal(t, server.latest.txid, txnew)
    assert.Equal(t, server.latest.clientHash, sidehash)
    assert.Equal(t, server.latestHeight, int32(10))

    bestblockhash, _ := client.sideClient.GetBestBlockHash()

    // Test various requests
    req := Request{"BestBlock", ""}
    resp1, _ := server.Respond(req).(BestBlockResponse)
    assert.Equal(t, resp1.Error, "")
    assert.Equal(t, resp1.BlockHash, bestblockhash.String())

    req = Request{"LatestAttestation", ""}
    resp2, _ := server.Respond(req).(LatestAttestationResponse)
    assert.Equal(t, resp2.Error, "")
    assert.Equal(t, resp2.TxHash, txnew.String())

    req = Request{"Block", "1"}
    resp3, _ := server.Respond(req).(BlockResponse)
    assert.Equal(t, resp3.Error, "")
    assert.Equal(t, resp3.Attested, true)

    req = Request{"Block", "11"}
    resp4, _ := server.Respond(req).(BlockResponse)
    assert.Equal(t, resp4.Error, "")
    assert.Equal(t, resp4.Attested, false)

    req = Request{"WhenMoon", ""}
    resp5, _ := server.Respond(req).(Response)
    assert.Equal(t, resp5.Error, "**AttestServer** Non supported request type")

    // Test requests for the attested best block and a new generated block
    client.sideClient.Generate(1)
    bestblockhashnew, _ := client.sideClient.GetBestBlockHash()

    req = Request{"Block", bestblockhash.String()}
    resp6, _ := server.Respond(req).(BlockResponse)
    assert.Equal(t, resp6.Error, "")
    assert.Equal(t, resp6.Attested, true)

    req = Request{"Block", bestblockhashnew.String()}
    resp7, _ := server.Respond(req).(BlockResponse)
    assert.Equal(t, resp7.Error, "")
    assert.Equal(t, resp7.Attested, false)

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
    assert.Equal(t, resp8.Error, "")
    assert.Equal(t, resp8.Attested, true)

    req = Request{"Transaction", txNotAttested}
    resp9, _ := server.Respond(req).(TransactionResponse)
    assert.Equal(t, resp9.Error, "")
    assert.Equal(t, resp9.Attested, false)
}
