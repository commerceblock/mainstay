// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	"encoding/hex"
	"errors"
	"math"
	"testing"

	"mainstay/clients"
	confpkg "mainstay/config"
	"mainstay/crypto"
	"mainstay/models"
	"mainstay/test"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"github.com/stretchr/testify/assert"
)

const TOPUP_LEVEL = 3 // level at which to do the topup transaction
const ITER_NUM = 5    // num of iterations

// Find transaction unspent for a specific address
func getUnspentForAddress(addr string, unspent []btcjson.ListUnspentResult) btcjson.ListUnspentResult {
	for _, u := range unspent {
		if u.Address == addr {
			return u
		}
	}
	return btcjson.ListUnspentResult{}
}

// Create topup unspent transaction
func createTopupUnspent(t *testing.T, config *confpkg.Config) chainhash.Hash {
	topupAddr, topupAddrErr := btcutil.DecodeAddress(config.TopupAddress(), config.MainChainCfg())
	assert.Equal(t, nil, topupAddrErr)
	topupTxHash, topupTxErr := config.MainClient().SendToAddress(topupAddr, 50*COIN)
	assert.Equal(t, nil, topupTxErr)
	return *topupTxHash
}

// Verify and get topup unspent transaction
func getTopUpUnspent(t *testing.T, client *AttestClient, config *confpkg.Config, topupHash chainhash.Hash) btcjson.ListUnspentResult {
	unspents, _ := client.MainClient.ListUnspent()
	topupUnspent := getUnspentForAddress(config.TopupAddress(), unspents)
	assert.Equal(t, config.TopupAddress(), topupUnspent.Address)
	assert.Equal(t, topupHash.String(), topupUnspent.TxID)
	return topupUnspent
}

// verify commitment and return
func verifyCommitment(t *testing.T, sideClientFake *clients.SidechainClientFake) *models.Commitment {
	hash, _ := sideClientFake.GetBestBlockHash()
	commitment, errCommitment := models.NewCommitment([]chainhash.Hash{*hash})
	assert.Equal(t, nil, errCommitment)

	return commitment
}

// verify first unspent and return
func verifyFirstUnspent(t *testing.T, client *AttestClient) btcjson.ListUnspentResult {
	success, unspent, errUnspent := client.findLastUnspent()
	assert.Equal(t, true, success)
	assert.Equal(t, nil, errUnspent)
	return unspent
}

// verify key derivation and return address
func verifyKeysAndAddr(t *testing.T, client *AttestClient, hash chainhash.Hash) btcutil.Address {
	// test getting next attestation key
	key, errKey := client.GetNextAttestationKey(hash)
	assert.Equal(t, nil, errKey)

	// test getting next attestation address
	addr, script := client.GetNextAttestationAddr(key, hash)

	// test GetKeyAndScriptFromHash returns the same results
	keyTest := client.GetKeyFromHash(hash)
	scriptTest := client.GetScriptFromHash(hash)
	assert.Equal(t, *key, keyTest)
	assert.Equal(t, script, scriptTest)

	// test importing address
	importErr := client.ImportAttestationAddr(addr)
	assert.Equal(t, nil, importErr)

	return addr
}

// verify unconfirmed attestation
func verifyUnconfirmed(t *testing.T, client *AttestClient, txid chainhash.Hash, commitment *models.Commitment) {
	unconf, unconfTxid, unconfErr := client.getUnconfirmedTx() // new tx is unconfirmed
	unconfirmed := models.NewAttestation(unconfTxid, commitment)
	assert.Equal(t, nil, unconfErr)
	assert.Equal(t, true, unconf)
	assert.Equal(t, txid, unconfirmed.Txid)
	assert.Equal(t, commitment.GetCommitmentHash(), unconfirmed.CommitmentHash())
}

// verify no longer unconfirmed attestation
func verifyNoUnconfirmed(t *testing.T, client *AttestClient) {
	unconfRe, unconfTxidRe, unconfReErr := client.getUnconfirmedTx()
	assert.Equal(t, nil, unconfReErr)
	assert.Equal(t, false, unconfRe)
	assert.Equal(t, chainhash.Hash{}, unconfTxidRe) // new tx no longer unconfirmed
}

// verify new unspent transaction and return
func verifyNewUnspent(t *testing.T, client *AttestClient, txid chainhash.Hash) btcjson.ListUnspentResult {
	success, unspent, errUnspentNew := client.findLastUnspent()
	assert.Equal(t, nil, errUnspentNew)
	assert.Equal(t, true, success)
	assert.Equal(t, txid.String(), unspent.TxID) // last unspent txnew is txnew vout
	return unspent
}

// verify attestation transactions
func verifyTxs(t *testing.T, client *AttestClient, txs []string) {
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
// Chaining calls together because it is easier to test
// Using intermediate calls to provide data for next calls
// Test case with single Client Signer
func TestAttestClient_Signer(t *testing.T) {
	// TEST INIT
	test := test.NewTest(false, false)
	sideClientFake := test.OceanClient.(*clients.SidechainClientFake)
	client := NewAttestClient(test.Config, true) // set isSigner flag
	txs := []string{client.txid0}

	// Find unspent and verify is it the genesis transaction
	unspent := verifyFirstUnspent(t, client)
	assert.Equal(t, txs[0], unspent.TxID)

	lastHash := chainhash.Hash{}
	topupHash := chainhash.Hash{}

	client.Fees.ResetFee(true) // reset fee to minimum

	// Do attestations using attest client
	for i := 1; i <= ITER_NUM; i++ {
		// Generate attestation transaction with the unspent vout
		oceanCommitment := verifyCommitment(t, sideClientFake)
		oceanCommitmentHash := oceanCommitment.GetCommitmentHash()

		// get addr
		addr := verifyKeysAndAddr(t, client, oceanCommitmentHash)

		var unspentList []btcjson.ListUnspentResult
		unspentList = append(unspentList, unspent)
		unspentAmount := unspent.Amount
		// add topup unspent to unspent list
		if i == TOPUP_LEVEL+1 {
			topupUnspent := getTopUpUnspent(t, client, test.Config, topupHash)
			unspentList = append(unspentList, topupUnspent)
			unspentAmount += topupUnspent.Amount
		}
		t.Logf("%v\n", unspentList)
		// test creating attestation transaction
		tx, attestationErr := client.createAttestation(addr, unspentList)
		assert.Equal(t, nil, attestationErr)
		assert.Equal(t, 1-1*int(math.Min(0, float64((i%(TOPUP_LEVEL+1)-1)))), len(tx.TxIn))
		assert.Equal(t, 1, len(tx.TxOut))
		assert.Equal(t, false, (unspentAmount-(float64(tx.TxOut[0].Value)/COIN)) <= 0)

		// check fee value and bump
		assert.Equal(t, client.Fees.minFee+(i-1)*client.Fees.feeIncrement, client.Fees.GetFee())
		client.Fees.BumpFee()

		// test signing and sending attestation
		signedTx, signErr := client.signAttestation(tx, [][]crypto.Sig{}, lastHash)
		assert.Equal(t, nil, signErr)
		txid, sendErr := client.sendAttestation(signedTx)
		assert.Equal(t, nil, sendErr)

		sideClientFake.Generate(1)
		lastHash = oceanCommitmentHash

		// Verify getUnconfirmedTx gives the unconfirmed transaction just submitted
		verifyUnconfirmed(t, client, txid, oceanCommitment)

		// create topup unspent
		if i == TOPUP_LEVEL {
			topupHash = createTopupUnspent(t, test.Config)
		}

		client.MainClient.Generate(1)

		// Verify no more unconfirmed transactions after new block generation
		verifyNoUnconfirmed(t, client)
		txs = append(txs, txid.String())

		// Now check that the new unspent is the vout from the transaction just submitted
		unspent = verifyNewUnspent(t, client, txid)
	}

	assert.Equal(t, len(txs), ITER_NUM+1)

	verifyTxs(t, client, txs)
}

// Attest Client Test for AttestClient struct and methods
// Test 2 attest clients, one signing, one not signing
func TestAttestClient_SignerAndNoSigner(t *testing.T) {
	// TEST INIT
	test := test.NewTest(false, false)
	sideClientFake := test.OceanClient.(*clients.SidechainClientFake)
	client := NewAttestClient(test.Config) // set isSigner flag
	clientSigner := NewAttestClient(test.Config, true)
	txs := []string{client.txid0}

	// Find unspent and verify is it the genesis transaction
	unspent := verifyFirstUnspent(t, client)
	assert.Equal(t, txs[0], unspent.TxID)

	lastHash := chainhash.Hash{}
	topupHash := chainhash.Hash{}

	client.Fees.ResetFee(true) // reset fee to minimum

	// Do attestations using attest client
	for i := 1; i <= ITER_NUM; i++ {
		// Generate attestation transaction with the unspent vout
		oceanCommitment := verifyCommitment(t, sideClientFake)
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

		var unspentList []btcjson.ListUnspentResult
		unspentList = append(unspentList, unspent)
		unspentAmount := unspent.Amount
		// add topup unspent to unspent list
		if i == TOPUP_LEVEL+1 {
			topupUnspent := getTopUpUnspent(t, client, test.Config, topupHash)
			unspentList = append(unspentList, topupUnspent)
			unspentAmount += topupUnspent.Amount
		}

		// test creating attestation transaction
		tx, attestationErr := client.createAttestation(addr, unspentList)
		assert.Equal(t, nil, attestationErr)
		assert.Equal(t, 1-1*int(math.Min(0, float64((i%(TOPUP_LEVEL+1)-1)))), len(tx.TxIn))
		assert.Equal(t, 1, len(tx.TxOut))
		assert.Equal(t, false, (unspentAmount-(float64(tx.TxOut[0].Value)/COIN)) <= 0)

		// check fee value and bump
		assert.Equal(t, client.Fees.minFee+(i-1)*client.Fees.feeIncrement, client.Fees.GetFee())
		client.Fees.BumpFee()

		// test signing and sending attestation
		signedTx, signErr := client.signAttestation(tx, [][]crypto.Sig{}, lastHash)
		assert.Equal(t, errors.New(ERROR_SIGS_MISSING_FOR_TX), signErr)

		txid, sendErr := client.sendAttestation(signedTx)
		assert.Equal(t, true, sendErr != nil)

		// client can't sign - we need to sign using clientSigner
		signedTxSigner, signedScriptSigner, signErrSigner := clientSigner.SignTransaction(lastHash, *tx)
		assert.Equal(t, nil, signErrSigner)
		assert.Equal(t, true, len(signedTxSigner.TxIn[0].SignatureScript) > 0)
		// extract sig
		sigs, sigScript := crypto.ParseScriptSig(signedTxSigner.TxIn[0].SignatureScript)
		assert.Equal(t, signedScriptSigner, hex.EncodeToString(sigScript))
		assert.Equal(t, 1, len(sigs))

		// test signing and sending attestation again
		signedTx, signErr = client.signAttestation(tx, [][]crypto.Sig{[]crypto.Sig{sigs[0]}}, lastHash)
		// exceptional top-up case - need to include additional unspent + signatures
		if i == TOPUP_LEVEL+1 {
			// test error for not enough sigs
			assert.Equal(t, errors.New(ERROR_SIGS_MISSING_FOR_TX), signErr)
			sigsTopup, sigScriptTopup := crypto.ParseScriptSig(signedTxSigner.TxIn[1].SignatureScript)
			assert.Equal(t, client.script0, hex.EncodeToString(sigScriptTopup))
			assert.Equal(t, 1, len(sigsTopup))

			// test error for not enough sigs
			signedTx, signErr = client.signAttestation(tx,
				[][]crypto.Sig{[]crypto.Sig{sigs[0]}, []crypto.Sig{}}, lastHash)
			assert.Equal(t, errors.New(ERROR_SIGS_MISSING_FOR_VIN), signErr)

			signedTx, signErr = client.signAttestation(tx,
				[][]crypto.Sig{[]crypto.Sig{sigs[0]}, []crypto.Sig{sigsTopup[0]}}, lastHash)
			assert.Equal(t, nil, signErr)
			assert.Equal(t, true, len(signedTx.TxIn[1].SignatureScript) > 0)

		} else {
			assert.Equal(t, nil, signErr)
		}
		assert.Equal(t, true, len(signedTx.TxIn[0].SignatureScript) > 0)

		txid, sendErr = client.sendAttestation(signedTx)
		assert.Equal(t, nil, sendErr)

		sideClientFake.Generate(1)
		lastHash = oceanCommitmentHash

		// Verify getUnconfirmedTx gives the unconfirmed transaction just submitted
		verifyUnconfirmed(t, client, txid, oceanCommitment)

		// create topup unspent
		if i == TOPUP_LEVEL {
			topupHash = createTopupUnspent(t, test.Config)
		}

		client.MainClient.Generate(1)

		// Verify no more unconfirmed transactions after new block generation
		verifyNoUnconfirmed(t, client)
		txs = append(txs, txid.String())

		// Now check that the new unspent is the vout from the transaction just submitted
		unspent = verifyNewUnspent(t, client, txid)
	}

	assert.Equal(t, len(txs), ITER_NUM+1)

	verifyTxs(t, client, txs)
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
	unspent := verifyFirstUnspent(t, client)
	assert.Equal(t, txs[0], unspent.TxID)

	lastHash := chainhash.Hash{}

	client.Fees.ResetFee(true) // reset fee to minimum

	// Do attestations using attest client
	for i := 1; i <= ITER_NUM; i++ {
		// Generate attestation transaction with the unspent vout
		oceanCommitment := verifyCommitment(t, sideClientFake)
		oceanCommitmentHash := oceanCommitment.GetCommitmentHash()

		// get address
		addr := verifyKeysAndAddr(t, client, oceanCommitmentHash)

		// test creating attestation transaction
		tx, attestationErr := client.createAttestation(addr, []btcjson.ListUnspentResult{unspent})
		assert.Equal(t, nil, attestationErr)
		assert.Equal(t, 1, len(tx.TxIn))
		assert.Equal(t, 1, len(tx.TxOut))
		if (unspent.Amount - (float64(tx.TxOut[0].Value) / COIN)) <= 0 {
			t.Fail()
		}
		currentValue := tx.TxOut[0].Value
		currentFee := client.Fees.GetFee()

		// test signing and sending attestation
		signedTx, signErr := client.signAttestation(tx, [][]crypto.Sig{}, lastHash)
		assert.Equal(t, nil, signErr)
		txid, sendErr := client.sendAttestation(signedTx)
		assert.Equal(t, nil, sendErr)

		// test fees too high
		prevMaxFee := client.Fees.maxFee
		client.Fees.maxFee = 999999999999
		tx, attestationErr = client.createAttestation(addr, []btcjson.ListUnspentResult{unspent})
		assert.Equal(t, errors.New(ERROR_INSUFFICIENT_FUNDS), attestationErr)
		client.Fees.maxFee = prevMaxFee
		tx, attestationErr = client.createAttestation(addr, []btcjson.ListUnspentResult{unspent})
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
		signedTx, signErr = client.signAttestation(tx, [][]crypto.Sig{}, lastHash)
		assert.Equal(t, nil, signErr)
		txid, sendErr = client.sendAttestation(signedTx)
		assert.Equal(t, nil, sendErr)

		sideClientFake.Generate(1)
		lastHash = oceanCommitmentHash

		// Verify getUnconfirmedTx gives the unconfirmed transaction just submitted
		verifyUnconfirmed(t, client, txid, oceanCommitment)

		client.MainClient.Generate(1)

		// Verify no more unconfirmed transactions after new block generation
		verifyNoUnconfirmed(t, client)
		txs = append(txs, txid.String())

		client.Fees.ResetFee(true) // reset fees again

		// Now check that the new unspent is the vout from the transaction just submitted
		unspent = verifyNewUnspent(t, client, txid)
	}

	assert.Equal(t, len(txs), ITER_NUM+1)
	verifyTxs(t, client, txs)
}
