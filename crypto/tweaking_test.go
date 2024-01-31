// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package crypto

import (
	"encoding/hex"
	"math/big"
	"testing"

	"mainstay/clients"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/assert"
)

// Test Tweaking utility
func TestTweaking(t *testing.T) {
	sideClientFake := oceanClient.(*clients.SidechainClientFake)

	tweak, _ := sideClientFake.GetBestBlockHash()

	// test GetWalletPrivKey
	privKey, errPrivKey := GetWalletPrivKey(testConfig.InitPK())
	assert.Equal(t, nil, errPrivKey)
	assert.Equal(t, "cQca2KvrBnJJUCYa2tD4RXhiQshWLNMSK2A96ZKWo1SZkHhh3YLz", privKey.String())

	// test TweakprivKey
	tweakedPrivKey, errTweak := TweakPrivKey(privKey, tweak.CloneBytes(), mainChainCfg)
	assert.Equal(t, nil, errTweak)
	assert.Equal(t, "cQca2KvrBnJJUCYa2tD4RXhiQshWLNMSK2A96ZKWo2XQUs8qChQu", tweakedPrivKey.String())

	// test GetAddressFromPrivKey and IsAddrTweakedFromHash
	addr, errAddr := GetAddressFromPrivKey(tweakedPrivKey, mainChainCfg)
	assert.Equal(t, nil, errAddr)
	assert.Equal(t, true, IsAddrTweakedFromHash(addr.String(), tweak.CloneBytes(), privKey, mainChainCfg))
	assert.Equal(t, "mhUEBanz8ytATniaVvVNyERkHP9Vc9rpHj", addr.String())

	// Test TweakPubKey and GetAddressFromPubKey
	pubkey := privKey.PrivKey.PubKey()
	tweakedPubKey := TweakPubKey(pubkey, tweak.CloneBytes())
	tweakedPubKeyAddr, errTweakedAddr := GetAddressFromPubKey(tweakedPubKey, mainChainCfg)
	assert.Equal(t, nil, errTweakedAddr)
	assert.Equal(t, addr.String(), tweakedPubKeyAddr.String())

	// Validate tweaked private key
	importErr := testConfig.MainClient().ImportPrivKey(tweakedPrivKey)
	assert.Equal(t, nil, importErr)

	// Validate address generated from tweaked key
	tx, newTxErr := testConfig.MainClient().SendToAddress(addr, 10000)
	assert.Equal(t, nil, newTxErr)

	// Check transaction is in the wallet after importing key
	txres, txResErr := testConfig.MainClient().GetTransaction(tx)
	assert.Equal(t, nil, txResErr)
	assert.Equal(t, tx.String(), txres.TxID)
}

// Test get derivation path from tweak
func TestTweaking_getDerivationPathFromTweak(t *testing.T) {
	// use some random hash
	hashX, _ := chainhash.NewHashFromStr("abcadae1214d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hashXBytes := []byte{183, 163, 173, 152, 127, 242, 97, 102, 66, 132, 158, 94, 199, 107, 16, 71, 119, 165, 74, 181, 24, 52, 220, 108, 30, 154, 77, 33, 225, 218, 202, 171}
	assert.Equal(t, hashXBytes, hashX.CloneBytes())

	// test function with random hash
	path := getDerivationPathFromTweak(hashX.CloneBytes())

	// test output lengths
	assert.Equal(t, derivationPathSize, len(path))
	assert.Equal(t, derivationSize, len(hashXBytes))

	// test all child paths
	var testPathChild derivationPathChild

	for i := 0; i < derivationPathSize; i++ {
		copy(testPathChild[:], hashXBytes[i*derivationPathChildSize:(i+1)*derivationPathChildSize])
		assert.Equal(t, testPathChild, path[i])
	}

	// test function with invalid hash sizes
	path = getDerivationPathFromTweak([]byte{1, 2, 3})
	assert.Equal(t, derivationPath{}, path)

	path = getDerivationPathFromTweak([]byte{})
	assert.Equal(t, derivationPath{}, path)

	path = getDerivationPathFromTweak([]byte{1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3})
	assert.Equal(t, derivationPath{}, path)
}

// Test tweaking methods for child path
// Test both tweakval/tweak pub end with the
// exact same result for backwards compatiblity
func TestTweaking_childPathTweaking(t *testing.T) {
	// use some random hash
	hashX, _ := chainhash.NewHashFromStr("abcadae1214d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	hashXBytes := []byte{183, 163, 173, 152, 127, 242, 97, 102, 66, 132, 158, 94, 199, 107, 16, 71, 119, 165, 74, 181, 24, 52, 220, 108, 30, 154, 77, 33, 225, 218, 202, 171}
	assert.Equal(t, hashXBytes, hashX.CloneBytes())

	// get privkey / pubkey
	priv, _ := GetWalletPrivKey(testConfig.InitPK())
	pub := priv.PrivKey.PubKey()

	// get initial priv val and pub coordinates
	privVal := new(big.Int).SetBytes(priv.PrivKey.Serialize())
	pubX := pub.ToECDSA().X
	pubY := pub.ToECDSA().Y

	// test all child paths
	var testPathChild derivationPathChild

	copy(testPathChild[:], hashXBytes[0:derivationPathChildSize])
	tweakedVal := tweakValWithPathChild(testPathChild, privVal)
	tweakedPubX, tweakedPubY := tweakPubWithPathChild(testPathChild, pubX, pubY)

	// test matching priv - pub
	_, tweakedPrivPub := btcec.PrivKeyFromBytes(tweakedVal.Bytes())
	x := new(btcec.FieldVal)
	y := new(btcec.FieldVal)
	x.SetByteSlice(tweakedPubX.Bytes())
	y.SetByteSlice(tweakedPubY.Bytes())

	tweakedPub := btcec.NewPublicKey(x, y)
	assert.Equal(t, tweakedPrivPub, tweakedPub)

	for it := 1; it < derivationPathSize; it++ {
		copy(testPathChild[:], hashXBytes[it*derivationPathChildSize:it*derivationPathChildSize+derivationPathChildSize])
		tweakedVal = tweakValWithPathChild(testPathChild, tweakedVal)
		tweakedPubX, tweakedPubY = tweakPubWithPathChild(testPathChild, tweakedPubX, tweakedPubY)

		// test matching priv - pub
		_, tweakedPrivPub := btcec.PrivKeyFromBytes(tweakedVal.Bytes())
		x := new(btcec.FieldVal)
		y := new(btcec.FieldVal)
		x.SetByteSlice(tweakedPubX.Bytes())
		y.SetByteSlice(tweakedPubY.Bytes())

		tweakedPub := btcec.NewPublicKey(x, y)
		assert.Equal(t, tweakedPrivPub, tweakedPub)
	}

	// test tweaking whole thing brings same result
	tweakedPrivKey, errTweakPriv := TweakPrivKey(priv, hashX.CloneBytes(), mainChainCfg)
	assert.Equal(t, nil, errTweakPriv)

	addr, errGetAddr := GetAddressFromPrivKey(tweakedPrivKey, mainChainCfg)
	assert.Equal(t, nil, errGetAddr)

	tweakedPubKey := TweakPubKey(pub, hashX.CloneBytes())
	tweakedPubKeyAddr, errTweakedAddr := GetAddressFromPubKey(tweakedPubKey, mainChainCfg)
	assert.Equal(t, nil, errTweakedAddr)
	assert.Equal(t, addr.String(), tweakedPubKeyAddr.String())
}

// Test ExtendedKey Tweaking
// Test both priv/pub extended key tweaking
// verifying same result for backwards compatibility
func TestTweaking_extendedKey(t *testing.T) {
	// some random chaincode
	chainCodeBytes, _ := hex.DecodeString("abcdef710e47968aee906804f211cf10cde9a11e14908ca0f78cc55dd190ceaa")

	// get wif from config
	wif, errWif := GetWalletPrivKey(testConfig.InitPK())
	assert.Equal(t, nil, errWif)
	assert.Equal(t, "cQca2KvrBnJJUCYa2tD4RXhiQshWLNMSK2A96ZKWo1SZkHhh3YLz", wif.String())

	// get extended key from priv and pub keys
	privExtended := hdkeychain.NewExtendedKey([]byte{}, wif.PrivKey.Serialize(), chainCodeBytes, []byte{}, 0, 0, true)
	pubExtended := hdkeychain.NewExtendedKey([]byte{}, wif.PrivKey.PubKey().SerializeCompressed(), chainCodeBytes, []byte{}, 0, 0, false)

	// random hash for tweaking
	hashX, _ := chainhash.NewHashFromStr("abcadae1214d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")

	// tweak both pub and priv extended keys
	privTweaked, privTweakErr := TweakExtendedKey(privExtended, hashX.CloneBytes())
	assert.Equal(t, nil, privTweakErr)
	pubTweaked, pubTweakErr := TweakExtendedKey(pubExtended, hashX.CloneBytes())
	assert.Equal(t, nil, pubTweakErr)

	// get equivalent ECPub to test tweaking equivalence
	privTweakedECPub, privECPubErr := privTweaked.ECPubKey()
	assert.Equal(t, nil, privECPubErr)
	pubTweakedECPub, pubECPubErr := pubTweaked.ECPubKey()
	assert.Equal(t, nil, pubECPubErr)

	// cover future changes by hard-coding expected returned keys
	assert.Equal(t, privTweakedECPub, pubTweakedECPub)
	assert.Equal(t, "YZ7jREvsw9bHjkqVtQJDgKJCvthBSXbvXDmB6wSxDhm8xxWKmX94MmTDHtjRekEJBTuKwvKZhDXxUZpWV2DA5C7L7mSerASe1KtLjsnNgR",
		privTweaked.String())
	assert.Equal(t, "YZ7jREvsw9bHjkqVtQJDgKJCvthBSXbvXDmB6wSxDhm8xxWKmX94MmTDNSHMcSTojdiHtQ1UnhEvW5sUf4xuL4SCPirMeJdVwEJ3BZ74CS",
		pubTweaked.String())
}
