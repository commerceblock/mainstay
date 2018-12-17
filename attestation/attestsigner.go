// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	"mainstay/crypto"
)

// AttestSigner interface
//
// Provides the interface for communication with
// transaction signers. Main functionalitites are:
// - sending the last confirmed commitment hash
// - sending the new commitment (for tweaking)
// - sending the new generated transaction for signing
// - getting the signatures from signers
//
// This interface allows building communication with
// various ways - currently supporting zmq only
// This interface allows building mock struct for testing
type AttestSigner interface {
	SendConfirmedHash([]byte)
	SendTxPreImages([][]byte)
	GetSigs() [][]crypto.Sig
}
