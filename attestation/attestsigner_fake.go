// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	confpkg "mainstay/config"
	"mainstay/crypto"
	"mainstay/log"

	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// AttestSignerFake struct
//
// Implements AttestSigner interface and provides
// mock functionality for receiving sigs from signers
type AttestSignerFake struct {
	clients []*AttestClient
}

// store latest hash and transaction
var signerTxPreImageBytesFake []byte
var signerConfirmedHashBytesFake []byte

// Return new AttestSignerFake instance
func NewAttestSignerFake(configs []*confpkg.Config) AttestSignerFake {

	var clients []*AttestClient
	for _, config := range configs {
		// isSigner flag set to allow signing transactions
		clients = append(clients, NewAttestClient(config, true))
	}

	return AttestSignerFake{clients: clients}
}

// Resubscribe - do nothing
func (f AttestSignerFake) ReSubscribe() {
	return
}

// Store received confirmed hash
func (f AttestSignerFake) SendConfirmedHash(hash []byte) {
	signerConfirmedHashBytesFake = hash
}

// Store received new tx
func (f AttestSignerFake) SendTxPreImages(txs [][]byte) {
	signerTxPreImageBytesFake = SerializeBytes(txs)
}

// Return signatures for received tx and hashes
func (f AttestSignerFake) GetSigs(txHash string, redeem_script string, merkle_root string) [][]crypto.Sig {
	// get confirmed hash from received confirmed hash bytes
	hash, hashErr := chainhash.NewHash(signerConfirmedHashBytesFake)
	if hashErr != nil {
		log.Infof("%v\n", hashErr)
		return nil
	}

	// get unserialized tx pre images
	txPreImages := UnserializeBytes(signerTxPreImageBytesFake)

	sigs := make([][]crypto.Sig, len(txPreImages)) // init sigs

	// get sigs from each client
	for _, client := range f.clients {
		// process each pre image transaction and sign
		for i_tx, txPreImage := range txPreImages {
			// add hash type to tx serialization
			txPreImage = append(txPreImage, []byte{1, 0, 0, 0}...)
			txPreImageHash := chainhash.DoubleHashH(txPreImage)

			// sign first tx with tweaked priv key and
			// any remaining txs with topup key
			var sig *ecdsa.Signature
			if i_tx == 0 {
				priv := client.GetKeyFromHash(*hash).PrivKey
				sig = ecdsa.Sign(priv, txPreImageHash.CloneBytes())
			} else {
				sig = ecdsa.Sign(client.WalletPrivTopup.PrivKey, txPreImageHash.CloneBytes())
			}

			// add hash type to signature as well
			sigBytes := append(sig.Serialize(), []byte{byte(1)}...)
			sigs[i_tx] = append(sigs[i_tx], sigBytes)
		}
	}

	return sigs
}
