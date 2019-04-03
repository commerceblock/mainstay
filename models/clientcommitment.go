// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package models

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"go.mongodb.org/mongo-driver/bson"
)

// struct for db ClientCommitment
type ClientCommitment struct {
	Commitment     chainhash.Hash
	ClientPosition int32
}

// Implement bson.Marshaler MarshalBSON() method for use with db_mongo interface
func (c ClientCommitment) MarshalBSON() ([]byte, error) {
	commitmentBSON := ClientCommitmentBSON{c.Commitment.String(), c.ClientPosition}
	return bson.Marshal(commitmentBSON)

}

// Implement bson.Unmarshaler UnmarshalJSON() method for use with db_mongo interface
func (c *ClientCommitment) UnmarshalBSON(b []byte) error {
	var commitmentBSON ClientCommitmentBSON
	if err := bson.Unmarshal(b, &commitmentBSON); err != nil {
		return err
	}
	commitmentHash, errHash := chainhash.NewHashFromStr(commitmentBSON.Commitment)
	if errHash != nil {
		return errHash
	}
	c.ClientPosition = commitmentBSON.ClientPosition
	c.Commitment = *commitmentHash
	return nil
}

// Commitment field names
const (
	ClientCommitmentClientPositionName = "client_position"
	ClientCommitmentCommitmentName     = "commitment"
)

// ClientCommitmentBSON structure for mongoDB
type ClientCommitmentBSON struct {
	Commitment     string `bson:"commitment"`
	ClientPosition int32  `bson:"client_position"`
}
