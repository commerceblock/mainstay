// Crypto utility Test

package main

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestCrypto(t *testing.T) {
    // TEST INIT
    test := NewTest()

    tweak, _ := test.ocean.GetBestBlockHash()
    privKey := GetWalletPrivKey(test.tx0pk)

    // Tweak private key and generate a new pay to pub key hash address
    tweakedPrivKey := TweakPrivKey(privKey, tweak.CloneBytes())
    addr := GetAddressFromPrivKey(tweakedPrivKey)

    // Validate tweaked private key
    importErr := test.btc.ImportPrivKey(tweakedPrivKey)
    assert.Equal(t, nil, importErr)

    // Validate address generated from tweaked key
    tx, newTxErr := test.btc.SendToAddress(addr, 100000)
    assert.Equal(t, nil, newTxErr)

    // Check transaction is in the wallet after importing key
    txres, txResErr := test.btc.GetTransaction(tx)
    assert.Equal(t, nil, txResErr)
    assert.Equal(t, tx.String(), txres.TxID)

    // Test that tweaking the pubkey instead produces the same address
    pubkey := privKey.PrivKey.PubKey()
    tweakedPubKey := TweakPubKey(pubkey, tweak.CloneBytes())
    tweakedPubKeyAddr := GetAddressFromPubKey(tweakedPubKey)

    assert.Equal(t, addr.String(), tweakedPubKeyAddr.String())
}
