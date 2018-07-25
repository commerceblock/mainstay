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
    var sideClientFake *clients.SidechainClientFake
    sideClientFake = test.GetSidechainClient().(*clients.SidechainClientFake)

    tweak, _ := sideClientFake.GetBestBlockHash()
    privKey := GetWalletPrivKey(test.Tx0pk)

    // Tweak private key and generate a new pay to pub key hash address
    tweakedPrivKey := TweakPrivKey(privKey, tweak.CloneBytes(), test.BtcConfig)
    addr := GetAddressFromPrivKey(tweakedPrivKey, test.BtcConfig)
    assert.Equal(t, true, IsAddrTweakedFromHash(addr.String(), tweak.CloneBytes(), privKey, test.BtcConfig))

    // Validate tweaked private key
    importErr := test.Btc.ImportPrivKey(tweakedPrivKey)
    assert.Equal(t, nil, importErr)

    // Validate address generated from tweaked key
    tx, newTxErr := test.Btc.SendToAddress(addr, 10000)
    assert.Equal(t, nil, newTxErr)

    // Check transaction is in the wallet after importing key
    txres, txResErr := test.Btc.GetTransaction(tx)
    assert.Equal(t, nil, txResErr)
    assert.Equal(t, tx.String(), txres.TxID)

    // Test that tweaking the pubkey instead produces the same address
    pubkey := privKey.PrivKey.PubKey()
    tweakedPubKey := TweakPubKey(pubkey, tweak.CloneBytes())
    tweakedPubKeyAddr := GetAddressFromPubKey(tweakedPubKey, test.BtcConfig)

    assert.Equal(t, addr.String(), tweakedPubKeyAddr.String())
}
