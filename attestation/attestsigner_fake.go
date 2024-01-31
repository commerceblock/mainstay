// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	confpkg "mainstay/config"
	"mainstay/log"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
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
func (f AttestSignerFake) GetSigs(sigHashes [][]byte, merkle_root string) []wire.TxWitness {

	merkle_root_bytes, _ := hex.DecodeString(merkle_root)
	reversed_merkle_root := make([]byte, len(merkle_root_bytes))

	// Copy the reversed bytes
	for i, _ := range merkle_root_bytes {
		reversed_merkle_root[len(merkle_root_bytes)-1-i] = merkle_root_bytes[i]
	}
	hash, hashErr := chainhash.NewHash(reversed_merkle_root)

	if hashErr != nil {
		log.Infof("%v\n", hashErr)
		return nil
	}

	witness := make([]wire.TxWitness, len(sigHashes))

	// get sigs from each client
	for _, client := range f.clients {
		// process each sighash and sign
		for i_tx, sigHash := range sigHashes {

			// sign first tx with tweaked priv key and
			// any remaining txs with topup key
			var sig *ecdsa.Signature
			var sigBytes []byte
			if i_tx == 0 {
				priv := client.GetKeyFromHash(*hash)
				sig = ecdsa.Sign(priv.PrivKey, sigHash)
				sigBytes = append(sig.Serialize(), []byte{byte(1)}...)
				witness[i_tx] = wire.TxWitness{sigBytes, priv.PrivKey.PubKey().SerializeCompressed()}
			} else {
				sig = ecdsa.Sign(client.WalletPrivTopup.PrivKey, sigHash)
				sigBytes = append(sig.Serialize(), []byte{byte(1)}...)
				witness[i_tx] = wire.TxWitness{sigBytes, client.WalletPrivTopup.PrivKey.PubKey().SerializeCompressed()}
			}
		}
	}

	return witness
}
