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
	"github.com/btcsuite/btcd/wire"
)

// AttestSignerFake struct
//
// Implements AttestSigner interface and provides
// mock functionality for receiving sigs from signers
type AttestSignerFake struct {
	client *AttestClient
}

// store latest hash and transaction
var signerTxPreImageBytes []byte
var signerConfirmedHashBytes []byte
var signerNewHashBytes []byte

// Return new AttestSignerFake instance
func NewAttestSignerFake(config *confpkg.Config) AttestSignerFake {

	client := NewAttestClient(config, true) // isSigner flag set to allow signing transactions

	return AttestSignerFake{client: client}
}

// Store received confirmed hash
func (f AttestSignerFake) SendConfirmedHash(hash []byte) {
	signerConfirmedHashBytes = hash
}

// Store received new hash
func (f AttestSignerFake) SendNewHash(hash []byte) {
	signerNewHashBytes = hash
}

// Store received new tx
func (f AttestSignerFake) SendTxPreImages(tx []wire.MsgTx) {
	signerTxPreImageBytes = getBytesFromTx(tx)
}

// Return signatures for received tx and hashes
func (f AttestSignerFake) GetSigs() [][]crypto.Sig {
	var sigs [][]crypto.Sig

	// get confirmed hash from received confirmed hash bytes
	hash, hashErr := chainhash.NewHash(signerConfirmedHashBytes)
	if hashErr != nil {
		log.Printf("%v\n", hashErr)
		return sigs
	}

	// process each pre image transaction and sign
	txIt := 0
	for {
		// get next tx by reading byte size
		txSize := signerTxPreImageBytes[txIt]
		txPreImage := append([]byte{}, signerTxPreImageBytes[txIt+1:txIt+1+int(txSize)]...)

		// add hash type to tx serialization
		txPreImage = append(txPreImage, []byte{1, 0, 0, 0}...)
		txPreImageHash := chainhash.DoubleHashH(txPreImage)

		// sign first tx with tweaked priv key and
		// any remaining txs with topup key
		var sig *btcec.Signature
		var signErr error
		if txIt == 0 {
			priv := f.client.GetKeyFromHash(*hash).PrivKey
			sig, signErr = priv.Sign(txPreImageHash.CloneBytes())
		} else {
			sig, signErr = f.client.WalletPrivTopup.PrivKey.Sign(txPreImageHash.CloneBytes())
		}
		if signErr != nil {
			log.Printf("%v\n", signErr)
			return sigs
		}

		// add hash type to signature as well
		sigBytes := append(sig.Serialize(), []byte{byte(1)}...)
		sigs = append(sigs, []crypto.Sig{sigBytes})

		txIt += 1 + int(txSize)

		if len(signerTxPreImageBytes) <= txIt {
			break
		}
	}

	return sigs
}
