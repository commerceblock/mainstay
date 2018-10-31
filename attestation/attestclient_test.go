package attestation

import (
	"testing"

	"mainstay/clients"
	"mainstay/models"
	"mainstay/test"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/assert"
)

// Attest Client Test for AttestClient struct and methods
// Chaining calls together because it is easier to test
// Using intermediate calls to provide data for next calls
func TestAttestClient(t *testing.T) {
	// TEST INIT
	test := test.NewTest(false, false)
	sideClientFake := test.OceanClient.(*clients.SidechainClientFake)
	client := NewAttestClient(test.Config)
	txs := []string{client.txid0}

	// Find unspent and verify is it the genesis transaction
	success, unspent, errUnspent := client.findLastUnspent()
	assert.Equal(t, true, success)
	assert.Equal(t, txs[0], unspent.TxID)
	assert.Equal(t, nil, errUnspent)

	lastHash := chainhash.Hash{}

	// Do attestations using attest client
	for i := 0; i < 10; i++ {
		// Generate attestation transaction with the unspent vout
		oceanhash, _ := sideClientFake.GetBestBlockHash()

		// test getting next attestation key
		key, errKey := client.GetNextAttestationKey(*oceanhash)
		assert.Equal(t, nil, errKey)

		// test getting next attestation address
		addr, script := client.GetNextAttestationAddr(key, *oceanhash)

		// test GetKeyAndScriptFromHash returns the same results
		keyTest, scriptTest := client.GetKeyAndScriptFromHash(*oceanhash)
		assert.Equal(t, *key, keyTest)
		assert.Equal(t, script, scriptTest)

		// test importing address
		importErr := client.ImportAttestationAddr(addr)
		assert.Equal(t, nil, importErr)

		// test creating attestation transaction
		tx, attestationErr := client.createAttestation(addr, unspent, true)
		assert.Equal(t, nil, attestationErr)
		assert.Equal(t, 1, len(tx.TxIn))
		assert.Equal(t, 1, len(tx.TxOut))
		if (unspent.Amount - (float64(tx.TxOut[0].Value) / 100000000)) <= 0 {
			t.Fail()
		}

		// test signing and sending attestation
		signedTx, signErr := client.signAttestation(tx, [][]byte{}, lastHash)
		assert.Equal(t, nil, signErr)
		txid, sendErr := client.sendAttestation(signedTx)
		assert.Equal(t, nil, sendErr)

		sideClientFake.Generate(1)
		lastHash = *oceanhash

		// Verify getUnconfirmedTx gives the unconfirmed transaction just submitted
		unconf, unconfTxid, unconfErr := client.getUnconfirmedTx() // new tx is unconfirmed
		commitment, _ := models.NewCommitment([]chainhash.Hash{lastHash})
		unconfirmed := models.NewAttestation(unconfTxid, commitment)
		assert.Equal(t, nil, unconfErr)
		assert.Equal(t, true, unconf)
		assert.Equal(t, txid, unconfirmed.Txid)
		assert.Equal(t, *oceanhash, unconfirmed.CommitmentHash())

		client.MainClient.Generate(1)

		// Verify no more unconfirmed transactions after new block generation
		unconfRe, unconfTxidRe, unconfReErr := client.getUnconfirmedTx()
		assert.Equal(t, nil, unconfReErr)
		assert.Equal(t, false, unconfRe)
		assert.Equal(t, chainhash.Hash{}, unconfTxidRe) // new tx no longer unconfirmed
		txs = append(txs, txid.String())

		// Now check that the new unspent is the vout from the transaction just submitted
		var errUnspentNew error
		success, unspent, errUnspentNew = client.findLastUnspent()
		assert.Equal(t, nil, errUnspentNew)
		assert.Equal(t, true, success)
		assert.Equal(t, txid.String(), unspent.TxID) // last unspent txnew is txnew vout
	}

	assert.Equal(t, len(txs), 11)

	for _, txid := range txs {
		// Verify transaction subchain correctness
		txhash, _ := chainhash.NewHashFromStr(txid)
		assert.Equal(t, client.verifyTxOnSubchain(*txhash), true)

		txraw, err := client.MainClient.GetRawTransaction(txhash)
		assert.Equal(t, nil, err)

		// Test attestation transactions have a single vout
		assert.Equal(t, 1, len(txraw.MsgTx().TxOut))
	}
}
