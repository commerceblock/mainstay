// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	"encoding/hex"
	"errors"
	"fmt"
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
	"github.com/btcsuite/btcutil/hdkeychain"
)

// error - warning consts
const (
	WarningInsufficientFunds            = `Warning - Last unspent vout value low (less than 100*maxFee target)`
	WarningTopupInfoMissing             = `Warning - Topup Address and/or Topup Script not set in config`
	WarningTopupPkMissing               = `Warning - Topup Private Key not set in config`
	WarningFailureImportingTopupAddress = `Could not import topup address`
	WarningFailedDecodingTopupMultisig  = `Could not decode multisig topup script`

	ErrorInsufficientFunds          = `Insufficient unspent vout value (less than the 5*maxFee target)`
	ErrorMissingMultisig            = `No multisig used - Client must be signer and include private key`
	ErrorFailedDecodingInitMultisig = `Could not decode multisig init script`
	ErrorMissingAddress             = `Client address missing from multisig script`
	ErrorInvalidPk                  = `Invalid private key`
	ErrorFailureImportingPk         = `Could not import initial private key`
	ErrorSigsMissingForTx           = `Missing signatures for transaction`
	ErrorSigsMissingForVin          = `Missing signatures for transaction input`
	ErrorInputMissingForTx          = `Missing input for transaction`
	ErrorInvalidChaincode           = `Invalid chaincode provided`
	ErrorMissingChaincodes          = `Missing chaincodes for pubkeys`
)

// coin in satoshis
const Coin = 100000000

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
	txid0           string
	script0         string
	pubkeysExtended []*hdkeychain.ExtendedKey
	pubkeys         []*btcec.PublicKey
	chaincodes      [][]byte
	numOfSigs       int
	addrTopup       string
	scriptTopup     string

	// states whether Attest Client struct is used for transaction
	// signing or simply for address tweaking and transaction creation
	// in signer case the wallet priv key of the signer is imported
	// in no signer case the wallet priv is a nil pointer
	WalletPriv      *btcutil.WIF
	WalletPrivTopup *btcutil.WIF
	WalletChainCode []byte
}

// Parse topup configuration and return private keys related to topup addresses
func parseTopupKeys(config *confpkg.Config, isSigner bool) *btcutil.WIF {
	if isSigner {
		pkTopup := config.TopupPK()
		if pkTopup != "" {
			topupWif, errPkWifTopup := crypto.GetWalletPrivKey(pkTopup)
			if errPkWifTopup != nil {
				log.Fatalf("%s %s\n%v\n", ErrorInvalidPk, pkTopup, errPkWifTopup)
			}
			return topupWif
		} else {
			log.Println(WarningTopupPkMissing)
		}
	}
	return nil
}

// Parse main configuration and return private keys related to main addresses
func parseMainKeys(config *confpkg.Config, isSigner bool) *btcutil.WIF {
	if isSigner { // signer case import private key
		// Get initial private key
		pk := config.InitPK()
		pkWif, errPkWif := crypto.GetWalletPrivKey(pk)
		if errPkWif != nil {
			log.Fatalf("%s %s\n%v\n", ErrorInvalidPk, pk, errPkWif)
		}
		return pkWif
	} else if config.InitScript() == "" {
		log.Fatal(ErrorMissingMultisig)
	}
	return nil
}

// Return new AttestClient instance for the non multisig case
// Any multisig related parameters are irrelevant and set to nil
func newNonMultisigAttestClient(config *confpkg.Config, isSigner bool, wif *btcutil.WIF, wifTopup *btcutil.WIF) *AttestClient {
	return &AttestClient{
		MainClient:      config.MainClient(),
		MainChainCfg:    config.MainChainCfg(),
		Fees:            NewAttestFees(config.FeesConfig()),
		txid0:           config.InitTx(),
		script0:         "",
		pubkeysExtended: nil,
		pubkeys:         nil,
		chaincodes:      nil,
		numOfSigs:       1,
		addrTopup:       config.TopupAddress(),
		scriptTopup:     config.TopupScript(),
		WalletPriv:      wif,
		WalletPrivTopup: wifTopup,
		WalletChainCode: []byte{}}
}

// Return new AttestClient instance for the multisig case
// Parse all the relevant pubkey and chaincode config required by the multisig
func newMultisigAttestClient(config *confpkg.Config, isSigner bool, wif *btcutil.WIF, wifTopup *btcutil.WIF) *AttestClient {
	multisig := config.InitScript()
	pubkeys, numOfSigs := crypto.ParseRedeemScript(multisig)

	// get chaincodes of pubkeys from config
	chaincodesStr := config.InitChaincodes()
	if len(chaincodesStr) != len(pubkeys) {
		log.Fatal(fmt.Sprintf("%s %d != %d", ErrorMissingChaincodes, len(chaincodesStr), len(pubkeys)))
	}
	chaincodes := make([][]byte, len(pubkeys))
	for i_c := range chaincodesStr {
		ccBytes, ccBytesErr := hex.DecodeString(chaincodesStr[i_c])
		if ccBytesErr != nil || len(ccBytes) != 32 {
			log.Fatal(fmt.Sprintf("%s %s", ErrorInvalidChaincode, chaincodesStr[i_c]))
		}
		chaincodes[i_c] = append(chaincodes[i_c], ccBytes...)
	}

	// verify our key is one of the multisig keys in signer case
	var myChaincode []byte
	if isSigner {
		myFound := false
		for i_p, pub := range pubkeys {
			if wif.PrivKey.PubKey().IsEqual(pub) {
				myFound = true
				myChaincode = chaincodes[i_p]
			}
		}
		if !myFound {
			log.Fatal(ErrorMissingAddress)
		}
	}

	// create extended keys from multisig pubs, to be used for tweaking and address generation
	// using extended keys instead of normal pubkeys in order to perform key tweaking
	// via bip-32 child derivation as opposed to regular cryptograpic tweaking
	var pubkeysExtended []*hdkeychain.ExtendedKey
	for i_p, pub := range pubkeys {
		// Ignoring any fields except key and chaincode, as these are only used for
		// child derivation and these two fields are the only required for this
		// Since any child key will be derived from these, depth limits makes no sense
		// Xpubs/xprivs are also never exported so full configuration is irrelevant
		pubkeysExtended = append(pubkeysExtended,
			hdkeychain.NewExtendedKey([]byte{}, pub.SerializeCompressed(), chaincodes[i_p], []byte{}, 0, 0, false))
	}

	return &AttestClient{
		MainClient:      config.MainClient(),
		MainChainCfg:    config.MainChainCfg(),
		Fees:            NewAttestFees(config.FeesConfig()),
		txid0:           config.InitTx(),
		script0:         multisig,
		pubkeysExtended: pubkeysExtended,
		pubkeys:         pubkeys,
		chaincodes:      chaincodes,
		numOfSigs:       numOfSigs,
		addrTopup:       config.TopupAddress(),
		scriptTopup:     config.TopupScript(),
		WalletPriv:      wif,
		WalletPrivTopup: wifTopup,
		WalletChainCode: myChaincode}
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
		log.Printf("*Client* importing top-up addr: %s ...\n", topupAddrStr)
		importErr := config.MainClient().ImportAddressRescan(topupAddrStr, "", false)
		if importErr != nil {
			log.Printf("%s (%s)\n%v\n", WarningFailureImportingTopupAddress, topupAddrStr, importErr)
		}
		pkWifTopup = parseTopupKeys(config, isSigner)
	} else {
		log.Println(WarningTopupInfoMissing)
	}

	// main config
	multisig := config.InitScript()
	var pkWif = parseMainKeys(config, isSigner)

	if multisig != "" { // if multisig is set, parse pubkeys
		return newMultisigAttestClient(config, isSigner, pkWif, pkWifTopup)
	}
	return newNonMultisigAttestClient(config, isSigner, pkWif, pkWifTopup)
}

// Get next attestation key by tweaking with latest commitment hash
// If attestation client is not a signer, then no key is returned
// Error handling excluded here, as in prod case (nil,nil) are returned
func (w *AttestClient) GetNextAttestationKey(hash chainhash.Hash) (*btcutil.WIF, error) {

	// in no signer case, client has no key - return nil
	if w.WalletPriv == nil {
		return nil, nil
	}

	// get extended key from wallet priv to do tweaking
	// pseudo bip-32 child derivation to do priv key tweaking
	// fields except key/chain code are irrelevant for child derivation
	extndKey := hdkeychain.NewExtendedKey([]byte{}, w.WalletPriv.PrivKey.Serialize(), w.WalletChainCode, []byte{}, 0, 0, true)
	tweakedExtndKey, tweakErr := crypto.TweakExtendedKey(extndKey, hash.CloneBytes())
	if tweakErr != nil {
		return nil, tweakErr
	}
	tweakedExtndPriv, tweakPrivErr := tweakedExtndKey.ECPrivKey()
	if tweakPrivErr != nil {
		return nil, tweakPrivErr
	}

	// Return priv key in wallet readable format
	tweakedWalletPriv, err := btcutil.NewWIF(tweakedExtndPriv, w.MainChainCfg, w.WalletPriv.CompressPubKey)
	if err != nil {
		return nil, err
	}

	// // Import tweaked priv key to wallet
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
// TODO: error handling
func (w *AttestClient) GetNextAttestationAddr(key *btcutil.WIF, hash chainhash.Hash) (
	btcutil.Address, string, error) {

	// In multisig case tweak all initial pubkeys and import
	// a multisig address to the main client wallet
	if len(w.pubkeysExtended) > 0 {
		// empty hash - no tweaking
		if hash.IsEqual(&chainhash.Hash{}) {
			multisigAddr, multisigScript := crypto.CreateMultisig(w.pubkeys, w.numOfSigs, w.MainChainCfg)
			return multisigAddr, multisigScript, nil
		}

		// hash non empty - tweak each pubkey
		var tweakedPubs []*btcec.PublicKey
		hashBytes := hash.CloneBytes()
		for _, pub := range w.pubkeysExtended {
			// tweak extended pubkeys
			// pseudo bip-32 child derivation to do pub key tweaking
			tweakedKey, tweakErr := crypto.TweakExtendedKey(pub, hashBytes)
			if tweakErr != nil {
				return nil, "", tweakErr
			}
			tweakedPub, tweakPubErr := tweakedKey.ECPubKey()
			if tweakPubErr != nil {
				return nil, "", tweakPubErr
			}
			tweakedPubs = append(tweakedPubs, tweakedPub)
		}

		// construct multisig and address from pubkey of extended key
		multisigAddr, redeemScript := crypto.CreateMultisig(
			tweakedPubs, w.numOfSigs, w.MainChainCfg)

		return multisigAddr, redeemScript, nil
	}

	// no multisig - signer case - use client key
	myAddr, myAddrErr := crypto.GetAddressFromPrivKey(key, w.MainChainCfg)
	if myAddrErr != nil {
		return nil, "", myAddrErr
	}
	return myAddr, "", nil
}

// Method to import address to client rpc wallet and report import error
// This address is required to watch unspent and mempool transactions
// IDEALLY would import the P2SH script as well, but not supported by btcsuite
// Optional argument to set rescan flag for import - default value set to true
func (w *AttestClient) ImportAttestationAddr(addr btcutil.Address, rescan ...bool) error {

	// check if rescan is set - defaults to true
	var isRescan = true
	if len(rescan) > 0 {
		isRescan = rescan[0]
	}

	// import address for unspent watching
	importErr := w.MainClient.ImportAddressRescan(addr.String(), "", isRescan)
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
		amounts[paytoaddr] += btcutil.Amount(unspent[i].Amount * Coin)
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
	maxFee := calcSignedTxFee(w.Fees.maxFee, msgTx.SerializeSize(),
		len(w.script0)/2, w.numOfSigs, len(inputs))
	if msgTx.TxOut[0].Value < 5*maxFee {
		return nil, errors.New(ErrorInsufficientFunds)
	}

	// print warning if txout value less than 100*maxfee target
	if msgTx.TxOut[0].Value < 100*maxFee {
		log.Println(WarningInsufficientFunds)
	}

	// add fees using best fee-per-byte estimate
	feePerByte := w.Fees.GetFee()
	fee := calcSignedTxFee(feePerByte, msgTx.SerializeSize(),
		len(w.script0)/2, w.numOfSigs, len(inputs))
	msgTx.TxOut[0].Value -= fee

	return msgTx, nil
}

// Create new attestation transaction by removing sigs and
// bumping fee of existing transaction with incremented fee
// The latest fee is fetched from the AttestFees API, which
// has fixed uppwer/lower fee limit and fee increment
func (w *AttestClient) bumpAttestationFees(msgTx *wire.MsgTx, isFeeBumped bool) error {
	// first remove any sigs
	for i := 0; i < len(msgTx.TxIn); i++ {
		msgTx.TxIn[i].SignatureScript = []byte{}
	}

	// bump fees and calculate fee increment
	var prevFeePerByte int
	if isFeeBumped {
		prevFeePerByte = w.Fees.GetPrevFee()
	} else {
		prevFeePerByte = w.Fees.GetFee()
		w.Fees.BumpFee()
	}
	feePerByteIncrement := w.Fees.GetFee() - prevFeePerByte

	// increase tx fees by fee difference
	feeIncrement := calcSignedTxFee(feePerByteIncrement, msgTx.SerializeSize(),
		len(w.script0)/2, w.numOfSigs, len(msgTx.TxIn))
	msgTx.TxOut[0].Value -= feeIncrement

	return nil
}

// Calculate the size of a signed transaction by summing the unsigned tx size
// and the redeem script size and estimated signature size of the scriptsig
func calcSignedTxSize(unsignedTxSize int, scriptSize int, numOfSigs int, numOfInputs int) int {
	return unsignedTxSize + /*script size byte*/ (1+scriptSize+
		/*00 scriptsig byte*/ 1+numOfSigs*( /*sig size byte*/ 1+72))*numOfInputs
}

// Calculate the actual fee of an unsigned transaction by taking into consideration
// the size of the script and the number of signatures required and calculating the
// aggregated transaction size with the fee per byte provided
func calcSignedTxFee(feePerByte int, unsignedTxSize int, scriptSize int, numOfSigs int, numOfInputs int) int64 {
	return int64(feePerByte * calcSignedTxSize(unsignedTxSize, scriptSize, numOfSigs, numOfInputs))
}

// Given a commitment hash return the corresponding client private key tweaked
// This method should only be used in the attestation client signer case
// Error handling excluded here as method is only for testing purposes
func (w *AttestClient) GetKeyFromHash(hash chainhash.Hash) btcutil.WIF {
	if !hash.IsEqual(&chainhash.Hash{}) {
		// get extended key from wallet priv to do tweaking
		// pseudo bip-32 child derivation to do priv key tweaking
		// fields except key/chain code are irrelevant for child derivation
		extndKey := hdkeychain.NewExtendedKey([]byte{}, w.WalletPriv.PrivKey.Serialize(), w.WalletChainCode, []byte{}, 0, 0, true)
		tweakedExtndKey, _ := crypto.TweakExtendedKey(extndKey, hash.CloneBytes())
		tweakedExtndPriv, _ := tweakedExtndKey.ECPrivKey()

		// Return priv key in wallet readable format
		tweakedKey, _ := btcutil.NewWIF(tweakedExtndPriv, w.MainChainCfg, w.WalletPriv.CompressPubKey)
		return *tweakedKey
	}
	return *w.WalletPriv
}

// Given a commitment hash return the corresponding redeemscript for the particular tweak
func (w *AttestClient) GetScriptFromHash(hash chainhash.Hash) (string, error) {
	if !hash.IsEqual(&chainhash.Hash{}) {
		_, redeemScript, scriptErr := w.GetNextAttestationAddr(w.WalletPriv, hash)
		if scriptErr != nil {
			return "", scriptErr
		}
		return redeemScript, nil
	}
	return w.script0, nil
}

// Given a bitcoin transaction generate and return the transaction pre-image for
// each of the inputs in the transaction. For each pre-image set the signature script
// of the corresponding transaction input to the redeem script for this input
// The redeem script of the first input is the tweaked (with latest commitment) init
// script and the redeem script of any excess input is set to the topup script
func (w *AttestClient) getTransactionPreImages(hash chainhash.Hash, msgTx *wire.MsgTx) ([]wire.MsgTx, error) {

	// check tx in size first
	if len(msgTx.TxIn) <= 0 {
		return []wire.MsgTx{}, errors.New(ErrorInputMissingForTx)
	}

	// pre-image txs
	var preImageTxs []wire.MsgTx

	// If init script set, add to transaction pre-image
	script, scriptErr := w.GetScriptFromHash(hash)
	if scriptErr != nil {
		return nil, scriptErr
	}
	scriptSer, decodeErr := hex.DecodeString(script)
	if decodeErr != nil {
		return []wire.MsgTx{},
			errors.New(fmt.Sprintf("%s for init script:%s\n", ErrorFailedDecodingInitMultisig, script))
	}
	// add init script bytes to txin script
	preImageTx0 := msgTx.Copy()
	preImageTx0.TxIn[0].SignatureScript = scriptSer
	preImageTxs = append(preImageTxs, *preImageTx0)

	// Add topup script to tx pre-image
	if len(msgTx.TxIn) > 1 {
		topupScriptSer, topupDecodeErr := hex.DecodeString(w.scriptTopup)
		if topupDecodeErr != nil {
			log.Printf("%s %s\n", WarningFailedDecodingTopupMultisig, w.scriptTopup)
			return preImageTxs, nil
		}
		for i := 1; i < len(msgTx.TxIn); i++ {
			// add topup script bytes to txin script
			preImageTxi := msgTx.Copy()
			preImageTxi.TxIn[i].SignatureScript = topupScriptSer
			preImageTxs = append(preImageTxs, *preImageTxi)
		}
	}

	return preImageTxs, nil
}

// Sign transaction using key/redeemscript pair generated by previous attested hash
// This method should only be used in the attestation client signer case
// Any excess transaction inputs are signed using the topup private key
// and the topup script, assuming they are used to topup the attestation service
func (w *AttestClient) SignTransaction(hash chainhash.Hash, msgTx wire.MsgTx) (
	*wire.MsgTx, string, error) {

	// Calculate private key and redeemScript from hash
	key := w.GetKeyFromHash(hash)
	redeemScript, redeemScriptErr := w.GetScriptFromHash(hash)
	if redeemScriptErr != nil {
		return nil, "", redeemScriptErr
	}

	// check tx in size first
	if len(msgTx.TxIn) <= 0 {
		return nil, "", errors.New(ErrorInputMissingForTx)
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

	// for any remaining vins - sign with topup privkey
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
	redeemScript, redeemScriptErr := w.GetScriptFromHash(hash)
	if redeemScriptErr != nil {
		return nil, redeemScriptErr
	}
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
					return nil, errors.New(ErrorSigsMissingForVin)
				}
				// take up to numOfSigs sigs
				combinedScriptSig := crypto.CreateScriptSig(mySigs[:w.numOfSigs], script)
				signedMsgTx.TxIn[i].SignatureScript = combinedScriptSig
			} else {
				// check we have all the sigs required
				if len(sigs) < len(signedMsgTx.TxIn) {
					return nil, errors.New(ErrorSigsMissingForTx)
				}
				if len(sigs[i]) < w.numOfSigs {
					return nil, errors.New(ErrorSigsMissingForVin)
				}
				// no mySigs - just use received client sigs and script
				var redeemScriptBytes []byte
				if i == 0 {
					// for vin 0, use last attestation script
					redeemScriptBytes, _ = hex.DecodeString(redeemScript)
				} else {
					// for any other vin, use topup script as we assume topup use only
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
