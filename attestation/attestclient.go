package attestation

import (
	"encoding/hex"
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

// AttestClient structure
// Maintains RPC connections to main chain client
// Handles generating staychain next address and next transaction
// and verifying that the correct chain of transactions is maintained
type AttestClient struct {
	MainClient   *rpcclient.Client
	MainChainCfg *chaincfg.Params
	Fees         AttestFees
	pk0          string
	txid0        string
	script0      string
	pubkeys      []*btcec.PublicKey
	numOfSigs    int
	WalletPriv   *btcutil.WIF
}

// NewAttestClient returns a pointer to a new AttestClient instance
// Initially locates the genesis transaction in the main chain wallet
// and verifies that the corresponding private key is in the wallet
func NewAttestClient(config *confpkg.Config) *AttestClient {
	// Get initial private key from initial funding transaction of main client
	pk := config.InitPK()
	pkWif, errPkWif := crypto.GetWalletPrivKey(pk)
	if errPkWif != nil {
		log.Printf("Invalid private key %s\n", pk)
		log.Fatal(errPkWif)
	}
	importErr := config.MainClient().ImportPrivKeyRescan(pkWif, "init", false)
	if importErr != nil {
		log.Printf("Could not import initial private key %s\n", pk)
		log.Fatal(importErr)
	}

	multisig := config.MultisigScript()
	if multisig != "" { // if multisig attestation, parse pubkeys
		pubkeys, numOfSigs := crypto.ParseRedeemScript(config.MultisigScript())

		// verify our key is one of the multisig keys
		myFound := false
		for _, pub := range pubkeys {
			if pkWif.PrivKey.PubKey().IsEqual(pub) {
				myFound = true
			}
		}
		if !myFound {
			log.Fatal("Client address missing from multisig script")
		}

		return &AttestClient{
			MainClient:   config.MainClient(),
			MainChainCfg: config.MainChainCfg(),
			Fees:         NewAttestFees(config.FeesConfig()),
			pk0:          pk,
			txid0:        config.InitTX(),
			script0:      multisig,
			pubkeys:      pubkeys,
			numOfSigs:    numOfSigs,
			WalletPriv:   pkWif}
	}
	return &AttestClient{
		MainClient:   config.MainClient(),
		MainChainCfg: config.MainChainCfg(),
		Fees:         NewAttestFees(config.FeesConfig()),
		pk0:          pk,
		txid0:        config.InitTX(),
		script0:      multisig,
		pubkeys:      []*btcec.PublicKey{},
		numOfSigs:    1,
		WalletPriv:   pkWif}
}

// Get next attestation key by tweaking with latest hash
func (w *AttestClient) GetNextAttestationKey(hash chainhash.Hash) (*btcutil.WIF, error) {
	// Tweak priv key with the latest commitment hash
	tweakedWalletPriv, tweakErr := crypto.TweakPrivKey(w.WalletPriv, hash.CloneBytes(), w.MainChainCfg)
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

// Get next attestation address from private key
func (w *AttestClient) GetNextAttestationAddr(key *btcutil.WIF, hash chainhash.Hash) (btcutil.Address, string) {

	myAddr, _ := crypto.GetAddressFromPrivKey(key, w.MainChainCfg)

	// In multisig case tweak all initial pubkeys and import
	// a multisig address to the main client wallet
	if len(w.pubkeys) > 0 {
		var tweakedPubs []*btcec.PublicKey
		hashBytes := hash.CloneBytes()
		for _, pub := range w.pubkeys {
			tweakedPub := crypto.TweakPubKey(pub, hashBytes)
			tweakedPubs = append(tweakedPubs, tweakedPub)
		}

		multisigAddr, redeemScript := crypto.CreateMultisig(tweakedPubs, w.numOfSigs, w.MainChainCfg)

		return multisigAddr, redeemScript
	}

	return myAddr, ""
}

// Method to import address to client and report import error
func (w *AttestClient) ImportAttestationAddr(addr btcutil.Address) error {
	importErr := w.MainClient.ImportAddress(addr.String())
	if importErr != nil {
		return importErr
	}
	// importaddress for P2SH no currently supported by btcd code
	// importErr2 := w.MainClient.ImportAddress(redeemScript)
	// if importErr2 != nil {
	//     log.Printf("import error")
	//     log.Fatal(importErr2)
	// }
	return nil
}

// Generate a new transaction paying to the tweaked address and add fees
func (w *AttestClient) createAttestation(paytoaddr btcutil.Address, txunspent btcjson.ListUnspentResult) (*wire.MsgTx, error) {
	inputs := []btcjson.TransactionInput{{Txid: txunspent.TxID, Vout: txunspent.Vout}}

	amounts := map[btcutil.Address]btcutil.Amount{paytoaddr: btcutil.Amount(txunspent.Amount * 100000000)}
	msgtx, errCreate := w.MainClient.CreateRawTransaction(inputs, amounts, nil)
	if errCreate != nil {
		return nil, errCreate
	}

	// set replace-by-fee flag
	msgtx.TxIn[0].Sequence = uint32(math.Pow(2, float64(32))) - 3

	feePerByte := w.Fees.GetFee()
	fee := int64(feePerByte * msgtx.SerializeSize())
	msgtx.TxOut[0].Value -= fee

	return msgtx, nil
}

// Create new attestation transaction by removing sigs and bumping fee of existing transaction
func (w *AttestClient) bumpAttestationFees(msgtx *wire.MsgTx) error {
	// first remove any sigs
	msgtx.TxIn[0].SignatureScript = []byte{}

	// bump fees and calculate fee increment
	prevFeePerByte := w.Fees.GetFee()
	w.Fees.BumpFee()
	feePerByteIncrement := w.Fees.GetFee() - prevFeePerByte

	// increase tx fees by fee difference
	feeIncrement := int64(feePerByteIncrement * msgtx.SerializeSize())
	msgtx.TxOut[0].Value -= feeIncrement

	return nil
}

// Given a hash return the corresponding client private key and redeemscript
func (w *AttestClient) GetKeyAndScriptFromHash(hash chainhash.Hash) (btcutil.WIF, string) {
	var key btcutil.WIF
	var redeemScript string
	if !hash.IsEqual(&chainhash.Hash{}) {
		tweakedKey, _ := crypto.TweakPrivKey(w.WalletPriv, hash.CloneBytes(), w.MainChainCfg)
		key = *tweakedKey
		_, redeemScript = w.GetNextAttestationAddr(&key, hash)
	} else {
		key = *w.WalletPriv
		redeemScript = w.script0
	}
	return key, redeemScript
}

// Sign transaction using key/redeemscript pair generated by previous attested hash
func (w *AttestClient) SignTransaction(hash chainhash.Hash, msgTx wire.MsgTx) (*wire.MsgTx, string, error) {

	// Calculate private key and redeemScript from hash
	key, redeemScript := w.GetKeyAndScriptFromHash(hash)
	// Can't get redeem script from unspent as importaddress P2SH not supported
	// if txunspent.RedeemScript != "" {
	//     redeemScript = txunspent.RedeemScript
	// }

	// sign tx and send signature to main attestation client
	prevTxId := msgTx.TxIn[0].PreviousOutPoint.Hash
	prevTx, errRaw := w.MainClient.GetRawTransaction(&prevTxId)
	if errRaw != nil {
		return nil, "", errRaw
	}

	// Sign transaction
	rawTxInput := btcjson.RawTxInput{prevTxId.String(), 0, hex.EncodeToString(prevTx.MsgTx().TxOut[0].PkScript), redeemScript}
	signedMsgTx, _, errSign := w.MainClient.SignRawTransaction3(&msgTx, []btcjson.RawTxInput{rawTxInput}, []string{key.String()})
	if errSign != nil {
		return nil, "", errSign
	}
	return signedMsgTx, redeemScript, nil
}

// Sign the latest attestation transaction with the combined signatures
func (w *AttestClient) signAttestation(msgtx *wire.MsgTx, sigs [][]byte, hash chainhash.Hash) (*wire.MsgTx, error) {

	// sign generated transaction
	signedMsgTx, redeemScript, errSign := w.SignTransaction(hash, *msgtx)
	if errSign != nil {
		return nil, errSign
	}

	// MultiSig case - combine sigs and create new scriptSig
	if redeemScript != "" {
		mySigs, script := crypto.ParseScriptSig(signedMsgTx.TxIn[0].SignatureScript)
		if hex.EncodeToString(script) == redeemScript {
			combinedSigs := append(mySigs, sigs...)

			// take only numOfSigs required
			combinedScriptSig := crypto.CreateScriptSig(combinedSigs[:w.numOfSigs], script)
			signedMsgTx.TxIn[0].SignatureScript = combinedScriptSig
		}
	}

	return signedMsgTx, nil
}

// Send the latest attestation transaction
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
	} else { //might be better to store subchain on init and no need to parse all transactions every time
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
	if len(unspent) > 0 {
		for _, vout := range unspent {
			txhash, _ := chainhash.NewHashFromStr(vout.TxID)
			if w.verifyTxOnSubchain(*txhash) { //theoretically only one unspent vout, but check anyway
				return true, vout, nil
			}
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
