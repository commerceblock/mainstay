// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package crypto

import (
	"crypto/ecdsa"
	"encoding/binary"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
)

// Various utility functionalities concerning key tweaking under BIP-175

// Get private key wallet readable format from a string encoded private key
func GetWalletPrivKey(privKey string) (*btcutil.WIF, error) {
	key, err := btcutil.DecodeWIF(privKey)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// define consts for derivation path -- all unexported
const derivationPathSize = 16                                       // derivation path size set to 16
const derivationPathChildSize = 2                                   // derivation path child size set to 2 bytes
const derivationSize = derivationPathSize * derivationPathChildSize // derivation size 32 bytes

// type for derivation path child
type derivationPathChild [derivationPathChildSize]byte

// type for derivation path
type derivationPath [derivationPathSize]derivationPathChild

// Get key derivation path from tweak hash
func getDerivationPathFromTweak(tweak []byte) derivationPath {

	// check tweak is of correct size
	// assume always correct so no error handling
	if len(tweak) != derivationSize {
		return derivationPath{}
	}

	var path derivationPath

	// iterate through tweak and pick 4-byte sizes
	// appending them to the derivation path
	for it := 0; it < derivationPathSize; it++ {
		child := new(derivationPathChild)
		copy(child[:], tweak[it*derivationPathChildSize:it*derivationPathChildSize+derivationPathChildSize])
		path[it] = derivationPathChild(*child)
	}

	return path
}

// Tweak big int value with path child
func tweakValWithPathChild(child derivationPathChild, val *big.Int) *big.Int {
	// get bytes from child path
	childBytes := make([]byte, derivationPathChildSize)
	copy(childBytes, child[:])

	// Convert child path bytes to big Int
	twkVal := new(big.Int).SetBytes(childBytes)

	// Add the two Ints (tweaking by scalar addition of private keys)
	return val.Add(val, twkVal)
}

// Tweak a private key by adding the tweak to it's integer representation
func TweakPrivKey(walletPrivKey *btcutil.WIF, tweak []byte, chainCfg *chaincfg.Params) (*btcutil.WIF, error) {

	// set initial value to big int of priv key bytes
	resVal := new(big.Int).SetBytes(walletPrivKey.PrivKey.Serialize())

	// big int S256 limit for modulo
	n := new(big.Int).Set(btcec.S256().Params().N)

	path := getDerivationPathFromTweak(tweak) // get derivation path for tweak

	// tweak initial for each path child
	for _, pathChild := range path {

		// get tweaked value
		resVal = tweakValWithPathChild(pathChild, resVal)

		// In case of addition overflow, apply modulo of the max allowed Int by the curve params, to the result
		resVal.Mod(resVal, n)
	}

	// Conver the result back to bytes - new private key
	resPrivKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), resVal.Bytes())

	// Return priv key in wallet readable format
	resWalletPrivKey, err := btcutil.NewWIF(resPrivKey, chainCfg, walletPrivKey.CompressPubKey)
	if err != nil {
		return nil, err
	}
	return resWalletPrivKey, nil
}

// Get pay to pub key hash address from a private key
func GetAddressFromPrivKey(walletPrivKey *btcutil.WIF, chainCfg *chaincfg.Params) (*btcutil.AddressWitnessPubKeyHash, error) {
	return GetAddressFromPubKey(walletPrivKey.PrivKey.PubKey(), chainCfg)
}

// Tweak pubkey with path child
func tweakPubWithPathChild(child derivationPathChild, x *big.Int, y *big.Int) (*big.Int, *big.Int) {
	// get bytes from child path
	childBytes := make([]byte, derivationPathChildSize)
	copy(childBytes, child[:])

	// Get elliptic curve point for child path bytes
	_, twkPubKey := btcec.PrivKeyFromBytes(btcec.S256(), childBytes)

	// Add the two pub keys using addition on the elliptic curve
	return btcec.S256().Add(x, y, twkPubKey.ToECDSA().X, twkPubKey.ToECDSA().Y)
}

// Tweak a bip-32 extended key (public or private) with tweak hash
// Tweak takes the form of bip-32 child derivation using tweak as index
// Under the assumed conditions this method should never return an error
// but we are including the error check for any 100% completeness
func TweakExtendedKey(extndPubKey *hdkeychain.ExtendedKey, tweak []byte) (*hdkeychain.ExtendedKey, error) {

	path := getDerivationPathFromTweak(tweak) // get derivation path for tweak

	var childErr error

	// tweak pubkey for each path child
	for _, pathChild := range path {

		// get tweak index from path child
		childBytes := []byte{0, 0, pathChild[0], pathChild[1]}
		childInt := binary.BigEndian.Uint32(childBytes)

		// get tweaked pubkey
		extndPubKey, childErr = extndPubKey.Child(childInt)
		if childErr != nil {
			return nil, childErr
		}
	}

	return extndPubKey, nil
}

// Tweak a pub key by adding the elliptic curve representation of the tweak to the pub key
func TweakPubKey(pubKey *btcec.PublicKey, tweak []byte) *btcec.PublicKey {

	path := getDerivationPathFromTweak(tweak) // get derivation path for tweak

	// set initial pubkey X/Y to current pubkey
	resX := pubKey.ToECDSA().X
	resY := pubKey.ToECDSA().Y

	// tweak pubkey for each path child
	for _, pathChild := range path {
		// get tweaked pubkey
		resX, resY = tweakPubWithPathChild(pathChild, resX, resY)
	}

	return (*btcec.PublicKey)(&ecdsa.PublicKey{btcec.S256(), resX, resY})
}

// Get pay to pub key hash address from a pub key
func GetAddressFromPubKey(pubkey *btcec.PublicKey, chainCfg *chaincfg.Params) (*btcutil.AddressWitnessPubKeyHash, error) {
	pubKeyHash := btcutil.Hash160(pubkey.SerializeCompressed())            // get pub key hash
	addr, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, chainCfg) // encode to address
	if err != nil {
		return nil, err
	}
	return addr, nil
}

// Check whether an address has been tweaked by a specific hash
func IsAddrTweakedFromHash(address string, hash []byte, walletPrivKey *btcutil.WIF, chainCfg *chaincfg.Params) bool {
	tweakedPriv, _ := TweakPrivKey(walletPrivKey, hash, chainCfg)
	tweakedAddr, _ := GetAddressFromPrivKey(tweakedPriv, chainCfg)

	return address == tweakedAddr.String()
}
