package attestation

import (
	"mainstay/clients"
	"mainstay/models"
	"mainstay/test"
	"testing"

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
	if !success {
		t.Fail()
	}
	assert.Equal(t, txs[0], unspent.TxID)

	lastHash := chainhash.Hash{}
	// Do 10 attestations
	for i := 0; i < 10; i++ {
		// Generate attestation transaction with the unspent vout
		oceanhash, _ := sideClientFake.GetBestBlockHash()
		key := client.GetNextAttestationKey(*oceanhash)
		addr, _ := client.GetNextAttestationAddr(key, *oceanhash)

		tx := client.createAttestation(addr, unspent, true)
		txid := client.signAndSendAttestation(tx, [][]byte{}, lastHash)
		sideClientFake.Generate(1)

		lastHash = *oceanhash

		// Verify getUnconfirmedTx gives the unconfirmed transaction just submitted
		var unconfirmed *models.Attestation = &models.Attestation{}
		unconf, unconfTxid := client.getUnconfirmedTx() // new tx is unconfirmed
		unconfirmed = models.NewAttestation(unconfTxid, client.getTxAttestedHash(unconfTxid), models.ASTATE_UNCONFIRMED)
		assert.Equal(t, true, unconf)
		assert.Equal(t, txid, unconfirmed.Txid)
		assert.Equal(t, *oceanhash, unconfirmed.AttestedHash)

		// Verify no more unconfirmed transactions after new block generation
		client.MainClient.Generate(1)
		unconfRe, unconfTxidRe := client.getUnconfirmedTx()
		assert.Equal(t, false, unconfRe)
		assert.Equal(t, chainhash.Hash{}, unconfTxidRe) // new tx no longer unconfirmed
		txs = append(txs, txid.String())

		// Now check that the new unspent is the vout from the transaction just submitted
		success, unspent = client.findLastUnspent()
		if !success {
			t.Fail()
		}
		assert.Equal(t, txid.String(), unspent.TxID) // last unspent txnew is txnew vout
	}

	assert.Equal(t, len(txs), 11)

	for _, txid := range txs {
		// Verify transaction subchain correctness
		txhash, _ := chainhash.NewHashFromStr(txid)
		assert.Equal(t, client.verifyTxOnSubchain(*txhash), true)

		txraw, err := client.MainClient.GetRawTransaction(txhash)
		if err != nil {
			t.Fail()
		}
		// Test attestation transactions have a single vout
		assert.Equal(t, 1, len(txraw.MsgTx().TxOut))
	}
}
