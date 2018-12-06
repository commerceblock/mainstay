// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	"testing"

	_ "mainstay/config"
	"mainstay/crypto"

	"github.com/stretchr/testify/assert"
)

// Test util functions used in
// attestsignerzmq struct for
// processing incoming sig messages
func TestAttestSigner_ZmqUtils(t *testing.T) {
	sig1 := []byte{71, 48, 68, 2, 32, 100, 88, 73, 1, 86, 42, 210, 239, 196, 136, 107, 0, 178, 223, 59, 32, 235, 58, 231, 207, 168, 87, 95, 227, 83, 207, 67, 150, 254, 26, 99, 13, 2, 32, 0, 169, 167, 160, 35, 235, 221, 136, 214, 217, 143, 64, 105, 250, 180, 188, 109, 236, 175, 117, 198, 53, 180, 24, 223, 217, 44, 199, 54, 158, 230, 227, 1}
	sig2 := []byte{71, 48, 68, 2, 32, 17, 175, 6, 205, 216, 180, 188, 216, 38, 178, 109, 17, 145, 237, 148, 1, 30, 73, 161, 54, 176, 122, 66, 6, 211, 219, 90, 216, 219, 38, 162, 137, 2, 32, 14, 61, 139, 90, 233, 169, 9, 57, 249, 101, 38, 109, 147, 244, 151, 182, 93, 136, 64, 221, 158, 172, 238, 208, 71, 106, 39, 50, 194, 185, 230, 102, 1}
	sig3 := []byte{71, 48, 68, 2, 32, 17, 175, 6, 205, 216, 180, 188, 216, 38, 178, 109, 17, 145, 237, 148, 1, 30, 73, 161, 54, 176, 122, 66, 6, 211, 219, 90, 216, 219, 38, 162, 137, 2, 32, 14, 61, 139, 90, 233, 169, 9, 57, 249, 101, 38, 109, 145, 244, 151, 182, 93, 136, 64, 221, 158, 172, 238, 208, 71, 106, 39, 50, 194, 185, 230, 102, 1}

	var msgA []byte
	var msgB []byte
	var splitMsgA [][]byte
	var splitMsgB [][]byte
	var msgs [][][]byte
	var sigs [][]crypto.Sig

	numOfTxInputs := 0

	// test 1 message 0 signature
	splitMsgA = getSplitMsgFromMsg(msgA)
	numOfTxInputs = updateNumOfTxInputs(splitMsgA, numOfTxInputs)
	assert.Equal(t, 0, numOfTxInputs)
	assert.Equal(t, [][]byte{}, splitMsgA)
	assert.Equal(t, 0, len(splitMsgA))

	msgs = [][][]byte{[][]byte{}}
	sigs = getSigsFromMsgs(msgs, numOfTxInputs)
	assert.Equal(t, [][]crypto.Sig{}, sigs)

	// test 2 messages 0 signature
	splitMsgA = getSplitMsgFromMsg(msgA)
	numOfTxInputs = updateNumOfTxInputs(splitMsgA, numOfTxInputs)
	assert.Equal(t, 0, numOfTxInputs)
	splitMsgB = getSplitMsgFromMsg(msgB)
	numOfTxInputs = updateNumOfTxInputs(splitMsgB, numOfTxInputs)
	assert.Equal(t, 0, numOfTxInputs)
	assert.Equal(t, [][]byte{}, splitMsgA)
	assert.Equal(t, [][]byte{}, splitMsgB)
	assert.Equal(t, 0, len(splitMsgA))
	assert.Equal(t, 0, len(splitMsgB))

	msgs = [][][]byte{[][]byte{}, [][]byte{}}
	sigs = getSigsFromMsgs(msgs, numOfTxInputs)
	assert.Equal(t, [][]crypto.Sig{}, sigs)

	// test 1 message 1 signature
	numOfTxInputs = 0
	msgA = sig1

	splitMsgA = getSplitMsgFromMsg(msgA)
	numOfTxInputs = updateNumOfTxInputs(splitMsgA, numOfTxInputs)
	assert.Equal(t, 1, numOfTxInputs)
	assert.Equal(t, [][]byte{sig1[1:]}, splitMsgA)
	assert.Equal(t, 1, len(splitMsgA))

	msgs = [][][]byte{splitMsgA}
	sigs = getSigsFromMsgs(msgs, numOfTxInputs)
	assert.Equal(t, [][]crypto.Sig{
		[]crypto.Sig{crypto.Sig(sig1[1:])}}, sigs)

	// test 1 message 2 signature
	numOfTxInputs = 0
	msgA = sig1
	msgA = append(msgA, sig2...)

	splitMsgA = getSplitMsgFromMsg(msgA)
	numOfTxInputs = updateNumOfTxInputs(splitMsgA, numOfTxInputs)
	assert.Equal(t, 2, numOfTxInputs)
	assert.Equal(t, [][]byte{sig1[1:], sig2[1:]}, splitMsgA)
	assert.Equal(t, 2, len(splitMsgA))

	msgs = [][][]byte{splitMsgA}
	sigs = getSigsFromMsgs(msgs, numOfTxInputs)
	assert.Equal(t, [][]crypto.Sig{
		[]crypto.Sig{crypto.Sig(sig1[1:])},
		[]crypto.Sig{crypto.Sig(sig2[1:])}}, sigs)

	// test 2 messages 1 signature
	numOfTxInputs = 0
	msgA = sig1
	msgB = sig3

	splitMsgA = getSplitMsgFromMsg(msgA)
	numOfTxInputs = updateNumOfTxInputs(splitMsgA, numOfTxInputs)
	assert.Equal(t, 1, numOfTxInputs)
	splitMsgB = getSplitMsgFromMsg(msgB)
	numOfTxInputs = updateNumOfTxInputs(splitMsgB, numOfTxInputs)
	assert.Equal(t, 1, numOfTxInputs)
	assert.Equal(t, [][]byte{sig1[1:]}, splitMsgA)
	assert.Equal(t, [][]byte{sig3[1:]}, splitMsgB)
	assert.Equal(t, 1, len(splitMsgA))
	assert.Equal(t, 1, len(splitMsgB))

	msgs = [][][]byte{splitMsgA, splitMsgB}
	sigs = getSigsFromMsgs(msgs, numOfTxInputs)
	assert.Equal(t, [][]crypto.Sig{
		[]crypto.Sig{crypto.Sig(sig1[1:]), crypto.Sig(sig3[1:])}}, sigs)

	// test 2 messages 2 signatures
	numOfTxInputs = 0
	msgA = sig1
	msgA = append(msgA, sig2...)
	msgB = sig3
	msgB = append(msgB, sig3...)

	splitMsgA = getSplitMsgFromMsg(msgA)
	numOfTxInputs = updateNumOfTxInputs(splitMsgA, numOfTxInputs)
	assert.Equal(t, 2, numOfTxInputs)
	splitMsgB = getSplitMsgFromMsg(msgB)
	numOfTxInputs = updateNumOfTxInputs(splitMsgB, numOfTxInputs)
	assert.Equal(t, 2, numOfTxInputs)
	assert.Equal(t, [][]byte{sig1[1:], sig2[1:]}, splitMsgA)
	assert.Equal(t, [][]byte{sig3[1:], sig3[1:]}, splitMsgB)
	assert.Equal(t, 2, len(splitMsgA))
	assert.Equal(t, 2, len(splitMsgB))

	msgs = [][][]byte{splitMsgA, splitMsgB}
	sigs = getSigsFromMsgs(msgs, numOfTxInputs)
	assert.Equal(t, [][]crypto.Sig{
		[]crypto.Sig{crypto.Sig(sig1[1:]), crypto.Sig(sig3[1:])},
		[]crypto.Sig{crypto.Sig(sig2[1:]), crypto.Sig(sig3[1:])}}, sigs)

	// test 2 messages 0,2 signatures
	numOfTxInputs = 0
	msgA = []byte{}
	msgB = sig3
	msgB = append(msgB, sig3...)

	splitMsgA = getSplitMsgFromMsg(msgA)
	numOfTxInputs = updateNumOfTxInputs(splitMsgA, numOfTxInputs)
	assert.Equal(t, 0, numOfTxInputs)
	splitMsgB = getSplitMsgFromMsg(msgB)
	numOfTxInputs = updateNumOfTxInputs(splitMsgB, numOfTxInputs)
	assert.Equal(t, 2, numOfTxInputs)
	assert.Equal(t, [][]byte{}, splitMsgA)
	assert.Equal(t, [][]byte{sig3[1:], sig3[1:]}, splitMsgB)
	assert.Equal(t, 0, len(splitMsgA))
	assert.Equal(t, 2, len(splitMsgB))

	msgs = [][][]byte{splitMsgA, splitMsgB}
	sigs = getSigsFromMsgs(msgs, numOfTxInputs)
	assert.Equal(t, [][]crypto.Sig{
		[]crypto.Sig{crypto.Sig(sig3[1:])},
		[]crypto.Sig{crypto.Sig(sig3[1:])}}, sigs)

	// test 2 messages 2,0 signatures
	numOfTxInputs = 0
	msgA = sig1
	msgA = append(msgA, sig2...)
	msgB = []byte{}

	splitMsgA = getSplitMsgFromMsg(msgA)
	numOfTxInputs = updateNumOfTxInputs(splitMsgA, numOfTxInputs)
	assert.Equal(t, 2, numOfTxInputs)
	splitMsgB = getSplitMsgFromMsg(msgB)
	numOfTxInputs = updateNumOfTxInputs(splitMsgB, numOfTxInputs)
	assert.Equal(t, 2, numOfTxInputs)
	assert.Equal(t, [][]byte{sig1[1:], sig2[1:]}, splitMsgA)
	assert.Equal(t, [][]byte{}, splitMsgB)
	assert.Equal(t, 2, len(splitMsgA))
	assert.Equal(t, 0, len(splitMsgB))

	msgs = [][][]byte{splitMsgA, splitMsgB}
	sigs = getSigsFromMsgs(msgs, numOfTxInputs)
	assert.Equal(t, [][]crypto.Sig{
		[]crypto.Sig{crypto.Sig(sig1[1:])},
		[]crypto.Sig{crypto.Sig(sig2[1:])}}, sigs)

	// test 2 messages 1,2 signatures
	numOfTxInputs = 0
	msgA = sig1
	msgB = sig3
	msgB = append(msgB, sig3...)

	splitMsgA = getSplitMsgFromMsg(msgA)
	numOfTxInputs = updateNumOfTxInputs(splitMsgA, numOfTxInputs)
	assert.Equal(t, 1, numOfTxInputs)
	splitMsgB = getSplitMsgFromMsg(msgB)
	numOfTxInputs = updateNumOfTxInputs(splitMsgB, numOfTxInputs)
	assert.Equal(t, 2, numOfTxInputs)
	assert.Equal(t, [][]byte{sig1[1:]}, splitMsgA)
	assert.Equal(t, [][]byte{sig3[1:], sig3[1:]}, splitMsgB)
	assert.Equal(t, 1, len(splitMsgA))
	assert.Equal(t, 2, len(splitMsgB))

	msgs = [][][]byte{splitMsgA, splitMsgB}
	sigs = getSigsFromMsgs(msgs, numOfTxInputs)
	assert.Equal(t, [][]crypto.Sig{
		[]crypto.Sig{crypto.Sig(sig1[1:]), crypto.Sig(sig3[1:])},
		[]crypto.Sig{crypto.Sig(sig3[1:])}}, sigs)

	// test 2 messages 2,1 signatures
	numOfTxInputs = 0
	msgA = sig1
	msgA = append(msgA, sig2...)
	msgB = sig3

	splitMsgA = getSplitMsgFromMsg(msgA)
	numOfTxInputs = updateNumOfTxInputs(splitMsgA, numOfTxInputs)
	assert.Equal(t, 2, numOfTxInputs)
	splitMsgB = getSplitMsgFromMsg(msgB)
	numOfTxInputs = updateNumOfTxInputs(splitMsgB, numOfTxInputs)
	assert.Equal(t, 2, numOfTxInputs)
	assert.Equal(t, [][]byte{sig1[1:], sig2[1:]}, splitMsgA)
	assert.Equal(t, [][]byte{sig3[1:]}, splitMsgB)
	assert.Equal(t, 2, len(splitMsgA))
	assert.Equal(t, 1, len(splitMsgB))

	msgs = [][][]byte{splitMsgA, splitMsgB}
	sigs = getSigsFromMsgs(msgs, numOfTxInputs)
	assert.Equal(t, [][]crypto.Sig{
		[]crypto.Sig{crypto.Sig(sig1[1:]), crypto.Sig(sig3[1:])},
		[]crypto.Sig{crypto.Sig(sig2[1:])}}, sigs)

	// test 2 messages 1,0 signatures
	numOfTxInputs = 0
	msgA = sig1
	msgB = []byte{}

	splitMsgA = getSplitMsgFromMsg(msgA)
	numOfTxInputs = updateNumOfTxInputs(splitMsgA, numOfTxInputs)
	assert.Equal(t, 1, numOfTxInputs)
	splitMsgB = getSplitMsgFromMsg(msgB)
	numOfTxInputs = updateNumOfTxInputs(splitMsgB, numOfTxInputs)
	assert.Equal(t, 1, numOfTxInputs)
	assert.Equal(t, [][]byte{sig1[1:]}, splitMsgA)
	assert.Equal(t, [][]byte{}, splitMsgB)
	assert.Equal(t, 1, len(splitMsgA))
	assert.Equal(t, 0, len(splitMsgB))

	msgs = [][][]byte{splitMsgA, splitMsgB}
	sigs = getSigsFromMsgs(msgs, numOfTxInputs)
	assert.Equal(t, [][]crypto.Sig{
		[]crypto.Sig{crypto.Sig(sig1[1:])}}, sigs)
}
