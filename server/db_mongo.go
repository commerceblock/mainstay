package server

import (
	"context"
	"fmt"
	"os"
	"time"

	"mainstay/models"

	"github.com/mongodb/mongo-go-driver/bson"
	_ "github.com/mongodb/mongo-go-driver/bson/objectid"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/findopt"
)

// Method to connect to mongo database through config
func dbConnect(ctx context.Context) (*mongo.Database, error) {
	// get this from config
	uri := fmt.Sprintf(`mongodb://%s:%s@%s:%s/%s`,
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME_MAINSTAY"),
	)

	client, err := mongo.NewClient(uri)
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return client.Database(os.Getenv("DB_NAME_MAINSTAY")), nil
}

// DbMongo struct
type DbMongo struct {
	ctx context.Context
	db  *mongo.Database
}

// Return new DbMongo instance
func NewDbMongo(ctx context.Context) (DbMongo, error) {
	db, errConnect := dbConnect(ctx)

	if errConnect != nil {
		return DbMongo{}, errConnect
	}

	return DbMongo{ctx, db}, nil
}

// Save latest attestation to the database. If attestation already exists then update
func (d *DbMongo) saveAttestation(attestation models.Attestation, confirmed bool) error {

	// new attestation based on Attestation model
	newAttestation := bson.NewDocument(
		bson.EC.SubDocumentFromElements("$set", bson.EC.String("txid", attestation.Txid.String())),
		bson.EC.SubDocumentFromElements("$set", bson.EC.String("merkle_root", attestation.CommitmentHash().String())),
		bson.EC.SubDocumentFromElements("$set", bson.EC.DateTime("inserted_at", int64(time.Now().Unix())*1000)),
		bson.EC.SubDocumentFromElements("$set", bson.EC.Boolean("confirmed", confirmed)),
	)

	// search if attestation already exists
	filterAttestation := bson.NewDocument(
		bson.EC.String("txid", attestation.Txid.String()),
		bson.EC.String("merkle_root", attestation.CommitmentHash().String()),
	)

	// insert or update
	t := bson.NewDocument()
	res := d.db.Collection("Attestation").FindOneAndUpdate(d.ctx, filterAttestation, newAttestation, findopt.Upsert(true))
	resErr := res.Decode(t)
	if resErr != nil && resErr != mongo.ErrNoDocuments {
		fmt.Printf("couldn't be created: %v\n", resErr)
		return resErr
	}

	return nil
}

// Save merkle commitments of attestation to the database
func (d *DbMongo) saveMerkleCommitment(commitment models.Commitment) error {
	return nil
}

// Save merkle proofs of attestation to the database
func (d *DbMongo) saveMerkleProof(commitment models.Commitment) error {
	return nil
}

// Return latest from the database
func (d *DbMongo) getLatestAttestation() models.Attestation {
	return models.Attestation{}
}
