package models

import (
	"testing"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/stretchr/testify/assert"
)

// Test AttestationInfo high level interface
func TestAttestationInfo(t *testing.T) {
	info := AttestationInfo{
		Txid:      "f123434e881d9a1e6cdc3418b54bb57747106bc75e9e84426661f27f98ada3b7",
		Blockhash: "abcde34e881d9a1e6cdc3418b54bb57747106bc75e9e84426661f27f98ada3b7",
		Fee:       float64(0.0001),
		Time:      int64(1542121293)}
	assert.Equal(t, "f123434e881d9a1e6cdc3418b54bb57747106bc75e9e84426661f27f98ada3b7", info.Txid)
	assert.Equal(t, "abcde34e881d9a1e6cdc3418b54bb57747106bc75e9e84426661f27f98ada3b7", info.Blockhash)
	assert.Equal(t, float64(0.0001), info.Fee)
	assert.Equal(t, int64(1542121293), info.Time)
}

// Test AttestationInfo BSON interface
func AttestationInfoBSON(t *testing.T) {
	info := AttestationInfo{
		Txid:      "f123434e881d9a1e6cdc3418b54bb57747106bc75e9e84426661f27f98ada3b7",
		Blockhash: "abcde34e881d9a1e6cdc3418b54bb57747106bc75e9e84426661f27f98ada3b7",
		Fee:       float64(0.0001),
		Time:      int64(1542121293)}

	// test marshal AttestationInfo model
	bytes, errBytes := bson.Marshal(info)
	assert.Equal(t, nil, bytes)
	assert.Equal(t, nil, errBytes)

	// test unmarshal AttestationInfo model and verify reverse works
	testInfo := &AttestationInfo{}
	_ = bson.Unmarshal(bytes, testInfo)
	assert.Equal(t, testInfo.Txid, info.Txid)
	assert.Equal(t, testInfo.Blockhash, info.Blockhash)
	assert.Equal(t, testInfo.Fee, info.Fee)
	assert.Equal(t, testInfo.Time, info.Time)

	// test AttestationInfo model to document
	doc, docErr := GetDocumentFromModel(testInfo)
	assert.Equal(t, nil, docErr)
	assert.Equal(t, testInfo.Txid, doc.Lookup(ATTESTATION_INFO_TXID_NAME).StringValue())
	assert.Equal(t, testInfo.Blockhash, doc.Lookup(ATTESTATION_INFO_BLOCKHASH_NAME).StringValue())
	assert.Equal(t, testInfo.Fee, doc.Lookup(ATTESTATION_INFO_FEE_NAME).Double())
	assert.Equal(t, testInfo.Time, doc.Lookup(ATTESTATION_INFO_TIME_NAME).Int64())

	// test reverse document to AttestationInfo model
	testtestInfo := &AttestationInfo{}
	docErr = GetModelFromDocument(doc, testtestInfo)
	assert.Equal(t, nil, docErr)
	assert.Equal(t, info.Txid, testtestInfo.Txid)
	assert.Equal(t, info.Blockhash, testtestInfo.Blockhash)
	assert.Equal(t, info.Fee, testtestInfo.Fee)
	assert.Equal(t, info.Time, testtestInfo.Time)
}
