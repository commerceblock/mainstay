// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	_ "bytes"
	_ "encoding/hex"
	"testing"

	_ "mainstay/config"
	"mainstay/crypto"

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
	splitMsgA = UnserializeBytes(msgA)
	numOfTxInputs = updateNumOfTxInputs(splitMsgA, numOfTxInputs)
	assert.Equal(t, 0, numOfTxInputs)
	assert.Equal(t, [][]byte{}, splitMsgA)
	assert.Equal(t, 0, len(splitMsgA))

	msgs = [][][]byte{[][]byte{}}
	sigs = getSigsFromMsgs(msgs, numOfTxInputs)
	assert.Equal(t, [][]crypto.Sig{}, sigs)

	// test 2 messages 0 signature
	splitMsgA = UnserializeBytes(msgA)
	numOfTxInputs = updateNumOfTxInputs(splitMsgA, numOfTxInputs)
	assert.Equal(t, 0, numOfTxInputs)
	splitMsgB = UnserializeBytes(msgB)
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

	splitMsgA = UnserializeBytes(msgA)
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

	splitMsgA = UnserializeBytes(msgA)
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

	splitMsgA = UnserializeBytes(msgA)
	numOfTxInputs = updateNumOfTxInputs(splitMsgA, numOfTxInputs)
	assert.Equal(t, 1, numOfTxInputs)
	splitMsgB = UnserializeBytes(msgB)
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

	splitMsgA = UnserializeBytes(msgA)
	numOfTxInputs = updateNumOfTxInputs(splitMsgA, numOfTxInputs)
	assert.Equal(t, 2, numOfTxInputs)
	splitMsgB = UnserializeBytes(msgB)
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

	splitMsgA = UnserializeBytes(msgA)
	numOfTxInputs = updateNumOfTxInputs(splitMsgA, numOfTxInputs)
	assert.Equal(t, 0, numOfTxInputs)
	splitMsgB = UnserializeBytes(msgB)
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

	splitMsgA = UnserializeBytes(msgA)
	numOfTxInputs = updateNumOfTxInputs(splitMsgA, numOfTxInputs)
	assert.Equal(t, 2, numOfTxInputs)
	splitMsgB = UnserializeBytes(msgB)
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

	splitMsgA = UnserializeBytes(msgA)
	numOfTxInputs = updateNumOfTxInputs(splitMsgA, numOfTxInputs)
	assert.Equal(t, 1, numOfTxInputs)
	splitMsgB = UnserializeBytes(msgB)
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

	splitMsgA = UnserializeBytes(msgA)
	numOfTxInputs = updateNumOfTxInputs(splitMsgA, numOfTxInputs)
	assert.Equal(t, 2, numOfTxInputs)
	splitMsgB = UnserializeBytes(msgB)
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

	splitMsgA = UnserializeBytes(msgA)
	numOfTxInputs = updateNumOfTxInputs(splitMsgA, numOfTxInputs)
	assert.Equal(t, 1, numOfTxInputs)
	splitMsgB = UnserializeBytes(msgB)
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
	// empty input to Serialize
	assert.Equal(t, []byte{}, SerializeBytes([][]byte{}))
	assert.Equal(t, 0, len(SerializeBytes([][]byte{})))
	assert.Equal(t, []byte{}, SerializeBytes([][]byte(nil)))
	assert.Equal(t, 0, len(SerializeBytes([][]byte(nil))))

	// single vin unsigned tx
	tx1Bytes := []byte{2, 0, 0, 0, 1, 48, 38, 85, 184, 133, 101, 229, 118, 225, 243, 224, 5, 134, 231, 53, 91, 21, 77, 145, 198, 183, 163, 103, 103, 248, 234, 201, 83, 214, 206, 37, 195, 0, 0, 0, 0, 0, 253, 255, 255, 255, 1, 66, 158, 23, 168, 4, 0, 0, 0, 23, 169, 20, 160, 161, 96, 85, 138, 149, 193, 14, 237, 218, 58, 112, 171, 104, 24, 157, 212, 132, 203, 58, 135, 0, 0, 0, 0}

	tx1BytesWithLen := append([]byte{byte(len(tx1Bytes))}, tx1Bytes...)
	assert.Equal(t, tx1BytesWithLen, SerializeBytes([][]byte{tx1Bytes}))
	assert.Equal(t, len(tx1Bytes)+1, len(SerializeBytes([][]byte{tx1Bytes})))

	// two vin unsigned tx
	tx2Bytes := []byte{2, 0, 0, 0, 2, 108, 82, 16, 166, 228, 190, 231, 4, 131, 28, 47, 248, 172, 49, 84, 236, 95, 173, 60, 159, 155, 183, 19, 112, 116, 38, 150, 147, 8, 132, 97, 195, 0, 0, 0, 0, 0, 253, 255, 255, 255, 192, 186, 138, 193, 135, 96, 171, 236, 192, 227, 70, 94, 185, 205, 124, 215, 86, 75, 66, 176, 237, 171, 231, 118, 79, 135, 129, 194, 111, 101, 74, 159, 0, 0, 0, 0, 0, 255, 255, 255, 255, 1, 128, 161, 23, 168, 4, 0, 0, 0, 23, 169, 20, 255, 87, 124, 157, 17, 223, 243, 128, 122, 150, 92, 1, 101, 239, 50, 250, 202, 230, 56, 75, 135, 0, 0, 0, 0}

	tx2BytesWithLen := append([]byte{byte(len(tx2Bytes))}, tx2Bytes...)

	tx1and2BytesWithLen := append(tx1BytesWithLen, tx2BytesWithLen...)

	assert.Equal(t, tx1and2BytesWithLen, SerializeBytes([][]byte{tx1Bytes, tx2Bytes}))
	assert.Equal(t, len(tx1Bytes)+len(tx2Bytes)+2, len(SerializeBytes([][]byte{tx1Bytes, tx2Bytes})))

	// empty input to Unserialize
	assert.Equal(t, [][]byte{}, UnserializeBytes([]byte{}))
	assert.Equal(t, 0, len(UnserializeBytes([]byte{})))
	assert.Equal(t, [][]byte{}, UnserializeBytes([]byte(nil)))
	assert.Equal(t, 0, len(UnserializeBytes([]byte(nil))))

	// unserialize single vin
	serializedTxs := SerializeBytes([][]byte{tx1Bytes})
	assert.Equal(t, [][]byte{tx1Bytes}, UnserializeBytes(serializedTxs))

	// unserialize two vins
	serializedTxs = SerializeBytes([][]byte{tx1Bytes, tx2Bytes})
	assert.Equal(t, [][]byte{tx1Bytes, tx2Bytes}, UnserializeBytes(serializedTxs))

	// unserialize single vin with additional noise
	serializedTxs = SerializeBytes([][]byte{tx1Bytes})
	serializedTxs = append(serializedTxs, []byte{50, 1, 1}...) // add noise
	assert.Equal(t, [][]byte{tx1Bytes}, UnserializeBytes(serializedTxs))

	serializedTxs = SerializeBytes([][]byte{tx1Bytes})
	serializedTxs = append(serializedTxs, []byte{3, 1, 1}...) // add noise
	assert.Equal(t, [][]byte{tx1Bytes}, UnserializeBytes(serializedTxs))

	serializedTxs = SerializeBytes([][]byte{tx1Bytes})
	serializedTxs = append(serializedTxs, []byte{0, 1, 1}...) // add non noise edge case
	assert.Equal(t, [][]byte{tx1Bytes, []byte{}, []byte{1}}, UnserializeBytes(serializedTxs))

	serializedTxs = SerializeBytes([][]byte{tx1Bytes})
	serializedTxs = append(serializedTxs, []byte{2, 1, 1}...) // add non noise edge case
	assert.Equal(t, [][]byte{tx1Bytes, []byte{1, 1}}, UnserializeBytes(serializedTxs))
}
