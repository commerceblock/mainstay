package attestation

import (
    "testing"

    "ocean-attestation/test"
    "ocean-attestation/clients"

    "github.com/btcsuite/btcd/chaincfg/chainhash"
    "github.com/stretchr/testify/assert"
)

// Attest Client Test
func TestAttestClient(t *testing.T) {
    // TEST INIT
    var txs []string
    test := test.NewTest(false, false)
    testConfig := test.Config
    var sideClientFake *clients.SidechainClientFake
    sideClientFake = testConfig.OceanClient().(*clients.SidechainClientFake)

    client := NewAttestClient(testConfig)
    txs = append(txs, client.txid0)

    // Find unspent and verify is it the genesis transaction
    success, unspent := client.findLastUnspent()
    if (!success) {
        t.Fail()
    }
    assert.Equal(t, txs[0], unspent.TxID)

    lastHash := chainhash.Hash{}

    // Do 10 attestations
    for i := 0; i < 10; i++ {
        // Generate attestation transaction with the unspent vout
        oceanhash := client.getNextAttestationHash()
        key := client.getNextAttestationKey(oceanhash)
        addr, _ := client.getNextAttestationAddr(key, oceanhash)

        tx := client.createAttestation(addr, unspent, true)
        txid := client.signAndSendAttestation(tx, unspent, []string{}, lastHash)
        sideClientFake.Generate(1)

        lastHash = oceanhash

        // Verify getUnconfirmedTx gives the unconfirmed transaction just submitted
        var unconfirmed *Attestation = &Attestation{}
        unconf, unconftx := client.getUnconfirmedTx()    // new tx is unconfirmed
        *unconfirmed = unconftx
        assert.Equal(t, true, unconf)
        assert.Equal(t, txid, unconfirmed.txid)
        assert.Equal(t, oceanhash, unconfirmed.attestedHash)

        // Verify no more unconfirmed transactions after new block generation
        client.mainClient.Generate(1)
        unconfRe, unconftxRe := client.getUnconfirmedTx()
        *unconfirmed = unconftxRe
        assert.Equal(t, false, unconfRe)
        assert.Equal(t, chainhash.Hash{}, unconfirmed.txid) // new tx no longer unconfirmed
        assert.Equal(t, chainhash.Hash{}, unconfirmed.attestedHash)
        txs = append(txs, txid.String())

        // Now check that the new unspent is the vout from the transaction just submitted
        success, unspent = client.findLastUnspent()
        if (!success) {
            t.Fail()
        }
        assert.Equal(t, txid.String(), unspent.TxID) // last unspent txnew is txnew vout
    }

    assert.Equal(t, len(txs), 11)

    for _, txid := range txs {
        // Verify transaction subchain correctness
        txhash, _ := chainhash.NewHashFromStr(txid)
        assert.Equal(t, client.verifyTxOnSubchain(*txhash), true)

        txraw, err := client.mainClient.GetRawTransaction(txhash)
        if err != nil {
            t.Fail()
        }
        // Test attestation transactions have a single vout
        assert.Equal(t, 1, len(txraw.MsgTx().TxOut))
    }
}
