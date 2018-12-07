// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	"bytes"

	confpkg "mainstay/config"
	"mainstay/crypto"

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
var signerTxBytes []byte
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
func (f AttestSignerFake) SendTxPreImage(tx wire.MsgTx) {
	signerTxBytes = getBytesFromTx(tx)
}

// Return signatures for received tx and hashes
func (f AttestSignerFake) GetSigs() [][]crypto.Sig {
	var sigs [][]crypto.Sig

	var msgTx wire.MsgTx
	if err := msgTx.Deserialize(bytes.NewReader(signerTxBytes)); err != nil {
		return sigs
	}
	hash, hashErr := chainhash.NewHash(signerConfirmedHashBytes)
	if hashErr != nil {
		return sigs
	}
	signedMsgTx, _, signErr := f.client.SignTransaction(*hash, msgTx)
	if signErr != nil {
		return sigs
	}
	for _, txin := range signedMsgTx.TxIn {
		scriptSig := txin.SignatureScript
		if len(scriptSig) > 0 {
			sig, _ := crypto.ParseScriptSig(scriptSig)
			sigs = append(sigs, []crypto.Sig{sig[0]})
		}
	}
	return sigs
}
