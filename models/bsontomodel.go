// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package models

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

// Function to get bson Document from model interface that implements MarshalBSON
func GetDocumentFromModel(model interface{}) (*bsonx.Doc, error) {

	// model to bytes
	bytes, marshalErr := bson.Marshal(model)
	if marshalErr != nil {
		return nil, marshalErr
	}

	// bytes to bson document
	doc, docErr := bsonx.ReadDoc(bytes)
	if docErr != nil {
		return nil, docErr
	}
	return &doc, nil
}

// Function to get model interface that implements UnmarshalBSON from bson Document
func GetModelFromDocument(doc *bsonx.Doc, model interface{}) error {

	// bson document to bytes
	bytes, errDoc := doc.MarshalBSON()
	if errDoc != nil {
		return errDoc
	}

	// bytes to interface model
	unmarshalErr := bson.Unmarshal(bytes, model)
	if unmarshalErr != nil {
		return unmarshalErr
	}

	return nil
}
