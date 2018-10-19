package crypto

import (
	"encoding/hex"
	"mainstay/clients"
	"mainstay/test"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/stretchr/testify/assert"
)

// Crypto utility Test
func TestCrypto(t *testing.T) {
	// TEST INIT
	test := test.NewTest(false, false)
	testConfig := test.Config
	mainChainCfg := testConfig.MainChainCfg()
	var sideClientFake *clients.SidechainClientFake
	sideClientFake = testConfig.OceanClient().(*clients.SidechainClientFake)

	tweak, _ := sideClientFake.GetBestBlockHash()
	privKey := GetWalletPrivKey(testConfig.InitPK())

	// Tweak private key and generate a new pay to pub key hash address
	tweakedPrivKey := TweakPrivKey(privKey, tweak.CloneBytes(), mainChainCfg)
	addr := GetAddressFromPrivKey(tweakedPrivKey, mainChainCfg)
	assert.Equal(t, true, IsAddrTweakedFromHash(addr.String(), tweak.CloneBytes(), privKey, testConfig.MainChainCfg()))

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

	// Test that tweaking the pubkey instead produces the same address
	pubkey := privKey.PrivKey.PubKey()
	tweakedPubKey := TweakPubKey(pubkey, tweak.CloneBytes())
	tweakedPubKeyAddr := GetAddressFromPubKey(tweakedPubKey, mainChainCfg)

	assert.Equal(t, addr.String(), tweakedPubKeyAddr.String())

	// Test CreateMultisig, ParseRedeemScript - test works both ways
	multisig := "512103e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b332102f3a78a7bd6cf01c56312e7e828bef74134dfb109e59afd088526212d96518e7552ae"
	multisigAddr := "2MxBi6eodnuoVCw8McGrf1nuoVhastqoBXB"
	pubkeystr1 := "03e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b33"
	pubkeystr2 := "02f3a78a7bd6cf01c56312e7e828bef74134dfb109e59afd088526212d96518e75"
	nSigs := 1
	nPubs := 2

	msPubTest, nSigsTest := ParseRedeemScript(multisig)
	assert.Equal(t, nSigs, nSigsTest)
	assert.Equal(t, nPubs, len(msPubTest))
	assert.Equal(t, pubkeystr1, hex.EncodeToString(msPubTest[0].SerializeCompressed()))
	assert.Equal(t, pubkeystr2, hex.EncodeToString(msPubTest[1].SerializeCompressed()))

	msAddrTest, msTest := CreateMultisig([]*btcec.PublicKey{msPubTest[0], msPubTest[1]}, nSigs, mainChainCfg)
	assert.Equal(t, multisigAddr, msAddrTest.String())
	assert.Equal(t, multisig, msTest)
}
