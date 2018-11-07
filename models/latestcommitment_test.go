package models

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/assert"
)

// Test LatestCommitment high level interface
func TestLatestCommitment(t *testing.T) {
	hash0, _ := chainhash.NewHashFromStr("1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment := LatestCommitment{*hash0, int32(5)}
	assert.Equal(t, *hash0, latestCommitment.Commitment)
	assert.Equal(t, int32(5), latestCommitment.ClientPosition)
}

// Test LatestCommitment BSON interface
func TestLatestCommitmentBSON(t *testing.T) {
	hash0, _ := chainhash.NewHashFromStr("1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment := LatestCommitment{*hash0, int32(5)}

	// test marshal latestCommitment model
	bytes, errBytes := latestCommitment.MarshalBSON()
	assert.Equal(t, []uint8([]byte{0x6b, 0x0, 0x0, 0x0, 0x2, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x6d, 0x65, 0x6e, 0x74, 0x0, 0x41, 0x0, 0x0, 0x0, 0x31, 0x61, 0x33, 0x39, 0x65, 0x33, 0x34, 0x65, 0x38, 0x38, 0x31, 0x64, 0x39, 0x61, 0x31, 0x65, 0x36, 0x63, 0x64, 0x63, 0x33, 0x34, 0x31, 0x38, 0x62, 0x35, 0x34, 0x61, 0x61, 0x35, 0x37, 0x37, 0x34, 0x37, 0x31, 0x30, 0x36, 0x62, 0x63, 0x37, 0x35, 0x65, 0x39, 0x65, 0x38, 0x34, 0x34, 0x32, 0x36, 0x36, 0x36, 0x31, 0x66, 0x32, 0x37, 0x66, 0x39, 0x38, 0x61, 0x64, 0x61, 0x33, 0x62, 0x37, 0x0, 0x10, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x0, 0x5, 0x0, 0x0, 0x0, 0x0}), bytes)
	assert.Equal(t, nil, errBytes)

	// test unmarshal latestCommitment model and verify reverse works
	testLatestCommitment := &LatestCommitment{}
	testLatestCommitment.UnmarshalBSON(bytes)
	assert.Equal(t, latestCommitment.Commitment, testLatestCommitment.Commitment)
	assert.Equal(t, latestCommitment.ClientPosition, testLatestCommitment.ClientPosition)

	// test latestCommitment model to document
	doc, docErr := GetDocumentFromModel(testLatestCommitment)
	assert.Equal(t, nil, docErr)
	assert.Equal(t, latestCommitment.Commitment.String(), doc.Lookup(LATEST_COMMITMENT_COMMITMENT_NAME).StringValue())
	assert.Equal(t, latestCommitment.ClientPosition, doc.Lookup(LATEST_COMMITMENT_CLIENT_POSITION_NAME).Int32())

	// test reverse document to latestCommitment model
	testtestLatestCommitment := &LatestCommitment{}
	docErr = GetModelFromDocument(doc, testtestLatestCommitment)
	assert.Equal(t, nil, docErr)
	assert.Equal(t, latestCommitment.Commitment, testtestLatestCommitment.Commitment)
	assert.Equal(t, latestCommitment.ClientPosition, testtestLatestCommitment.ClientPosition)
}
