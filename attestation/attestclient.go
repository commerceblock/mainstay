// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	"encoding/hex"
	"errors"
	"log"
	"math"

	confpkg "mainstay/config"
	"mainstay/crypto"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

// error - warning consts
const (
	WARNING_INSUFFICIENT_FUNDS = `Warning - Last unspent vout value low (less than 100*maxFee target)`
	WARNING_TOPUP_INFO_MISSING = `Warning - Topup Address and Script not set in config`

	ERROR_INSUFFICIENT_FUNDS              = `Insufficient unspent vout value (less than the maxFee target)`
	ERROR_MISSING_MULTISIG                = `No multisig used - Client must be signer and include private key`
	ERROR_MISSING_ADDRESS                 = `Client address missing from multisig script`
	ERROR_INVALID_PK                      = `Invalid private key`
	ERROR_FAILURE_IMPORTING_PK            = `Could not import initial private key`
	ERROR_FAILURE_IMPORTING_TOPUP_ADDRESS = `Could not import topup address`
	ERROR_SIGS_MISSING_FOR_TX             = `Missing signatures for transaction`
	ERROR_SIGS_MISSING_FOR_VIN            = `Missing signatures for transaction input`
	ERROR_INPUT_MISSING_FOR_TX            = `Missing input for transaction`
)

// coin in satoshis
const COIN = 100000000

// AttestClient structure
//
// This struct maintains rpc connection to the main bitcoin client
// It implements all the functionality required to generate new
// attestation addresses and new attestation transactions, as well
// as to combine signatures and send transaction to bitcoin network
//
// The struct stores initial configuration for txid and redeemscript
// It parses the initial script to extract initial pubkeys and uses
// these to generate new addresses from client commitments
//
// The struct includes an optional flag 'signerFlag'
// If this is set to true this client also stores a private key
// and can sign transactions. This option is implemented by
// external tools used to sign transactions or in unit-tests
// In the case that no multisig is used, client must be a signer
//
type AttestClient struct {
	// rpc client connection to main bitcoin client
	MainClient *rpcclient.Client

	// chain config for main bitcoin client
	MainChainCfg *chaincfg.Params

	// fees interface for getting latest / bumping fees
	Fees AttestFees

	// init configuration parameters
	// store information on initial keys and txid
	// required to set chain start and do key tweaking
	txid0       string
	script0     string
	pubkeys     []*btcec.PublicKey
	numOfSigs   int
	addrTopup   string
	scriptTopup string

	// states whether Attest Client struct is used for transaction
	// signing or simply for address tweaking and transaction creation
	// in signer case the wallet priv key of the signer is imported
	// in no signer case the wallet priv is a nil pointer
	WalletPriv      *btcutil.WIF
	WalletPrivTopup *btcutil.WIF
}

// NewAttestClient returns a pointer to a new AttestClient instance
// Initially locates the genesis transaction in the main chain wallet
// and verifies that the corresponding private key is in the wallet
func NewAttestClient(config *confpkg.Config, signerFlag ...bool) *AttestClient {

	// optional flag to set attest client as signer
	isSigner := false
	if len(signerFlag) > 0 {
		isSigner = signerFlag[0]
	}

	// top up config
	topupAddrStr := config.TopupAddress()
	topupScriptStr := config.TopupScript()
	var pkWifTopup *btcutil.WIF
	if topupAddrStr != "" && topupScriptStr != "" {
		importErr := config.MainClient().ImportAddress(topupAddrStr)
		if importErr != nil {
			log.Fatalf("%s %s\n%v\n", ERROR_FAILURE_IMPORTING_TOPUP_ADDRESS, topupAddrStr, importErr)
		}
		if isSigner {
			pkTopup := config.TopupPK()
			var errPkWifTopup error
			pkWifTopup, errPkWifTopup = crypto.GetWalletPrivKey(pkTopup)
			if errPkWifTopup != nil {
				log.Fatalf("%s %s\n%v\n", ERROR_INVALID_PK, pkTopup, errPkWifTopup)
			}
		}
	} else {
		log.Println(WARNING_TOPUP_INFO_MISSING)
	}

	// main config
	multisig := config.InitScript()
	var pkWif *btcutil.WIF
	if isSigner { // signer case import private key
		// Get initial private key
		pk := config.InitPK()
		var errPkWif error
		pkWif, errPkWif = crypto.GetWalletPrivKey(pk)
		if errPkWif != nil {
			log.Fatalf("%s %s\n%v\n", ERROR_INVALID_PK, pk, errPkWif)
		}
	} else if multisig == "" {
		log.Fatal(ERROR_MISSING_MULTISIG)
	}

	if multisig != "" { // if multisig is set, parse pubkeys
		pubkeys, numOfSigs := crypto.ParseRedeemScript(config.InitScript())

		// verify our key is one of the multisig keys in signer case
		if isSigner {
			myFound := false
			for _, pub := range pubkeys {
				if pkWif.PrivKey.PubKey().IsEqual(pub) {
					myFound = true
				}
			}
			if !myFound {
				log.Fatal(ERROR_MISSING_ADDRESS)
			}
		}

		return &AttestClient{
			MainClient:      config.MainClient(),
			MainChainCfg:    config.MainChainCfg(),
			Fees:            NewAttestFees(config.FeesConfig()),
			txid0:           config.InitTx(),
			script0:         multisig,
			pubkeys:         pubkeys,
			numOfSigs:       numOfSigs,
			addrTopup:       topupAddrStr,
			scriptTopup:     topupScriptStr,
			WalletPriv:      pkWif,
			WalletPrivTopup: pkWifTopup}
	}
	return &AttestClient{
		MainClient:      config.MainClient(),
		MainChainCfg:    config.MainChainCfg(),
		Fees:            NewAttestFees(config.FeesConfig()),
		txid0:           config.InitTx(),
		script0:         multisig,
		pubkeys:         []*btcec.PublicKey{},
		numOfSigs:       1,
		addrTopup:       topupAddrStr,
		scriptTopup:     topupScriptStr,
		WalletPriv:      pkWif,
		WalletPrivTopup: pkWifTopup}
}

// Get next attestation key by tweaking with latest commitment hash
// If attestation client is not a signer, then no key is returned
func (w *AttestClient) GetNextAttestationKey(hash chainhash.Hash) (*btcutil.WIF, error) {

	// in no signer case, client has no key - return nil
	if w.WalletPriv == nil {
		return nil, nil
	}

	// Tweak priv key with the latest commitment hash
	tweakedWalletPriv, tweakErr := crypto.TweakPrivKey(w.WalletPriv,
		hash.CloneBytes(), w.MainChainCfg)
	if tweakErr != nil {
		return nil, tweakErr
	}

	// Import tweaked priv key to wallet
	// importErr := w.MainClient.ImportPrivKeyRescan(tweakedWalletPriv, hash.String(), false)
	// if importErr != nil {
	// 	return nil, importErr
	// }

	return tweakedWalletPriv, nil
}

// Get next attestation address using the commitment hash provided
// In the multisig case this is generated by tweaking all the original
// of the multisig redeem script used to setup attestation, while in
// the single key - attest client signer case the privkey is used
func (w *AttestClient) GetNextAttestationAddr(key *btcutil.WIF, hash chainhash.Hash) (
	btcutil.Address, string) {

	// In multisig case tweak all initial pubkeys and import
	// a multisig address to the main client wallet
	if len(w.pubkeys) > 0 {
		var tweakedPubs []*btcec.PublicKey
		hashBytes := hash.CloneBytes()
		for _, pub := range w.pubkeys {
			tweakedPub := crypto.TweakPubKey(pub, hashBytes)
			tweakedPubs = append(tweakedPubs, tweakedPub)
		}

		multisigAddr, redeemScript := crypto.CreateMultisig(
			tweakedPubs, w.numOfSigs, w.MainChainCfg)

		return multisigAddr, redeemScript
	}

	// no multisig - signer case - use client key
	myAddr, _ := crypto.GetAddressFromPrivKey(key, w.MainChainCfg)
	return myAddr, ""
}

// Method to import address to client rpc wallet and report import error
// This address is required to watch unspent and mempool transactions
// IDEALLY would import the P2SH script as well, but not supported by btcsuite
func (w *AttestClient) ImportAttestationAddr(addr btcutil.Address) error {
	// import address for unspent watching
	importErr := w.MainClient.ImportAddress(addr.String())
	if importErr != nil {
		return importErr
	}

	return nil
}

// Generate a new transaction paying to the tweaked address
// Transaction inputs are generated using the previous attestation
// unspent as well as any additional topup inputs paid to wallet
// Fees are calculated using AttestFees interface and RBF flag is set manually
func (w *AttestClient) createAttestation(paytoaddr btcutil.Address, unspent []btcjson.ListUnspentResult) (
	*wire.MsgTx, error) {

	// add inputs and amount for each unspent tx
	var inputs []btcjson.TransactionInput
	amounts := map[btcutil.Address]btcutil.Amount{
		paytoaddr: btcutil.Amount(0)}

	// pay all funds to single address
	for i := 0; i < len(unspent); i++ {
		inputs = append(inputs, btcjson.TransactionInput{
			Txid: unspent[i].TxID,
			Vout: unspent[i].Vout,
		})
		amounts[paytoaddr] += btcutil.Amount(unspent[i].Amount * COIN)
	}

	// attempt to create raw transaction
	msgTx, errCreate := w.MainClient.CreateRawTransaction(inputs, amounts, nil)
	if errCreate != nil {
		return nil, errCreate
	}

	// set replace-by-fee flag
	// TODO: ? - currently only set RBF flag for attestation vin
	msgTx.TxIn[0].Sequence = uint32(math.Pow(2, float64(32))) - 3

	// return error if txout value is less than maxFee target
	maxFee := int64(w.Fees.maxFee * msgTx.SerializeSize())
	if msgTx.TxOut[0].Value < maxFee {
		return nil, errors.New(ERROR_INSUFFICIENT_FUNDS)
	}

	// print warning if txout value less than 100*maxfee target
	if msgTx.TxOut[0].Value < 100*maxFee {
		log.Println(WARNING_INSUFFICIENT_FUNDS)
	}

	// add fees using best fee-per-byte estimate
	feePerByte := w.Fees.GetFee()
	fee := int64(feePerByte * msgTx.SerializeSize())
	msgTx.TxOut[0].Value -= fee

	return msgTx, nil
}

// Create new attestation transaction by removing sigs and
// bumping fee of existing transaction with incremented fee
// The latest fee is fetched from the AttestFees API, which
// has fixed uppwer/lower fee limit and fee increment
func (w *AttestClient) bumpAttestationFees(msgTx *wire.MsgTx) error {
	// first remove any sigs
	for i := 0; i < len(msgTx.TxIn); i++ {
		msgTx.TxIn[i].SignatureScript = []byte{}
	}

	// bump fees and calculate fee increment
	prevFeePerByte := w.Fees.GetFee()
	w.Fees.BumpFee()
	feePerByteIncrement := w.Fees.GetFee() - prevFeePerByte

	// increase tx fees by fee difference
	feeIncrement := int64(feePerByteIncrement * msgTx.SerializeSize())
	msgTx.TxOut[0].Value -= feeIncrement

	return nil
}

// Given a commitment hash return the corresponding client private key tweaked
// This method should only be used in the attestation client signer case
func (w *AttestClient) GetKeyFromHash(hash chainhash.Hash) btcutil.WIF {
	if !hash.IsEqual(&chainhash.Hash{}) {
		tweakedKey, _ := crypto.TweakPrivKey(w.WalletPriv, hash.CloneBytes(), w.MainChainCfg)
		return *tweakedKey
	}
	return *w.WalletPriv
}

// Given a commitment hash return the corresponding redeemscript for the particular tweak
func (w *AttestClient) GetScriptFromHash(hash chainhash.Hash) string {
	if !hash.IsEqual(&chainhash.Hash{}) {
		_, redeemScript := w.GetNextAttestationAddr(w.WalletPriv, hash)
		return redeemScript
	}
	return w.script0
}

// Sign transaction using key/redeemscript pair generated by previous attested hash
// This method should only be used in the attestation client signer case
// Any excess transaction inputs are signed using the initial/topup private key
// and the initial/topup script, assuming they are used to topup the attestation service
func (w *AttestClient) SignTransaction(hash chainhash.Hash, msgTx wire.MsgTx) (
	*wire.MsgTx, string, error) {

	// Calculate private key and redeemScript from hash
	key := w.GetKeyFromHash(hash)
	redeemScript := w.GetScriptFromHash(hash)

	// fetch previous attestation transaction
	if len(msgTx.TxIn) <= 0 {
		return nil, "", errors.New(ERROR_INPUT_MISSING_FOR_TX)
	}

	// get prev outpoint hash in order to generate tx inputs for signing
	prevTxId := msgTx.TxIn[0].PreviousOutPoint.Hash
	prevTx, prevTxErr := w.MainClient.GetRawTransaction(&prevTxId)
	if prevTxErr != nil {
		return nil, "", prevTxErr
	}

	var inputs []btcjson.RawTxInput // new tx inputs
	var keys []string               // keys to sign inputs

	// add prev attestation tx input info and priv key
	inputs = append(inputs, btcjson.RawTxInput{prevTxId.String(), 0,
		hex.EncodeToString(prevTx.MsgTx().TxOut[0].PkScript), redeemScript})
	keys = append(keys, key.String())

	// for any remaining vins - sign with initial/topup privkey
	// this should be a very rare occasion
	for i := 1; i < len(msgTx.TxIn); i++ {
		// fetch previous attestation transaction
		prevTxId = msgTx.TxIn[i].PreviousOutPoint.Hash
		prevTx, prevTxErr = w.MainClient.GetRawTransaction(&prevTxId)
		if prevTxErr != nil {
			return nil, "", prevTxErr
		}
		inputs = append(inputs, btcjson.RawTxInput{prevTxId.String(), 0,
			hex.EncodeToString(prevTx.MsgTx().TxOut[0].PkScript), w.scriptTopup})
		keys = append(keys, w.WalletPrivTopup.String())
	}

	// attempt to sign transcation with provided inputs - keys
	signedMsgTx, _, errSign := w.MainClient.SignRawTransaction3(
		&msgTx, inputs, keys)
	if errSign != nil {
		return nil, "", errSign
	}
	return signedMsgTx, redeemScript, nil
}

// Sign the attestation transaction provided with the received signatures
// In the client signer case, client additionally adds sigs as well to the transaction
// Sigs are then combined and added to the attestation transaction inputs
func (w *AttestClient) signAttestation(msgtx *wire.MsgTx, sigs [][]crypto.Sig, hash chainhash.Hash) (
	*wire.MsgTx, error) {
	// set tx pointer and redeem script
	signedMsgTx := msgtx
	redeemScript := w.GetScriptFromHash(hash)
	if w.WalletPriv != nil { // sign transaction - signer case only
		// sign generated transaction
		var errSign error
		signedMsgTx, redeemScript, errSign = w.SignTransaction(hash, *msgtx)
		if errSign != nil {
			return nil, errSign
		}
	}

	// Check for multisig case
	// Almost always multisig is used, but we retain this backward compatible
	if redeemScript != "" {
		for i := 0; i < len(signedMsgTx.TxIn); i++ {
			// attempt to get mySigs first
			mySigs, script := crypto.ParseScriptSig(signedMsgTx.TxIn[i].SignatureScript)
			if len(mySigs) > 0 && len(script) > 0 {
				if len(sigs) > i {
					mySigs = append(mySigs, sigs[i]...)
				}
				// check we have the required number of sigs for vin
				if len(mySigs) < w.numOfSigs {
					return nil, errors.New(ERROR_SIGS_MISSING_FOR_VIN)
				}
				// take up to numOfSigs sigs
				combinedScriptSig := crypto.CreateScriptSig(mySigs[:w.numOfSigs], script)
				signedMsgTx.TxIn[i].SignatureScript = combinedScriptSig
			} else {
				// check we have all the sigs required
				if len(sigs) < len(signedMsgTx.TxIn) {
					return nil, errors.New(ERROR_SIGS_MISSING_FOR_TX)
				}
				if len(sigs[i]) < w.numOfSigs {
					return nil, errors.New(ERROR_SIGS_MISSING_FOR_VIN)
				}
				// no mySigs - just use received client sigs and script
				var redeemScriptBytes []byte
				if i == 0 {
					// for vin 0, use last attestation script
					redeemScriptBytes, _ = hex.DecodeString(redeemScript)
				} else {
					// for any other vin, use initial/topup script as we assume topup use only
					redeemScriptBytes, _ = hex.DecodeString(w.scriptTopup)
				}
				combinedScriptSig := crypto.CreateScriptSig(sigs[i][:w.numOfSigs], redeemScriptBytes)
				signedMsgTx.TxIn[i].SignatureScript = combinedScriptSig
			}
		}
	}

	return signedMsgTx, nil
}

// Send the latest attestation transaction through rpc bitcoin client connection
func (w *AttestClient) sendAttestation(msgtx *wire.MsgTx) (chainhash.Hash, error) {

	// send signed attestation
	txhash, errSend := w.MainClient.SendRawTransaction(msgtx, false)
	if errSend != nil {
		return chainhash.Hash{}, errSend
	}

	return *txhash, nil
}

// Verify that an unspent vout is on the tip of the subchain attestations
func (w *AttestClient) verifyTxOnSubchain(txid chainhash.Hash) bool {
	if txid.String() == w.txid0 { // genesis transaction
		return true
	} else {
		// might be better to store subchain on init
		// and no need to parse all transactions every time
		txraw, err := w.MainClient.GetRawTransaction(&txid)
		if err != nil {
			return false
		}

		prevtxid := txraw.MsgTx().TxIn[0].PreviousOutPoint.Hash
		return w.verifyTxOnSubchain(prevtxid)
	}
	return false
}

// Find the latest unspent vout that is on the tip of subchain attestations
func (w *AttestClient) findLastUnspent() (bool, btcjson.ListUnspentResult, error) {
	unspent, err := w.MainClient.ListUnspent()
	if err != nil {
		return false, btcjson.ListUnspentResult{}, err
	}
	for _, vout := range unspent {
		txhash, _ := chainhash.NewHashFromStr(vout.TxID)
		if w.verifyTxOnSubchain(*txhash) {
			//theoretically only one unspent vout, but check anyway
			return true, vout, nil
		}
	}
	return false, btcjson.ListUnspentResult{}, nil
}

// Find unspent vout for topup address specified in attestation client init
func (w *AttestClient) findTopupUnspent() (bool, btcjson.ListUnspentResult, error) {
	unspent, err := w.MainClient.ListUnspent()
	if err != nil {
		return false, btcjson.ListUnspentResult{}, err
	}
	for _, u := range unspent {
		// search for an address matching the topup address provided in config
		// exclude txid0, as this signals the first staychain transaction
		if u.Address == w.addrTopup && u.TxID != w.txid0 {
			return true, u, nil
		}
	}
	return false, btcjson.ListUnspentResult{}, nil
}

// Find any previously unconfirmed transactions in the client
func (w *AttestClient) getUnconfirmedTx() (bool, chainhash.Hash, error) {
	mempool, err := w.MainClient.GetRawMempool()
	if err != nil {
		return false, chainhash.Hash{}, err
	}
	for _, hash := range mempool {
		if w.verifyTxOnSubchain(*hash) {
			return true, *hash, nil
		}
	}
	return false, chainhash.Hash{}, nil
}
