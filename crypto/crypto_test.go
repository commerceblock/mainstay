// Crypto utility Test

package crypto

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "ocean-attestation/test"
)

func TestCrypto(t *testing.T) {
    // TEST INIT
    test := test.NewTest()

    tweak, _ := test.Ocean.GetBestBlockHash()
    privKey := GetWalletPrivKey(test.Tx0pk)

    // Tweak private key and generate a new pay to pub key hash address
    tweakedPrivKey := TweakPrivKey(privKey, tweak.CloneBytes(), test.BtcConfig)
    addr := GetAddressFromPrivKey(tweakedPrivKey, test.BtcConfig)

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
