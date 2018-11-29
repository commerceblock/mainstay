// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Dummy struct {
	Name     string `bson:"name"`
	Verified bool   `bson:"verified"`
}

// Test BSON Model utils
func TestBsonToModel(t *testing.T) {
	dummyVal := Dummy{"Nick", true}

	// test model to document
	doc, docErr := GetDocumentFromModel(dummyVal)
	assert.Equal(t, nil, docErr)
	assert.Equal(t, dummyVal.Name, doc.Lookup("name").StringValue())
	assert.Equal(t, dummyVal.Verified, doc.Lookup("verified").Boolean())

	// test document to model
	testDummy := &Dummy{}
	docErr = GetModelFromDocument(doc, testDummy)
	assert.Equal(t, nil, docErr)
	assert.Equal(t, dummyVal.Name, testDummy.Name)
	assert.Equal(t, dummyVal.Verified, testDummy.Verified)
}
