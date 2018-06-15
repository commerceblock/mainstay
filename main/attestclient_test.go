// Attest Client Test

package main

import (
    "testing"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
    "github.com/stretchr/testify/assert"
)

func TestAttestClient(t *testing.T) {
    // TEST INIT
    var txs []string
    var oceanhashes []chainhash.Hash
    test := NewTest()
    client := NewAttestClient(test.btc, test.ocean, test.tx0pk, test.tx0hash)
    txs = append(txs, client.txid0)

    // Find unspent and verify is it the genesis transaction
    success, unspent := client.findLastUnspent()
    if (!success) {
        t.Fail()
    }
    assert.Equal(t, unspent.TxID, txs[0])

    // Do 10 attestations
    for i := 0; i < 10; i++ {
        // Generate attestation transaction with the unspent vout
        oceanhash, addr := client.getNextAttestationAddr()
        txnew := client.sendAttestation(addr, unspent)
        oceanhashes = append(oceanhashes, oceanhash)

        // Verify getUnconfirmedTx gives the unconfirmed transaction just submitted
        var unconfirmed *Attestation = &Attestation{}
        client.getUnconfirmedTx(unconfirmed)    // new tx is unconfirmed
        assert.Equal(t, txnew, unconfirmed.txid)

        // Verify no more unconfirmed transactions after new block generation
        client.mainClient.Generate(1)
        unconfirmed = &Attestation{}
        client.getUnconfirmedTx(unconfirmed)
        assert.Equal(t, chainhash.Hash{}, unconfirmed.txid) // new tx no longer unconfirmed
        txs = append(txs, txnew.String())

        // Now check that the new unspent is the vout from the transaction just submitted
        success, unspent = client.findLastUnspent()
        if (!success) {
            t.Fail()
        }
        assert.Equal(t, unspent.TxID, txnew.String()) // last unspent txnew is txnew vout
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
        assert.Equal(t, len(txraw.MsgTx().TxOut), 1)
    }

    // TODO
    // key generation testing
    // find ocean hash from tx testing
}
