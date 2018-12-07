// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	"bytes"
	"encoding/hex"
	"testing"

	_ "mainstay/config"
	"mainstay/crypto"

	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"
)

// Test util functions used in
// attestsignerzmq struct for
// processing incoming sig messages
func TestAttestSigner_SigUtils(t *testing.T) {
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

// Test util functions used in
// attestsigner struct for
// processing incoming tx messages
func TestAttestSigner_TxUtils(t *testing.T) {
	// single vin unsigned tx
	tx1Bytes := []byte{2, 0, 0, 0, 1, 48, 38, 85, 184, 133, 101, 229, 118, 225, 243, 224, 5, 134, 231, 53, 91, 21, 77, 145, 198, 183, 163, 103, 103, 248, 234, 201, 83, 214, 206, 37, 195, 0, 0, 0, 0, 0, 253, 255, 255, 255, 1, 66, 158, 23, 168, 4, 0, 0, 0, 23, 169, 20, 160, 161, 96, 85, 138, 149, 193, 14, 237, 218, 58, 112, 171, 104, 24, 157, 212, 132, 203, 58, 135, 0, 0, 0, 0}

	var msgTx1 wire.MsgTx
	errDe1 := msgTx1.Deserialize(bytes.NewReader(tx1Bytes))
	assert.Equal(t, nil, errDe1)

	assert.Equal(t, 1, len(msgTx1.TxIn))
	assert.Equal(t, "", hex.EncodeToString(msgTx1.TxIn[0].SignatureScript))

	assert.Equal(t, tx1Bytes, getBytesFromTx(msgTx1))

	// two vin unsigned tx
	tx2Bytes := []byte{2, 0, 0, 0, 2, 108, 82, 16, 166, 228, 190, 231, 4, 131, 28, 47, 248, 172, 49, 84, 236, 95, 173, 60, 159, 155, 183, 19, 112, 116, 38, 150, 147, 8, 132, 97, 195, 0, 0, 0, 0, 0, 253, 255, 255, 255, 192, 186, 138, 193, 135, 96, 171, 236, 192, 227, 70, 94, 185, 205, 124, 215, 86, 75, 66, 176, 237, 171, 231, 118, 79, 135, 129, 194, 111, 101, 74, 159, 0, 0, 0, 0, 0, 255, 255, 255, 255, 1, 128, 161, 23, 168, 4, 0, 0, 0, 23, 169, 20, 255, 87, 124, 157, 17, 223, 243, 128, 122, 150, 92, 1, 101, 239, 50, 250, 202, 230, 56, 75, 135, 0, 0, 0, 0}

	var msgTx2 wire.MsgTx
	errDe2 := msgTx2.Deserialize(bytes.NewReader(tx2Bytes))
	assert.Equal(t, nil, errDe2)

	assert.Equal(t, 2, len(msgTx2.TxIn))
	assert.Equal(t, "", hex.EncodeToString(msgTx2.TxIn[0].SignatureScript))
	assert.Equal(t, "", hex.EncodeToString(msgTx2.TxIn[1].SignatureScript))

	assert.Equal(t, tx2Bytes, getBytesFromTx(msgTx2))
}
