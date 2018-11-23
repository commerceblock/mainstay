// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	confpkg "mainstay/config"
)

// AttestSignerFake struct
//
// Implements AttestSigner interface and provides
// mock functionality for receiving sigs from signers
type AttestSignerFake struct{}

// Return new AttestSignerFake instance
func NewAttestSignerFake(config *confpkg.Config) AttestSignerFake {
	return AttestSignerFake{}
}

// Store received confirmed hash
func (f AttestSignerFake) SendConfirmedHash(hash []byte) {
	return
}

// Store received new hash
func (f AttestSignerFake) SendNewHash(hash []byte) {
	return
}

// Store received new tx
func (f AttestSignerFake) SendNewTx(hash []byte) {
	return
}

// Return signatures for received tx and hashes
func (f AttestSignerFake) GetSigs() [][]byte {
	var sigs [][]byte

	return sigs
}
