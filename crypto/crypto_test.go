package crypto

import (
    "testing"

    "ocean-attestation/test"
    "ocean-attestation/clients"

    "github.com/stretchr/testify/assert"
)

// Crypto utility Test
func TestCrypto(t *testing.T) {
    // TEST INIT
    test := test.NewTest(false, false)
    testConfig := test.Config
    var sideClientFake *clients.SidechainClientFake
    sideClientFake = testConfig.OceanClient().(*clients.SidechainClientFake)

    tweak, _ := sideClientFake.GetBestBlockHash()
    privKey := GetWalletPrivKey(test.Tx0pk)

    // Tweak private key and generate a new pay to pub key hash address
    tweakedPrivKey := TweakPrivKey(privKey, tweak.CloneBytes(), testConfig.MainChainCfg())
    addr := GetAddressFromPrivKey(tweakedPrivKey, testConfig.MainChainCfg())
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
    tweakedPubKeyAddr := GetAddressFromPubKey(tweakedPubKey, testConfig.MainChainCfg())

    assert.Equal(t, addr.String(), tweakedPubKeyAddr.String())
}
