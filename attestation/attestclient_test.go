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
	testpkg "mainstay/test"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/stretchr/testify/assert"
)

const topupLevel = 3 // level at which to do the topup transaction
const iterNum = 5    // num of iterations

// Create topup unspent transaction
func createTopupUnspent(t *testing.T, config *confpkg.Config) chainhash.Hash {
	topupAddr, topupAddrErr := btcutil.DecodeAddress(config.TopupAddress(), config.MainChainCfg())
	assert.Equal(t, nil, topupAddrErr)
	topupTxHash, topupTxErr := config.MainClient().SendToAddress(topupAddr, 50*Coin)
	assert.Equal(t, nil, topupTxErr)
	return *topupTxHash
}

// Verify and get topup unspent transaction
func getTopUpUnspent(t *testing.T, client *AttestClient, config *confpkg.Config, topupHash chainhash.Hash) btcjson.ListUnspentResult {
	success, unspent, errUnspent := client.findTopupUnspent()
	assert.Equal(t, true, success)
	assert.Equal(t, nil, errUnspent)
	assert.Equal(t, config.TopupAddress(), unspent.Address)
	assert.Equal(t, topupHash.String(), unspent.TxID)
	return unspent
}

// verify commitment and return
func verifyCommitment(t *testing.T, sideClientFake *clients.SidechainClientFake) *models.Commitment {
	hash, _ := sideClientFake.GetBestBlockHash()
	commitment, errCommitment := models.NewCommitment([]chainhash.Hash{*hash})
	assert.Equal(t, nil, errCommitment)

	return commitment
}

// verify no topup unspent and first unspent and return
func verifyFirstUnspent(t *testing.T, client *AttestClient) btcjson.ListUnspentResult {
	// check no topup-unspent
	success, unspent, errUnspent := client.findTopupUnspent()
	assert.Equal(t, false, success)
	assert.Equal(t, nil, errUnspent)
	assert.Equal(t, btcjson.ListUnspentResult{}, unspent)

	// check staychain unspent exists
	success, unspent, errUnspent = client.findLastUnspent()
	assert.Equal(t, true, success)
	assert.Equal(t, nil, errUnspent)
	return unspent
}

// verify key derivation and return address
func verifyKeysAndAddr(t *testing.T, client *AttestClient, hash chainhash.Hash) (btcutil.Address, string) {
	// test getting next attestation key
	key, errKey := client.GetNextAttestationKey(hash)
	assert.Equal(t, nil, errKey)

	// test getting next attestation address
	addr, script, nextAddrErr := client.GetNextAttestationAddr(key, hash)
	assert.Equal(t, nil, nextAddrErr)

	// test GetKeyAndScriptFromHash returns the same results
	keyTest := client.GetKeyFromHash(hash)
	scriptTest, scriptErr := client.GetScriptFromHash(hash)
	assert.Equal(t, nil, scriptErr)
	assert.Equal(t, *key, keyTest)
	assert.Equal(t, script, scriptTest)

	// test importing address
	importErr := client.ImportAttestationAddr(addr)
	assert.Equal(t, nil, importErr)

	return addr, script
}

// verify transaction pre image generation
func verifyTransactionPreImages(t *testing.T, client *AttestClient, tx *wire.MsgTx, script string, hash chainhash.Hash, i int) {

	// getTransactionPreImages with empty transaction
	_, emptyPreImageErr := client.getTransactionPreImages(hash, &(wire.MsgTx{}))
	assert.Equal(t, errors.New(ErrorInputMissingForTx), emptyPreImageErr)

	// getTransactionPreImages with actual transaction
	txPreImages, preImageErr := client.getTransactionPreImages(hash, tx)
	assert.Equal(t, nil, preImageErr)
	assert.Equal(t, 1-1*int(math.Min(0, float64((i%(topupLevel+1)-1)))), len(txPreImages))
	assert.Equal(t, 1-1*int(math.Min(0, float64((i%(topupLevel+1)-1)))), len(txPreImages[0].TxIn))

	// get tweaked script and topup script serialisation
	scriptSer, _ := hex.DecodeString(script)
	topupScriptSer, _ := hex.DecodeString(client.scriptTopup)

	// test signature script set correctly
	assert.Equal(t, scriptSer, txPreImages[0].TxIn[0].SignatureScript)
	if i == topupLevel+1 {
		assert.Equal(t, []byte(nil), txPreImages[0].TxIn[1].SignatureScript)
		assert.Equal(t, []byte(nil), txPreImages[1].TxIn[0].SignatureScript)
		assert.Equal(t, topupScriptSer, txPreImages[1].TxIn[1].SignatureScript)
	}
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

// verify if there is a topup unspent or not
func verifyTopup(t *testing.T, client *AttestClient, i int) {
	// check topup unspent only when iteration is topup
	success, _, errUnspent := client.findTopupUnspent()
	assert.Equal(t, i == topupLevel, success)
	assert.Equal(t, nil, errUnspent)
}

// verify new unspent transaction and return
func verifyNewUnspent(t *testing.T, client *AttestClient, txid chainhash.Hash) btcjson.ListUnspentResult {
	// check regular unspent cycle
	success, unspent, errUnspent := client.findLastUnspent()
	assert.Equal(t, nil, errUnspent)
	assert.Equal(t, true, success)
	assert.Equal(t, txid.String(), unspent.TxID) // last unspent txnew is txnew vout
	return unspent
}

// verify attestation transactions
func verifyTxs(t *testing.T, client *AttestClient, txs []string) {
	// verify random tx hashes
	fakehash1, _ := chainhash.NewHashFromStr("1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	fakehash2, _ := chainhash.NewHashFromStr("2a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	fakehash3, _ := chainhash.NewHashFromStr("3a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	assert.Equal(t, client.verifyTxOnSubchain(*fakehash1), false)
	assert.Equal(t, client.verifyTxOnSubchain(*fakehash2), false)
	assert.Equal(t, client.verifyTxOnSubchain(*fakehash3), false)

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
	test := testpkg.NewTest(false, false)
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
	for i := 1; i <= iterNum; i++ {
		// Generate attestation transaction with the unspent vout
		oceanCommitment := verifyCommitment(t, sideClientFake)
		oceanCommitmentHash := oceanCommitment.GetCommitmentHash()

		// get addr
		addr, script := verifyKeysAndAddr(t, client, oceanCommitmentHash)

		var unspentList []btcjson.ListUnspentResult
		unspentList = append(unspentList, unspent)
		unspentAmount := unspent.Amount
		// add topup unspent to unspent list
		if i == topupLevel+1 {
			topupUnspent := getTopUpUnspent(t, client, test.Config, topupHash)
			unspentList = append(unspentList, topupUnspent)
			unspentAmount += topupUnspent.Amount
		}

		// test creating attestation transaction
		tx, attestationErr := client.createAttestation(addr, unspentList)
		assert.Equal(t, nil, attestationErr)
		assert.Equal(t, 1-1*int(math.Min(0, float64((i%(topupLevel+1)-1)))), len(tx.TxIn))
		assert.Equal(t, 1, len(tx.TxOut))
		assert.Equal(t, false, (unspentAmount-(float64(tx.TxOut[0].Value)/Coin)) <= 0)

		// verify transaction pre-image generation
		verifyTransactionPreImages(t, client, tx, script, oceanCommitmentHash, i)

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
		if i == topupLevel {
			topupHash = createTopupUnspent(t, test.Config)
		}

		client.MainClient.Generate(1)

		// Verify no more unconfirmed transactions after new block generation
		verifyNoUnconfirmed(t, client)
		txs = append(txs, txid.String())

		// Now check that the new unspent is the vout from the transaction just submitted
		unspent = verifyNewUnspent(t, client, txid)
		// check whether there is a topup unspent or not
		verifyTopup(t, client, i)
	}

	assert.Equal(t, len(txs), iterNum+1)

	verifyTxs(t, client, txs)
}

// Attest Client Test for AttestClient struct and methods
// Test 2 attest clients, one signing, one not signing
func TestAttestClient_SignerAndNoSigner(t *testing.T) {
	// TEST INIT
	test := testpkg.NewTest(false, false)
	sideClientFake := test.OceanClient.(*clients.SidechainClientFake)
	client := NewAttestClient(test.Config) // set isSigner flag
	clientSigner := NewAttestClient(test.Config, true)
	txs := []string{client.txid0}

	// Find unspent and verify is it the genesis transaction
	unspent := verifyFirstUnspent(t, client)
	assert.Equal(t, txs[0], unspent.TxID)

	lastHash := chainhash.Hash{}
	topupHash := chainhash.Hash{}

	// test that when attempting to generate a new address with
	// an empty hash, the intial address/script are returned
	addrNoHash, scriptNoHash, nextAddrErr := client.GetNextAttestationAddr(nil, lastHash)
	assert.Equal(t, nil, nextAddrErr)
	assert.Equal(t, testpkg.Address, addrNoHash.String())
	assert.Equal(t, testpkg.Script, scriptNoHash)
	assert.Equal(t, client.script0, scriptNoHash)
	assert.Equal(t, unspent.Address, addrNoHash.String())
	addrNoHash, scriptNoHash, nextAddrErr = clientSigner.GetNextAttestationAddr(nil, lastHash)
	assert.Equal(t, nil, nextAddrErr)
	assert.Equal(t, testpkg.Address, addrNoHash.String())
	assert.Equal(t, testpkg.Script, scriptNoHash)
	assert.Equal(t, clientSigner.script0, scriptNoHash)
	assert.Equal(t, unspent.Address, addrNoHash.String())

	// test invalid tx signing
	_, _, errSign := clientSigner.SignTransaction(chainhash.Hash{}, wire.MsgTx{})
	assert.Equal(t, errSign, errors.New(ErrorInputMissingForTx))

	client.Fees.ResetFee(true) // reset fee to minimum

	// Do attestations using attest client
	for i := 1; i <= iterNum; i++ {
		// Generate attestation transaction with the unspent vout
		oceanCommitment := verifyCommitment(t, sideClientFake)
		oceanCommitmentHash := oceanCommitment.GetCommitmentHash()

		// test getting next attestation key
		key, errKey := client.GetNextAttestationKey(oceanCommitmentHash)
		assert.Equal(t, nil, errKey)
		assert.Equal(t, true, key == nil)

		// test getting next attestation address
		addr, script, nextAddrErr := client.GetNextAttestationAddr(key, oceanCommitmentHash)
		assert.Equal(t, nil, nextAddrErr)

		// test GetKeyAndScriptFromHash returns the same results
		// skip testing this - not applicable in no signer case
		//keyTest := clientSigner.GetKeyFromHash(oceanCommitmentHash)
		scriptTest, scriptErr := client.GetScriptFromHash(oceanCommitmentHash)
		assert.Equal(t, nil, scriptErr)
		assert.Equal(t, script, scriptTest)

		// test importing address
		importErr := client.ImportAttestationAddr(addr, false)
		assert.Equal(t, nil, importErr)

		var unspentList []btcjson.ListUnspentResult
		unspentList = append(unspentList, unspent)
		unspentAmount := unspent.Amount
		// add topup unspent to unspent list
		if i == topupLevel+1 {
			topupUnspent := getTopUpUnspent(t, client, test.Config, topupHash)
			unspentList = append(unspentList, topupUnspent)
			unspentAmount += topupUnspent.Amount
		}

		// test creating attestation transaction
		tx, attestationErr := client.createAttestation(addr, unspentList)
		assert.Equal(t, nil, attestationErr)
		assert.Equal(t, 1-1*int(math.Min(0, float64((i%(topupLevel+1)-1)))), len(tx.TxIn))
		assert.Equal(t, 1, len(tx.TxOut))
		assert.Equal(t, false, (unspentAmount-(float64(tx.TxOut[0].Value)/Coin)) <= 0)

		// verify transaction pre-image generation
		verifyTransactionPreImages(t, client, tx, script, oceanCommitmentHash, i)

		// check fee value and bump
		assert.Equal(t, client.Fees.minFee+(i-1)*client.Fees.feeIncrement, client.Fees.GetFee())
		client.Fees.BumpFee()

		// test signing and sending attestation
		signedTx, signErr := client.signAttestation(tx, [][]crypto.Sig{}, lastHash)
		assert.Equal(t, errors.New(ErrorSigsMissingForTx), signErr)

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
		if i == topupLevel+1 {
			// test error for not enough sigs
			assert.Equal(t, errors.New(ErrorSigsMissingForTx), signErr)
			sigsTopup, sigScriptTopup := crypto.ParseScriptSig(signedTxSigner.TxIn[1].SignatureScript)
			assert.Equal(t, client.scriptTopup, hex.EncodeToString(sigScriptTopup))
			assert.Equal(t, 1, len(sigsTopup))

			// test error for not enough sigs
			signedTx, signErr = client.signAttestation(tx,
				[][]crypto.Sig{[]crypto.Sig{sigs[0]}, []crypto.Sig{}}, lastHash)
			assert.Equal(t, errors.New(ErrorSigsMissingForVin), signErr)

			signedTx, signErr = client.signAttestation(tx,
				[][]crypto.Sig{[]crypto.Sig{sigs[0]}}, lastHash)
			assert.Equal(t, errors.New(ErrorSigsMissingForTx), signErr)

			signedTx, signErr = client.signAttestation(tx,
				[][]crypto.Sig{}, lastHash)
			assert.Equal(t, errors.New(ErrorSigsMissingForTx), signErr)

			// actually sign attestation transaction
			signedTx, signErr = client.signAttestation(tx,
				[][]crypto.Sig{[]crypto.Sig{sigs[0]}, []crypto.Sig{sigsTopup[0]}}, lastHash)
			assert.Equal(t, nil, signErr)
			assert.Equal(t, true, len(signedTx.TxIn[1].SignatureScript) > 0)

			// adding more signatures than required has the same result
			signedTx, signErr = client.signAttestation(tx,
				[][]crypto.Sig{[]crypto.Sig{sigs[0], sigs[0]}, []crypto.Sig{sigsTopup[0], sigs[0], sigs[0]}}, lastHash)
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
		if i == topupLevel {
			topupHash = createTopupUnspent(t, test.Config)
		}

		client.MainClient.Generate(1)

		// Verify no more unconfirmed transactions after new block generation
		verifyNoUnconfirmed(t, client)
		txs = append(txs, txid.String())

		// Now check that the new unspent is the vout from the transaction just submitted
		unspent = verifyNewUnspent(t, client, txid)
		// check whether there is a topup unspent or not
		verifyTopup(t, client, i)
	}

	assert.Equal(t, len(txs), iterNum+1)

	verifyTxs(t, client, txs)
}

// Attest Client Test for AttestClient struct and methods
// Test fee bumping on existing attestation
func TestAttestClient_FeeBumping(t *testing.T) {
	// TEST INIT
	test := testpkg.NewTest(false, false)
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
	for i := 1; i <= iterNum; i++ {
		// Generate attestation transaction with the unspent vout
		oceanCommitment := verifyCommitment(t, sideClientFake)
		oceanCommitmentHash := oceanCommitment.GetCommitmentHash()

		// get address
		addr, script := verifyKeysAndAddr(t, client, oceanCommitmentHash)

		// test creating attestation transaction
		tx, attestationErr := client.createAttestation(addr, []btcjson.ListUnspentResult{unspent})
		assert.Equal(t, nil, attestationErr)
		assert.Equal(t, 1, len(tx.TxIn))
		assert.Equal(t, 1, len(tx.TxOut))
		if (unspent.Amount - (float64(tx.TxOut[0].Value) / Coin)) <= 0 {
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
		tx2, attestationErr2 := client.createAttestation(addr, []btcjson.ListUnspentResult{unspent})
		assert.Equal(t, errors.New(ErrorInsufficientFunds), attestationErr2)
		client.Fees.maxFee = prevMaxFee

		var unspentList []btcjson.ListUnspentResult
		unspentList = append(unspentList, unspent)
		unspentAmount := unspent.Amount
		var topupValue int64
		// add topup unspent to unspent list
		if i == topupLevel+1 {
			topupUnspent := getTopUpUnspent(t, client, test.Config, topupHash)
			unspentList = append(unspentList, topupUnspent)
			topupValue = int64(topupUnspent.Amount * Coin)
			unspentAmount += topupUnspent.Amount
		}

		tx2, attestationErr2 = client.createAttestation(addr, unspentList)
		assert.Equal(t, nil, attestationErr2)

		// verify transaction pre-image generation
		verifyTransactionPreImages(t, client, tx2, script, oceanCommitmentHash, i)

		// test attestation transaction fee bumping
		bumpErr := client.bumpAttestationFees(tx2)
		assert.Equal(t, nil, bumpErr)
		assert.Equal(t, 1-1*int(math.Min(0, float64((i%(topupLevel+1)-1)))), len(tx2.TxIn))
		assert.Equal(t, 1, len(tx2.TxOut))
		assert.Equal(t, false, (unspentAmount-(float64(tx2.TxOut[0].Value)/Coin)) <= 0)

		newFee := client.Fees.GetFee()
		newValue := tx2.TxOut[0].Value
		assert.Equal(t, int64(newFee*tx2.SerializeSize()-currentFee*tx.SerializeSize()),
			(currentValue + topupValue - newValue))
		assert.Equal(t, client.Fees.minFee+client.Fees.feeIncrement, newFee)

		// test signing and sending attestation again
		signedTx, signErr = client.signAttestation(tx2, [][]crypto.Sig{}, lastHash)
		assert.Equal(t, nil, signErr)
		txid, sendErr = client.sendAttestation(signedTx)
		assert.Equal(t, nil, sendErr)

		sideClientFake.Generate(1)
		lastHash = oceanCommitmentHash

		// Verify getUnconfirmedTx gives the unconfirmed transaction just submitted
		verifyUnconfirmed(t, client, txid, oceanCommitment)

		// create topup unspent
		if i == topupLevel {
			topupHash = createTopupUnspent(t, test.Config)
		}

		client.MainClient.Generate(1)

		// Verify no more unconfirmed transactions after new block generation
		verifyNoUnconfirmed(t, client)
		txs = append(txs, txid.String())

		client.Fees.ResetFee(true) // reset fees again

		// Now check that the new unspent is the vout from the transaction just submitted
		unspent = verifyNewUnspent(t, client, txid)
	}

	assert.Equal(t, len(txs), iterNum+1)
	verifyTxs(t, client, txs)
}
