// Key extract tool
package main

import (
	"encoding/hex"

	"mainstay/crypto"
	"mainstay/log"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
)

// Tool that assists with calculating tweaked private key
// and redeem script the same way AttestClient does
// These can be used to move the funds from a staychain tx
func main() {
	tweak := "3330e863b85e4df79370b1252a44cbd1aea2f42c586eba20b833762171dbaefa"
	tweakHash, _ := chainhash.NewHashFromStr(tweak)

	privkey := "xxx-priv-xxx"
	wif, _ := crypto.GetWalletPrivKey(privkey)

	chaincode := "ffdf7ece79e83f0f479a37832d770294014edc6884b0c8bfa2e0aaf51fb00229"
	chaincodeBytes, _ := hex.DecodeString(chaincode)

	extndKey := hdkeychain.NewExtendedKey([]byte{}, wif.PrivKey.Serialize(), chaincodeBytes, []byte{}, 0, 0, true)
	tweakedExtndKey, _ := crypto.TweakExtendedKey(extndKey, tweakHash.CloneBytes())
	tweakedExtndPriv, _ := tweakedExtndKey.ECPrivKey()
	tweakedWif, _ := btcutil.NewWIF(tweakedExtndPriv, &chaincfg.MainNetParams, wif.CompressPubKey)
	log.Infoln(tweakedWif.String())

	redeemScript := "xxx-redeemscript-xxx"
	chaincodes := []string{"ffdf7ece79e83f0f479a37832d770294014edc6884b0c8bfa2e0aaf51fb00229", "ffdf7ece79e83f0f479a37832d770294014edc6884b0c8bfa2e0aaf51fb00229"}
	pubkeys, numOfSigs := crypto.ParseRedeemScript(redeemScript)
	chaincodesBytes := make([][]byte, len(pubkeys))
	for i_c := range chaincodes {
		ccBytes, _ := hex.DecodeString(chaincodes[i_c])
		chaincodesBytes[i_c] = append(chaincodesBytes[i_c], ccBytes...)
	}
	var pubkeysExtended []*hdkeychain.ExtendedKey
	for i_p, pub := range pubkeys {
		pubkeysExtended = append(pubkeysExtended,
			hdkeychain.NewExtendedKey([]byte{}, pub.SerializeCompressed(), chaincodesBytes[i_p], []byte{}, 0, 0, false))
	}
	var tweakedPubkeys []*btcec.PublicKey
	for _, pub := range pubkeysExtended {
		// tweak extended pubkeys
		// pseudo bip-32 child derivation to do pub key tweaking
		tweakedKey, _ := crypto.TweakExtendedKey(pub, tweakHash.CloneBytes())
		tweakedPub, _ := tweakedKey.ECPubKey()
		tweakedPubkeys = append(tweakedPubkeys, tweakedPub)
	}
	multisigAddr, redeemScript := crypto.CreateMultisig(tweakedPubkeys, numOfSigs, &chaincfg.MainNetParams)
	log.Infoln(multisigAddr)
	log.Infoln(redeemScript)
}
