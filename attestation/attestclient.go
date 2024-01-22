// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"

	confpkg "mainstay/config"
	"mainstay/crypto"
	"mainstay/log"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
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
	ErrorTopUpScriptNumSigs         = `Different number of signatures in Init script to top-up script`
)

// coin in satoshis
const Coin = 100000000

const signedTxSize = 110 // bytes

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
				log.Errorf("%s %s\n%v\n", ErrorInvalidPk, pkTopup, errPkWifTopup)
			}
			return topupWif
		} else {
			log.Warnln(WarningTopupPkMissing)
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
			log.Errorf("%s %s\n%v\n", ErrorInvalidPk, pk, errPkWif)
		}
		return pkWif
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
		log.Infof("*Client* importing top-up addr: %s ...\n", topupAddrStr)
		importErr := config.MainClient().ImportAddressRescan(topupAddrStr, "", false)
		if importErr != nil {
			log.Warnf("%s (%s)\n%v\n", WarningFailureImportingTopupAddress, topupAddrStr, importErr)
		}
		pkWifTopup = parseTopupKeys(config, isSigner)
	} else {
		log.Warnln(WarningTopupInfoMissing)
	}

	// main config
	var pkWif = parseMainKeys(config, isSigner)

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
// In case of single key - attest client signer case the privkey is used
// TODO: error handling
func (w *AttestClient) GetNextAttestationAddr(key *btcutil.WIF, hash chainhash.Hash) (
	*btcutil.AddressWitnessPubKeyHash, string, error) {
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

	// Create the msgTx using wire.NewMsgTx
    msgTx := wire.NewMsgTx(wire.TxVersion)

    // Add transaction inputs
    for i := 0; i < len(unspent); i++ {
		hash, err := chainhash.NewHashFromStr(unspent[i].TxID)
		if err != nil {
			panic("invalid txid: "+ unspent[i].TxID)
		}
		outpoint := wire.NewOutPoint(hash, unspent[i].Vout)
        in := wire.NewTxIn(outpoint, nil, nil)
        msgTx.AddTxIn(in)
    }

    // Add transaction output
    totalAmount := btcutil.Amount(0)
    for _, utxo := range unspent {
        totalAmount += btcutil.Amount(utxo.Amount * Coin)
    }
	paytoaddrScript, err := txscript.PayToAddrScript(paytoaddr)
	if err != nil {
		panic("invalid paytoaddr: " + paytoaddr.String())
	}
    out := wire.NewTxOut(int64(totalAmount), paytoaddrScript)
    msgTx.AddTxOut(out)

	// set replace-by-fee flag
	// TODO: ? - currently only set RBF flag for attestation vin
	msgTx.TxIn[0].Sequence = uint32(math.Pow(2, float64(32))) - 3

	// return error if txout value is less than maxFee target
	maxFee := calcSignedTxFee(w.Fees.maxFee)
	if msgTx.TxOut[0].Value < 5*maxFee {
		return nil, errors.New(ErrorInsufficientFunds)
	}

	// print warning if txout value less than 100*maxfee target
	if msgTx.TxOut[0].Value < 100*maxFee {
		log.Warnln(WarningInsufficientFunds)
	}

	// add fees using best fee-per-byte estimate
	feePerByte := w.Fees.GetFee()
	fee := calcSignedTxFee(feePerByte)
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
	feeIncrement := calcSignedTxFee(feePerByteIncrement)
	msgTx.TxOut[0].Value -= feeIncrement

	return nil
}

// Calculate the actual fee of an unsigned transaction by taking into consideration
// the size of the script and the number of signatures required and calculating the
// aggregated transaction size with the fee per byte provided
func calcSignedTxFee(feePerByte int) int64 {
	return int64(feePerByte * signedTxSize)
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
			log.Warnf("%s %s\n", WarningFailedDecodingTopupMultisig, w.scriptTopup)
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
    *wire.MsgTx, error) {
    // Calculate private key
    key := w.GetKeyFromHash(hash)
    // check tx in size first
    if len(msgTx.TxIn) <= 0 {
        return nil, errors.New(ErrorInputMissingForTx)
    }
    // get prev outpoint hash in order to generate tx inputs for signing
    prevTxId := msgTx.TxIn[0].PreviousOutPoint.Hash
    prevTx, prevTxErr := w.MainClient.GetRawTransaction(&prevTxId)
    if prevTxErr != nil {
        log.Infof("Error: %v", prevTxErr)
    }
	txOut := wire.NewTxOut(prevTx.MsgTx().TxOut[0].Value, prevTx.MsgTx().TxOut[0].PkScript)

    a := txscript.NewCannedPrevOutputFetcher(txOut.PkScript, txOut.Value)
    sigHashes := txscript.NewTxSigHashes(&msgTx, a)
    signature, err := txscript.WitnessSignature(&msgTx, sigHashes, 0, txOut.Value, txOut.PkScript, txscript.SigHashAll, key.PrivKey, true)
    if err != nil {
        log.Infof("WitnessSignature err %v", err)
    }
    msgTx.TxIn[0].Witness = signature

    if len(msgTx.TxIn) > 1 {
		topupTxId := msgTx.TxIn[1].PreviousOutPoint.Hash
		topupTxIndex := msgTx.TxIn[1].PreviousOutPoint.Index
		topupTx, topupTxErr := w.MainClient.GetRawTransaction(&topupTxId)
		if topupTxErr != nil {
			log.Infof("Error: %v", topupTxErr)
		}
		topupTxOut := wire.NewTxOut(topupTx.MsgTx().TxOut[topupTxIndex].Value, topupTx.MsgTx().TxOut[topupTxIndex].PkScript)
        a = txscript.NewCannedPrevOutputFetcher(topupTxOut.PkScript, topupTxOut.Value)
        sigHashes = txscript.NewTxSigHashes(&msgTx, a)
        signature, err = txscript.WitnessSignature(&msgTx, sigHashes, 1, topupTxOut.Value, topupTxOut.PkScript, txscript.SigHashAll, w.WalletPrivTopup.PrivKey, true)
        if err != nil {
            log.Infof("WitnessSignature err %v", err)
        }
        msgTx.TxIn[1].Witness = signature
    }

    return &msgTx, nil
}

// Sign the attestation transaction provided with the received signatures
// In the client signer case, client additionally adds sigs as well to the transaction
// Sigs are then combined and added to the attestation transaction inputs
func (w *AttestClient) signAttestation(msgtx *wire.MsgTx, sigs [][]crypto.Sig, hash chainhash.Hash) (
	*wire.MsgTx, error) {
	// set tx pointer and redeem script
	signedMsgTx := msgtx
	if w.WalletPriv != nil { // sign transaction - signer case only
		// sign generated transaction
		var errSign error
		signedMsgTx, _, errSign = w.SignTransaction(hash, *msgtx)
		if errSign != nil {
			return nil, errSign
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
