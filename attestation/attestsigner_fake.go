// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	"log"

	confpkg "mainstay/config"
	"mainstay/crypto"

	"github.com/btcsuite/btcd/btcec"
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
var signerTxPreImageBytes []byte
var signerConfirmedHashBytes []byte

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
	signerConfirmedHashBytes = hash
}

// Store received new tx
func (f AttestSignerFake) SendTxPreImages(txs [][]byte) {
	signerTxPreImageBytes = SerializeBytes(txs)
}

// Return signatures for received tx and hashes
func (f AttestSignerFake) GetSigs() [][]crypto.Sig {
	// get confirmed hash from received confirmed hash bytes
	hash, hashErr := chainhash.NewHash(signerConfirmedHashBytes)
	if hashErr != nil {
		log.Printf("%v\n", hashErr)
		return nil
	}

	// get unserialized tx pre images
	txPreImages := UnserializeBytes(signerTxPreImageBytes)

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
			var sig *btcec.Signature
			var signErr error
			if i_tx == 0 {
				priv := client.GetKeyFromHash(*hash).PrivKey
				sig, signErr = priv.Sign(txPreImageHash.CloneBytes())
			} else {
				sig, signErr = client.WalletPrivTopup.PrivKey.Sign(txPreImageHash.CloneBytes())
			}
			if signErr != nil {
				log.Printf("%v\n", signErr)
				return nil
			}

			// add hash type to signature as well
			sigBytes := append(sig.Serialize(), []byte{byte(1)}...)
			sigs[i_tx] = append(sigs[i_tx], sigBytes)
		}
	}

	return sigs
}
