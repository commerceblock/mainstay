package crypto

import (
	"crypto/ecdsa"
	"log"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)

// Various utility functionalities concerning key tweaking under BIP-175

// Get private key wallet readable format from a string encoded private key
func GetWalletPrivKey(privKey string) *btcutil.WIF {
	key, err := btcutil.DecodeWIF(privKey)
	if err != nil {
		log.Fatal(err)
	}
	return key
}

// Tweak a private key by adding the tweak to it's integer representation
func TweakPrivKey(walletPrivKey *btcutil.WIF, tweak []byte, chainCfg *chaincfg.Params) *btcutil.WIF {
	// Convert private key and tweak to big Int
	keyVal := new(big.Int).SetBytes(walletPrivKey.PrivKey.Serialize())
	twkVal := new(big.Int).SetBytes(tweak)

	// Add the two Ints (tweaking by scalar addition of private keys)
	resVal := new(big.Int)
	resVal.Add(keyVal, twkVal)

	// In case of addition overflow, apply modulo of the max allowed Int by the curve params, to the result
	n := new(big.Int).Set(btcec.S256().Params().N)
	resVal.Mod(resVal, n)

	// Conver the result back to bytes - new private key
	resPrivKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), resVal.Bytes())

	// Return priv key in wallet readable format
	resWalletPrivKey, err := btcutil.NewWIF(resPrivKey, chainCfg, walletPrivKey.CompressPubKey)
	if err != nil {
		log.Fatal(err)
	}
	return resWalletPrivKey
}

// Get pay to pub key hash address from a private key
func GetAddressFromPrivKey(walletPrivKey *btcutil.WIF, chainCfg *chaincfg.Params) btcutil.Address {
	return GetAddressFromPubKey(walletPrivKey.PrivKey.PubKey(), chainCfg)
}

// Tweak a pub key by adding the elliptic curve representation of the tweak to the pub key
func TweakPubKey(pubKey *btcec.PublicKey, tweak []byte) *btcec.PublicKey {
	// Get elliptic curve point for the tweak
	_, twkPubKey := btcec.PrivKeyFromBytes(btcec.S256(), tweak)

	// Add the two pub keys using addition on the elliptic curve
	resX, resY := btcec.S256().Add(pubKey.ToECDSA().X, pubKey.ToECDSA().Y, twkPubKey.ToECDSA().X, twkPubKey.ToECDSA().Y)

	return (*btcec.PublicKey)(&ecdsa.PublicKey{btcec.S256(), resX, resY})
}

// Get pay to pub key hash address from a pub key
func GetAddressFromPubKey(pubkey *btcec.PublicKey, chainCfg *chaincfg.Params) btcutil.Address {
	pubKeyHash := btcutil.Hash160(pubkey.SerializeCompressed())     // get pub key hash
	addr, err := btcutil.NewAddressPubKeyHash(pubKeyHash, chainCfg) // encode to address
	if err != nil {
		log.Fatal(err)
	}
	return addr
}

// Check whether an address has been tweaked by a specific hash
func IsAddrTweakedFromHash(address string, hash []byte, walletPrivKey *btcutil.WIF, chainCfg *chaincfg.Params) bool {
	tweakedPriv := TweakPrivKey(walletPrivKey, hash, chainCfg)
	tweakedAddr := GetAddressFromPrivKey(tweakedPriv, chainCfg)

	return address == tweakedAddr.String()
}
