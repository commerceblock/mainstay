package models

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/mongodb/mongo-go-driver/bson"
)

// struct for db LatestCommitment
type LatestCommitment struct {
	Commitment     chainhash.Hash
	ClientPosition int32
}

// Implement bson.Marshaler MarshalBSON() method for use with db_mongo interface
func (c LatestCommitment) MarshalBSON() ([]byte, error) {
	commitmentBSON := LatestCommitmentBSON{c.Commitment.String(), c.ClientPosition}
	return bson.Marshal(commitmentBSON)

}

// Implement bson.Unmarshaler UnmarshalJSON() method for use with db_mongo interface
func (c *LatestCommitment) UnmarshalBSON(b []byte) error {
	var commitmentBSON LatestCommitmentBSON
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
	LATEST_COMMITMENT_CLIENT_POSITION_NAME = "client_position"
	LATEST_COMMITMENT_COMMITMENT_NAME      = "commitment"
)

// LatestCommitmentBSON structure for mongoDB
type LatestCommitmentBSON struct {
	Commitment     string `bson:"commitment"`
	ClientPosition int32  `bson:"client_position"`
}
