// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package models

import (
	"testing"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/stretchr/testify/assert"
)

// Test ClientDetails high level interface
func TestClientDetails(t *testing.T) {
	clientDetails := ClientDetails{0, "04ddb0d6-ed74-4cc6-b9dc-72f2a809525b", "03e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b33"}
	assert.Equal(t, int32(0), clientDetails.ClientPosition)
	assert.Equal(t, "04ddb0d6-ed74-4cc6-b9dc-72f2a809525b", clientDetails.AuthToken)
	assert.Equal(t, "03e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b33", clientDetails.Pubkey)
}

// Test ClientDetails BSON interface
func TestClientDetailsBSON(t *testing.T) {
	clientDetails := ClientDetails{0, "04ddb0d6-ed74-4cc6-b9dc-72f2a809525b", "03e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b33"}

	// test marshal clientDetails model
	bytes, errBytes := bson.Marshal(clientDetails)
	assert.Equal(t, []uint8([]byte{0x9e, 0x0, 0x0, 0x0, 0x10, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x61, 0x75, 0x74, 0x68, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x0, 0x25, 0x0, 0x0, 0x0, 0x30, 0x34, 0x64, 0x64, 0x62, 0x30, 0x64, 0x36, 0x2d, 0x65, 0x64, 0x37, 0x34, 0x2d, 0x34, 0x63, 0x63, 0x36, 0x2d, 0x62, 0x39, 0x64, 0x63, 0x2d, 0x37, 0x32, 0x66, 0x32, 0x61, 0x38, 0x30, 0x39, 0x35, 0x32, 0x35, 0x62, 0x0, 0x2, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x0, 0x43, 0x0, 0x0, 0x0, 0x30, 0x33, 0x65, 0x35, 0x32, 0x63, 0x66, 0x31, 0x35, 0x65, 0x30, 0x61, 0x35, 0x63, 0x66, 0x36, 0x36, 0x31, 0x32, 0x33, 0x31, 0x34, 0x66, 0x30, 0x37, 0x37, 0x62, 0x62, 0x36, 0x35, 0x63, 0x66, 0x39, 0x61, 0x36, 0x35, 0x39, 0x36, 0x62, 0x37, 0x36, 0x63, 0x30, 0x66, 0x63, 0x62, 0x33, 0x34, 0x62, 0x36, 0x38, 0x32, 0x66, 0x36, 0x37, 0x33, 0x61, 0x38, 0x33, 0x31, 0x34, 0x63, 0x37, 0x62, 0x33, 0x33, 0x0, 0x0}), bytes)
	assert.Equal(t, nil, errBytes)

	// test unmarshal clientDetails model and verify reverse works
	testClientDetails := &ClientDetails{}
	_ = bson.Unmarshal(bytes, testClientDetails)
	assert.Equal(t, clientDetails.AuthToken, testClientDetails.AuthToken)
	assert.Equal(t, clientDetails.Pubkey, testClientDetails.Pubkey)
	assert.Equal(t, clientDetails.ClientPosition, testClientDetails.ClientPosition)

	// test clientDetails model to document
	doc, docErr := GetDocumentFromModel(testClientDetails)
	assert.Equal(t, nil, docErr)
	assert.Equal(t, clientDetails.AuthToken, doc.Lookup(CLIENT_DETAILS_AUTH_TOKEN_NAME).StringValue())
	assert.Equal(t, clientDetails.Pubkey, doc.Lookup(CLIENT_DETAILS_PUBKEY_NAME).StringValue())
	assert.Equal(t, clientDetails.ClientPosition, doc.Lookup(CLIENT_COMMITMENT_CLIENT_POSITION_NAME).Int32())

	// test reverse document to clientDetails model
	testtestClientDetails := &ClientDetails{}
	docErr = GetModelFromDocument(doc, testtestClientDetails)
	assert.Equal(t, nil, docErr)
	assert.Equal(t, clientDetails.AuthToken, testtestClientDetails.AuthToken)
	assert.Equal(t, clientDetails.Pubkey, testtestClientDetails.Pubkey)
	assert.Equal(t, clientDetails.ClientPosition, testtestClientDetails.ClientPosition)
}
