package crypto

import (
	"testing"

	"mainstay/clients"

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
	assert.Equal(t, "cNBumPP9tv3CZgBMnC6srrFMqkmCp5AhaYhWGLTQhs4zRKJCKRe4", tweakedPrivKey.String())

	// test GetAddressFromPrivKey and IsAddrTweakedFromHash
	addr, errAddr := GetAddressFromPrivKey(tweakedPrivKey, mainChainCfg)
	assert.Equal(t, nil, errAddr)
	assert.Equal(t, true, IsAddrTweakedFromHash(addr.String(), tweak.CloneBytes(), privKey, mainChainCfg))
	assert.Equal(t, "mgYhSzKCdWzV6c7mBn9EzXkEADVBmPJHmi", addr.String())

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
