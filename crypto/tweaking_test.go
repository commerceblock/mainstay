// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package crypto

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	"mainstay/clients"

	"github.com/btcsuite/btcd/btcec"
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
	_, tweakedPrivPub := btcec.PrivKeyFromBytes(btcec.S256(), tweakedVal.Bytes())
	tweakedPub := (*btcec.PublicKey)(&ecdsa.PublicKey{btcec.S256(), tweakedPubX, tweakedPubY})
	assert.Equal(t, tweakedPrivPub, tweakedPub)

	for it := 1; it < derivationPathSize; it++ {
		copy(testPathChild[:], hashXBytes[it*derivationPathChildSize:it*derivationPathChildSize+derivationPathChildSize])
		tweakedVal = tweakValWithPathChild(testPathChild, tweakedVal)
		tweakedPubX, tweakedPubY = tweakPubWithPathChild(testPathChild, tweakedPubX, tweakedPubY)

		// test matching priv - pub
		_, tweakedPrivPub := btcec.PrivKeyFromBytes(btcec.S256(), tweakedVal.Bytes())
		tweakedPub := (*btcec.PublicKey)(&ecdsa.PublicKey{btcec.S256(), tweakedPubX, tweakedPubY})
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
