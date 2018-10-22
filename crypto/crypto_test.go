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

	// Test ParseScriptSig and CreateScriptSig
	scriptSig := "00473044022077607e068a5e4570f28430e723a3292d2c01d798df0758978a8cbc1d045aa230022000d5f85d071e697369c7c4d6e3520aa719f728ed5b511f8aa4eb93ceb615ba6501473044022077607e068a5e4570f28430e723a3292d2c01d798df0758978a8cbc1d045aa230022000d5f85d071e697369c7c4d6e3520aa719f728ed5b511f8aa4eb93ceb615ba650247512103c67926d6c06af1b6536ed189889d0adf02b7119bbe7a9f95498eff6417341c9321039596c67851f22774aa6c159b31f1ebf6581038e3573fc5710bf3d91c328679e852ae"
	sig1 := "3044022077607e068a5e4570f28430e723a3292d2c01d798df0758978a8cbc1d045aa230022000d5f85d071e697369c7c4d6e3520aa719f728ed5b511f8aa4eb93ceb615ba6501"
	sig2 := "3044022077607e068a5e4570f28430e723a3292d2c01d798df0758978a8cbc1d045aa230022000d5f85d071e697369c7c4d6e3520aa719f728ed5b511f8aa4eb93ceb615ba6502"
	redeemScript := "512103c67926d6c06af1b6536ed189889d0adf02b7119bbe7a9f95498eff6417341c9321039596c67851f22774aa6c159b31f1ebf6581038e3573fc5710bf3d91c328679e852ae"
	scriptSigBytes, _ := hex.DecodeString(scriptSig)
	sigsTest, redeemTest := ParseScriptSig(scriptSigBytes)
	assert.Equal(t, 2, len(sigsTest))
	assert.Equal(t, redeemScript, hex.EncodeToString(redeemTest))
	assert.Equal(t, sig1, hex.EncodeToString(sigsTest[0]))
	assert.Equal(t, sig2, hex.EncodeToString(sigsTest[1]))

	sig1Bytes, _ := hex.DecodeString(sig1)
	sig2Bytes, _ := hex.DecodeString(sig2)
	redeemScriptBytes, _ := hex.DecodeString(redeemScript)
	scriptSigTest := CreateScriptSig([][]byte{sig1Bytes, sig2Bytes}, redeemScriptBytes)
	assert.Equal(t, scriptSig, hex.EncodeToString(scriptSigTest))
}
