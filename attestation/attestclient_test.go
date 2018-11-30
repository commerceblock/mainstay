// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	"encoding/hex"
	"errors"
	"testing"

	"mainstay/clients"
	"mainstay/crypto"
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
	client := NewAttestClient(test.Config, true) // set isSigner flag
	txs := []string{client.txid0}

	// Find unspent and verify is it the genesis transaction
	success, unspent, errUnspent := client.findLastUnspent()
	assert.Equal(t, true, success)
	assert.Equal(t, txs[0], unspent.TxID)
	assert.Equal(t, nil, errUnspent)

	lastHash := chainhash.Hash{}

	client.Fees.ResetFee(true) // reset fee to minimum

	// Do attestations using attest client
	for i := 0; i < 10; i++ {
		// Generate attestation transaction with the unspent vout
		oceanhash, _ := sideClientFake.GetBestBlockHash()
		oceanCommitment, errCommitment := models.NewCommitment([]chainhash.Hash{*oceanhash})
		assert.Equal(t, nil, errCommitment)
		oceanCommitmentHash := oceanCommitment.GetCommitmentHash()

		// test getting next attestation key
		key, errKey := client.GetNextAttestationKey(oceanCommitmentHash)
		assert.Equal(t, nil, errKey)

		// test getting next attestation address
		addr, script := client.GetNextAttestationAddr(key, oceanCommitmentHash)

		// test GetKeyAndScriptFromHash returns the same results
		keyTest := client.GetKeyFromHash(oceanCommitmentHash)
		scriptTest := client.GetScriptFromHash(oceanCommitmentHash)
		assert.Equal(t, *key, keyTest)
		assert.Equal(t, script, scriptTest)

		// test importing address
		importErr := client.ImportAttestationAddr(addr)
		assert.Equal(t, nil, importErr)

		// test creating attestation transaction
		tx, attestationErr := client.createAttestation(addr, unspent)
		assert.Equal(t, nil, attestationErr)
		assert.Equal(t, 1, len(tx.TxIn))
		assert.Equal(t, 1, len(tx.TxOut))
		if (unspent.Amount - (float64(tx.TxOut[0].Value) / COIN)) <= 0 {
			t.Fail()
		}

		// check fee value and bump
		assert.Equal(t, client.Fees.minFee+i*client.Fees.feeIncrement, client.Fees.GetFee())
		client.Fees.BumpFee()

		// test signing and sending attestation
		signedTx, signErr := client.signAttestation(tx, []crypto.Sig{}, lastHash)
		assert.Equal(t, nil, signErr)
		txid, sendErr := client.sendAttestation(signedTx)
		assert.Equal(t, nil, sendErr)

		sideClientFake.Generate(1)
		lastHash = oceanCommitmentHash

		// Verify getUnconfirmedTx gives the unconfirmed transaction just submitted
		unconf, unconfTxid, unconfErr := client.getUnconfirmedTx() // new tx is unconfirmed
		unconfirmed := models.NewAttestation(unconfTxid, oceanCommitment)
		assert.Equal(t, nil, unconfErr)
		assert.Equal(t, true, unconf)
		assert.Equal(t, txid, unconfirmed.Txid)
		assert.Equal(t, oceanCommitmentHash, unconfirmed.CommitmentHash())

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

// Attest Client Test for AttestClient struct and methods
// Test 2 attest clients, one signing, one not signing
func TestAttestClient_WithNoSigner(t *testing.T) {
	// TEST INIT
	test := test.NewTest(false, false)
	sideClientFake := test.OceanClient.(*clients.SidechainClientFake)
	client := NewAttestClient(test.Config) // set isSigner flag
	clientSigner := NewAttestClient(test.Config, true)
	txs := []string{client.txid0}

	// Find unspent and verify is it the genesis transaction
	success, unspent, errUnspent := client.findLastUnspent()
	assert.Equal(t, true, success)
	assert.Equal(t, txs[0], unspent.TxID)
	assert.Equal(t, nil, errUnspent)

	lastHash := chainhash.Hash{}

	client.Fees.ResetFee(true) // reset fee to minimum

	// Do attestations using attest client
	for i := 0; i < 10; i++ {
		// Generate attestation transaction with the unspent vout
		oceanhash, _ := sideClientFake.GetBestBlockHash()
		oceanCommitment, errCommitment := models.NewCommitment([]chainhash.Hash{*oceanhash})
		assert.Equal(t, nil, errCommitment)
		oceanCommitmentHash := oceanCommitment.GetCommitmentHash()

		// test getting next attestation key
		key, errKey := client.GetNextAttestationKey(oceanCommitmentHash)
		assert.Equal(t, nil, errKey)
		assert.Equal(t, true, key == nil)

		// test getting next attestation address
		addr, script := client.GetNextAttestationAddr(key, oceanCommitmentHash)

		// test GetKeyAndScriptFromHash returns the same results
		// skip testing this - not applicable in no signer case
		//keyTest := clientSigner.GetKeyFromHash(oceanCommitmentHash)
		scriptTest := client.GetScriptFromHash(oceanCommitmentHash)
		assert.Equal(t, script, scriptTest)

		// test importing address
		importErr := client.ImportAttestationAddr(addr)
		assert.Equal(t, nil, importErr)

		// test creating attestation transaction
		tx, attestationErr := client.createAttestation(addr, unspent)
		assert.Equal(t, nil, attestationErr)
		assert.Equal(t, 1, len(tx.TxIn))
		assert.Equal(t, 1, len(tx.TxOut))
		if (unspent.Amount - (float64(tx.TxOut[0].Value) / COIN)) <= 0 {
			t.Fail()
		}

		// check fee value and bump
		assert.Equal(t, client.Fees.minFee+i*client.Fees.feeIncrement, client.Fees.GetFee())
		client.Fees.BumpFee()

		// test signing and sending attestation
		signedTx, signErr := client.signAttestation(tx, []crypto.Sig{}, lastHash)
		assert.Equal(t, nil, signErr)
		assert.Equal(t, 0, len(signedTx.TxIn[0].SignatureScript))

		txid, sendErr := client.sendAttestation(signedTx)
		assert.Equal(t, true, sendErr != nil)

		// client can't sign - we need to sign using clientSigner
		signedTxSigner, signedScriptSigner, signErrSigner := clientSigner.SignTransaction(lastHash, *tx)
		assert.Equal(t, nil, signErrSigner)
		assert.Equal(t, true, len(signedTxSigner.TxIn[0].SignatureScript) > 0)
		// extract sig
		sigsSigner, scriptSigner := crypto.ParseScriptSig(signedTxSigner.TxIn[0].SignatureScript)
		assert.Equal(t, signedScriptSigner, hex.EncodeToString(scriptSigner))
		assert.Equal(t, 1, len(sigsSigner))

		// test signing and sending attestation again
		signedTx, signErr = client.signAttestation(tx, []crypto.Sig{sigsSigner[0]}, lastHash)
		assert.Equal(t, nil, signErr)
		assert.Equal(t, true, len(signedTx.TxIn[0].SignatureScript) > 0)

		txid, sendErr = client.sendAttestation(signedTx)
		assert.Equal(t, nil, sendErr)

		sideClientFake.Generate(1)
		lastHash = oceanCommitmentHash

		// Verify getUnconfirmedTx gives the unconfirmed transaction just submitted
		unconf, unconfTxid, unconfErr := client.getUnconfirmedTx() // new tx is unconfirmed
		unconfirmed := models.NewAttestation(unconfTxid, oceanCommitment)
		assert.Equal(t, nil, unconfErr)
		assert.Equal(t, true, unconf)
		assert.Equal(t, txid, unconfirmed.Txid)
		assert.Equal(t, oceanCommitmentHash, unconfirmed.CommitmentHash())

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

// Attest Client Test for AttestClient struct and methods
// Test fee bumping on existing attestation
func TestAttestClient_FeeBumping(t *testing.T) {
	// TEST INIT
	test := test.NewTest(false, false)
	sideClientFake := test.OceanClient.(*clients.SidechainClientFake)
	client := NewAttestClient(test.Config, true) // set isSigner flag
	txs := []string{client.txid0}

	// Find unspent and verify is it the genesis transaction
	success, unspent, errUnspent := client.findLastUnspent()
	assert.Equal(t, true, success)
	assert.Equal(t, txs[0], unspent.TxID)
	assert.Equal(t, nil, errUnspent)

	lastHash := chainhash.Hash{}

	client.Fees.ResetFee(true) // reset fee to minimum

	// Do attestations using attest client
	for i := 0; i < 10; i++ {
		// Generate attestation transaction with the unspent vout
		oceanhash, _ := sideClientFake.GetBestBlockHash()
		oceanCommitment, errCommitment := models.NewCommitment([]chainhash.Hash{*oceanhash})
		assert.Equal(t, nil, errCommitment)
		oceanCommitmentHash := oceanCommitment.GetCommitmentHash()

		// test getting next attestation key
		key, errKey := client.GetNextAttestationKey(oceanCommitmentHash)
		assert.Equal(t, nil, errKey)

		// test getting next attestation address
		addr, script := client.GetNextAttestationAddr(key, oceanCommitmentHash)

		// test GetKeyAndScriptFromHash returns the same results
		keyTest := client.GetKeyFromHash(oceanCommitmentHash)
		scriptTest := client.GetScriptFromHash(oceanCommitmentHash)
		assert.Equal(t, *key, keyTest)
		assert.Equal(t, script, scriptTest)

		// test importing address
		importErr := client.ImportAttestationAddr(addr)
		assert.Equal(t, nil, importErr)

		// test creating attestation transaction
		tx, attestationErr := client.createAttestation(addr, unspent)
		assert.Equal(t, nil, attestationErr)
		assert.Equal(t, 1, len(tx.TxIn))
		assert.Equal(t, 1, len(tx.TxOut))
		if (unspent.Amount - (float64(tx.TxOut[0].Value) / COIN)) <= 0 {
			t.Fail()
		}
		currentValue := tx.TxOut[0].Value
		currentFee := client.Fees.GetFee()

		// test signing and sending attestation
		signedTx, signErr := client.signAttestation(tx, []crypto.Sig{}, lastHash)
		assert.Equal(t, nil, signErr)
		txid, sendErr := client.sendAttestation(signedTx)
		assert.Equal(t, nil, sendErr)

		// test fees too high
		prevMaxFee := client.Fees.maxFee
		client.Fees.maxFee = 999999999999
		tx, attestationErr = client.createAttestation(addr, unspent)
		assert.Equal(t, errors.New(ERROR_INSUFFICIENT_FUNDS), attestationErr)
		client.Fees.maxFee = prevMaxFee
		tx, attestationErr = client.createAttestation(addr, unspent)
		assert.Equal(t, nil, attestationErr)

		// test attestation transaction fee bumping
		bumpErr := client.bumpAttestationFees(tx)
		assert.Equal(t, nil, bumpErr)
		assert.Equal(t, 1, len(tx.TxIn))
		assert.Equal(t, 1, len(tx.TxOut))
		if (unspent.Amount - (float64(tx.TxOut[0].Value) / COIN)) <= 0 {
			t.Fail()
		}
		newFee := client.Fees.GetFee()
		newValue := tx.TxOut[0].Value
		assert.Equal(t, int64((newFee-currentFee)*tx.SerializeSize()), (currentValue - newValue))
		assert.Equal(t, client.Fees.minFee+client.Fees.feeIncrement, newFee)

		// test signing and sending attestation again
		signedTx, signErr = client.signAttestation(tx, []crypto.Sig{}, lastHash)
		assert.Equal(t, nil, signErr)
		txid, sendErr = client.sendAttestation(signedTx)
		assert.Equal(t, nil, sendErr)

		sideClientFake.Generate(1)
		lastHash = oceanCommitmentHash

		// Verify getUnconfirmedTx gives the unconfirmed transaction just submitted
		unconf, unconfTxid, unconfErr := client.getUnconfirmedTx() // new tx is unconfirmed
		unconfirmed := models.NewAttestation(unconfTxid, oceanCommitment)
		assert.Equal(t, nil, unconfErr)
		assert.Equal(t, true, unconf)
		assert.Equal(t, txid, unconfirmed.Txid)
		assert.Equal(t, oceanCommitmentHash, unconfirmed.CommitmentHash())

		client.MainClient.Generate(1)

		// Verify no more unconfirmed transactions after new block generation
		unconfRe, unconfTxidRe, unconfReErr := client.getUnconfirmedTx()
		assert.Equal(t, nil, unconfReErr)
		assert.Equal(t, false, unconfRe)
		assert.Equal(t, chainhash.Hash{}, unconfTxidRe) // new tx no longer unconfirmed
		txs = append(txs, txid.String())

		client.Fees.ResetFee(true) // reset fees again

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
