// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package crypto

import (
	"encoding/hex"
	"testing"

	"mainstay/clients"
	"mainstay/config"
	"mainstay/test"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/assert"
)

var (
	testConfig   *config.Config
	mainChainCfg *chaincfg.Params
	oceanClient  clients.SidechainClient
)

func init() {
	test := test.NewTest(false, false)
	testConfig = test.Config
	oceanClient = test.OceanClient
	mainChainCfg = testConfig.MainChainCfg()
}

// Test Multisig utility
func TestMultisig(t *testing.T) {
	multisig := "512103e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b332102f3a78a7bd6cf01c56312e7e828bef74134dfb109e59afd088526212d96518e7552ae"
	multisigAddr := "2MxBi6eodnuoVCw8McGrf1nuoVhastqoBXB"
	pubkeystr1 := "03e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b33"
	pubkeystr2 := "02f3a78a7bd6cf01c56312e7e828bef74134dfb109e59afd088526212d96518e75"
	nSigs := 1
	nPubs := 2

	// Test ParseRedeemScript
	msPubTest, nSigsTest := ParseRedeemScript(multisig)
	assert.Equal(t, nSigs, nSigsTest)
	assert.Equal(t, nPubs, len(msPubTest))
	assert.Equal(t, pubkeystr1, hex.EncodeToString(msPubTest[0].SerializeCompressed()))
	assert.Equal(t, pubkeystr2, hex.EncodeToString(msPubTest[1].SerializeCompressed()))

	// Test CreateMultisig
	msAddrTest, msTest := CreateMultisig([]*btcec.PublicKey{msPubTest[0], msPubTest[1]}, nSigs, mainChainCfg)
	assert.Equal(t, multisigAddr, msAddrTest.String())
	assert.Equal(t, multisig, msTest)
}

// Test Script utility
func TestScript(t *testing.T) {
	scriptSig := "00473044022077607e068a5e4570f28430e723a3292d2c01d798df0758978a8cbc1d045aa230022000d5f85d071e697369c7c4d6e3520aa719f728ed5b511f8aa4eb93ceb615ba6501473044022077607e068a5e4570f28430e723a3292d2c01d798df0758978a8cbc1d045aa230022000d5f85d071e697369c7c4d6e3520aa719f728ed5b511f8aa4eb93ceb615ba650247512103c67926d6c06af1b6536ed189889d0adf02b7119bbe7a9f95498eff6417341c9321039596c67851f22774aa6c159b31f1ebf6581038e3573fc5710bf3d91c328679e852ae"
	sig1 := "3044022077607e068a5e4570f28430e723a3292d2c01d798df0758978a8cbc1d045aa230022000d5f85d071e697369c7c4d6e3520aa719f728ed5b511f8aa4eb93ceb615ba6501"
	sig2 := "3044022077607e068a5e4570f28430e723a3292d2c01d798df0758978a8cbc1d045aa230022000d5f85d071e697369c7c4d6e3520aa719f728ed5b511f8aa4eb93ceb615ba6502"
	redeemScript := "512103c67926d6c06af1b6536ed189889d0adf02b7119bbe7a9f95498eff6417341c9321039596c67851f22774aa6c159b31f1ebf6581038e3573fc5710bf3d91c328679e852ae"

	// Test empty ParseScriptSig
	noSigsTest, noRedeemTest := ParseScriptSig([]byte{})
	assert.Equal(t, 0, len(noSigsTest))
	assert.Equal(t, "", hex.EncodeToString(noRedeemTest))

	// Test invalid ParseScriptSig
	invalidSigsTest, invalidRedeemTest := ParseScriptSig([]byte{0, 5, 1, 1})
	assert.Equal(t, 0, len(invalidSigsTest))
	assert.Equal(t, "", hex.EncodeToString(invalidRedeemTest))

	// Test ParseScriptSig
	scriptSigBytes, _ := hex.DecodeString(scriptSig)
	sigsTest, redeemTest := ParseScriptSig(scriptSigBytes)
	assert.Equal(t, 2, len(sigsTest))
	assert.Equal(t, redeemScript, hex.EncodeToString(redeemTest))
	assert.Equal(t, sig1, hex.EncodeToString(sigsTest[0]))
	assert.Equal(t, sig2, hex.EncodeToString(sigsTest[1]))

	// Test empty CreateScriptsig
	emptyScriptSigTest := CreateScriptSig([]Sig{}, []byte{})
	assert.Equal(t, []byte{byte(0), byte(0)}, emptyScriptSigTest)

	// Test CreateScriptsig
	sig1Bytes, _ := hex.DecodeString(sig1)
	sig2Bytes, _ := hex.DecodeString(sig2)
	redeemScriptBytes, _ := hex.DecodeString(redeemScript)
	scriptSigTest := CreateScriptSig([]Sig{sig1Bytes, sig2Bytes}, redeemScriptBytes)
	assert.Equal(t, scriptSig, hex.EncodeToString(scriptSigTest))
}

// Test Script utility for large redeem scripts
func TestScript_withOpPushData1(t *testing.T) {
	scriptSig := "00473044022077607e068a5e4570f28430e723a3292d2c01d798df0758978a8cbc1d045aa230022000d5f85d071e697369c7c4d6e3520aa719f728ed5b511f8aa4eb93ceb615ba6501473044022077607e068a5e4570f28430e723a3292d2c01d798df0758978a8cbc1d045aa230022000d5f85d071e697369c7c4d6e3520aa719f728ed5b511f8aa4eb93ceb615ba65024c6952210325bf82856a8fdcc7a2c08a933343d2c6332c4c252974d6b09b6232ea4080462621028ed149d77203c79d7524048689a80cc98f27e3427f2edaec52eae1f630978e08210254a548b59741ba35bfb085744373a8e10b1cf96e71f53356d7d97f807258d38c53ae"
	sig1 := "3044022077607e068a5e4570f28430e723a3292d2c01d798df0758978a8cbc1d045aa230022000d5f85d071e697369c7c4d6e3520aa719f728ed5b511f8aa4eb93ceb615ba6501"
	sig2 := "3044022077607e068a5e4570f28430e723a3292d2c01d798df0758978a8cbc1d045aa230022000d5f85d071e697369c7c4d6e3520aa719f728ed5b511f8aa4eb93ceb615ba6502"
	redeemScript := "52210325bf82856a8fdcc7a2c08a933343d2c6332c4c252974d6b09b6232ea4080462621028ed149d77203c79d7524048689a80cc98f27e3427f2edaec52eae1f630978e08210254a548b59741ba35bfb085744373a8e10b1cf96e71f53356d7d97f807258d38c53ae"

	// Test empty ParseScriptSig
	noSigsTest, noRedeemTest := ParseScriptSig([]byte{})
	assert.Equal(t, 0, len(noSigsTest))
	assert.Equal(t, "", hex.EncodeToString(noRedeemTest))

	// Test invalid ParseScriptSig
	invalidSigsTest, invalidRedeemTest := ParseScriptSig([]byte{0, 5, 1, 1})
	assert.Equal(t, 0, len(invalidSigsTest))
	assert.Equal(t, "", hex.EncodeToString(invalidRedeemTest))

	// Test ParseScriptSig
	scriptSigBytes, _ := hex.DecodeString(scriptSig)
	sigsTest, redeemTest := ParseScriptSig(scriptSigBytes)
	assert.Equal(t, 2, len(sigsTest))
	assert.Equal(t, redeemScript, hex.EncodeToString(redeemTest))
	assert.Equal(t, sig1, hex.EncodeToString(sigsTest[0]))
	assert.Equal(t, sig2, hex.EncodeToString(sigsTest[1]))

	// Test empty CreateScriptsig
	emptyScriptSigTest := CreateScriptSig([]Sig{}, []byte{})
	assert.Equal(t, []byte{byte(0), byte(0)}, emptyScriptSigTest)

	// Test CreateScriptsig
	sig1Bytes, _ := hex.DecodeString(sig1)
	sig2Bytes, _ := hex.DecodeString(sig2)
	redeemScriptBytes, _ := hex.DecodeString(redeemScript)
	scriptSigTest := CreateScriptSig([]Sig{sig1Bytes, sig2Bytes}, redeemScriptBytes)
	assert.Equal(t, scriptSig, hex.EncodeToString(scriptSigTest))
}
