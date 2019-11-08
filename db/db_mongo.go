// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"errors"
	"fmt"

	"mainstay/config"
	"mainstay/models"
	"mainstay/log"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

const (
	// collection names
	ColNameAttestation      = "Attestation"
	ColNameAttestationInfo  = "AttestationInfo"
	ColNameMerkleCommitment = "MerkleCommitment"
	ColNameMerkleProof      = "MerkleProof"
	ColNameClientCommitment = "ClientCommitment"
	ColNameClientDetails    = "ClientDetails"

	// error messages
	ErrorMongoClient  = "could not create mongoDB client"
	ErrorMongoConnect = "could not connect to mongoDB client"
	ErrorMongoPing    = "could not ping mongoDB database"

	ErrorAttestationSave      = "could not save attestation"
	ErrorAttestationInfoSave  = "could not save attestation info"
	ErrorMerkleCommitmentSave = "could not save merkle commitment"
	ErrorMerkleProofSave      = "could not save merkle proof"
	ErrorClientDetailsSave    = "could not save client details"
	ErrorClientCommitmentSave = "could not save client commitment"

	ErrorAttestationGet      = "could not get attestation"
	ErrorMerkleCommitmentGet = "could not get merkle commitment"
	ErrorMerkleProofGet      = "could not get merkle proof"
	ErrorClientCommitmentGet = "could not get client commitment"
	ErrorClientDetailsGet    = "could not get client details"

	BadDataClientCommitmentCol = "bad data in client commitment collection"
	BadDataMerkleCommitmentCol = "bad data in merkle commitment collection"
	BadDataClientDetailsCol    = "bad data in client details collection"

	BadDataAttestationModel      = "bad data in attestation model"
	BadDataAttestationInfoModel  = "bad data in attestation info model"
	BadDataMerkleCommitmentModel = "bad data in merkle commitment model"
	BadDataMerkleProofModel      = "bad data in merkle proof model"
	BadDataClientDetailsModel    = "bad data in client details model"
	BadDataClientCommitmentModel = "bad data in client commitment model"
)

// Method to connect to mongo database through config
func dbConnect(ctx context.Context, dbConnectivity config.DbConfig) (*mongo.Database, error) {
	// get this from config
	uri := fmt.Sprintf(`mongodb://%s:%s@%s:%s/%s?connect=direct`,
		dbConnectivity.User,
		dbConnectivity.Password,
		dbConnectivity.Host,
		dbConnectivity.Port,
		dbConnectivity.Name,
	)

	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s %v", ErrorMongoClient, err))
	}

	err = client.Connect(ctx) // start background client routine
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s %v", ErrorMongoConnect, err))
	}

	err = client.Ping(ctx, nil) // use Ping to check if mongod is running
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s %v", ErrorMongoPing, err))
	}

	return client.Database(dbConnectivity.Name), nil
}

// DbMongo struct
type DbMongo struct {
	// context required by mongo interface
	ctx context.Context

	// database connectivity config
	dbConnectivity config.DbConfig

	// mongo interface connection
	db *mongo.Database
}

// Return new DbMongo instance
func NewDbMongo(ctx context.Context, dbConnectivity config.DbConfig) *DbMongo {
	db, errConnect := dbConnect(ctx, dbConnectivity)
	if errConnect != nil {
		log.Error(errConnect)
	}

	return &DbMongo{ctx, dbConnectivity, db}
}

// Save latest attestation to the Attestation collection
func (d *DbMongo) SaveAttestation(attestation models.Attestation) error {

	// get document representation of Attestation object
	docAttestation, docErr := models.GetDocumentFromModel(attestation)
	if docErr != nil {
		return errors.New(fmt.Sprintf("%s %v", BadDataAttestationModel, docErr))
	}

	newAttestation := bsonx.Doc{
		{"$set", bsonx.Document(*docAttestation)},
	}

	// search if attestation already exists
	filterAttestation := bsonx.Doc{
		{models.AttestationTxidName, bsonx.String(docAttestation.Lookup(models.AttestationTxidName).StringValue())},
		{models.AttestationMerkleRootName, bsonx.String(docAttestation.Lookup(models.AttestationMerkleRootName).StringValue())},
	}

	// insert or update attestation
	var t bsonx.Doc
	opts := &options.FindOneAndUpdateOptions{}
	opts.SetUpsert(true)
	res := d.db.Collection(ColNameAttestation).FindOneAndUpdate(d.ctx, filterAttestation, newAttestation, opts)
	resErr := res.Decode(&t)
	if resErr != nil && resErr != mongo.ErrNoDocuments {
		return errors.New(fmt.Sprintf("%s %v", ErrorAttestationSave, resErr))
	}

	return nil
}

// Save latest attestation info to the Attestation info collection
func (d *DbMongo) SaveAttestationInfo(attestationInfo models.AttestationInfo) error {

	// get document representation of AttestationInfo object
	docAttestationInfo, docErr := models.GetDocumentFromModel(attestationInfo)
	if docErr != nil {
		return errors.New(fmt.Sprintf("%s %v", BadDataAttestationInfoModel, docErr))
	}
	newAttestationInfo := bsonx.Doc{
		{"$set", bsonx.Document(*docAttestationInfo)},
	}

	// search if attestationInfo already exists
	filterAttestationInfo := bsonx.Doc{
		{models.AttestationInfoTxidName, bsonx.String(docAttestationInfo.Lookup(models.AttestationInfoTxidName).StringValue())},
	}

	// insert or update attestationInfo
	var t bsonx.Doc
	opts := &options.FindOneAndUpdateOptions{}
	opts.SetUpsert(true)
	res := d.db.Collection(ColNameAttestationInfo).FindOneAndUpdate(d.ctx, filterAttestationInfo, newAttestationInfo, opts)
	resErr := res.Decode(&t)
	if resErr != nil && resErr != mongo.ErrNoDocuments {
		return errors.New(fmt.Sprintf("%s %v", ErrorAttestationInfoSave, resErr))
	}

	return nil
}

// Save merkle commitments to the MerkleCommitment collection
func (d *DbMongo) SaveMerkleCommitments(commitments []models.CommitmentMerkleCommitment) error {
	for pos := range commitments {
		// get document representation of each commitment
		// get document representation of Attestation object
		docCommitment, docErr := models.GetDocumentFromModel(commitments[pos])
		if docErr != nil {
			return errors.New(fmt.Sprintf("%s %v", BadDataMerkleCommitmentModel, docErr))
		}

		newCommitment := bsonx.Doc{
			{"$set", bsonx.Document(*docCommitment)},
		}

		// search if merkle commitment already exists
		filterMerkleCommitment := bsonx.Doc{
			{models.CommitmentMerkleRootName,
				bsonx.String(docCommitment.Lookup(models.CommitmentMerkleRootName).StringValue())},
			{models.CommitmentClientPositionName,
				bsonx.Int32(docCommitment.Lookup(models.CommitmentClientPositionName).Int32())},
		}

		// insert or update merkle commitment
		var t bsonx.Doc
		opts := &options.FindOneAndUpdateOptions{}
		opts.SetUpsert(true)
		res := d.db.Collection(ColNameMerkleCommitment).FindOneAndUpdate(d.ctx, filterMerkleCommitment, newCommitment, opts)
		resErr := res.Decode(&t)
		if resErr != nil && resErr != mongo.ErrNoDocuments {
			return errors.New(fmt.Sprintf("%s %v", ErrorMerkleCommitmentSave, resErr))
		}
	}
	return nil
}

// Save merkle proofs to the MerkleProof collection
func (d *DbMongo) SaveMerkleProofs(proofs []models.CommitmentMerkleProof) error {
	for pos := range proofs {
		// get document representation of merkle proof
		docProof, docErr := models.GetDocumentFromModel(proofs[pos])
		if docErr != nil {
			return errors.New(fmt.Sprintf("%s %v", BadDataMerkleProofModel, docErr))
		}

		newProof := bsonx.Doc{
			{"$set", bsonx.Document(*docProof)},
		}

		// search if merkle proof already exists
		filterMerkleProof := bsonx.Doc{
			{models.ProofMerkleRootName,
				bsonx.String(docProof.Lookup(models.ProofMerkleRootName).StringValue())},
			{models.ProofClientPositionName,
				bsonx.Int32(docProof.Lookup(models.ProofClientPositionName).Int32())},
		}

		// insert or update merkle proof
		var t bsonx.Doc
		opts := &options.FindOneAndUpdateOptions{}
		opts.SetUpsert(true)
		res := d.db.Collection(ColNameMerkleProof).FindOneAndUpdate(d.ctx, filterMerkleProof, newProof, opts)
		resErr := res.Decode(&t)
		if resErr != nil && resErr != mongo.ErrNoDocuments {
			return errors.New(fmt.Sprintf("%s %v", ErrorMerkleProofSave, resErr))
		}
	}
	return nil
}

// Save client details to ClientDetails collection
func (d *DbMongo) SaveClientDetails(details models.ClientDetails) error {
	// get document representation of client details
	docDetails, docErr := models.GetDocumentFromModel(details)
	if docErr != nil {
		return errors.New(fmt.Sprintf("%s %v", BadDataClientDetailsModel, docErr))
	}

	newDetails := bsonx.Doc{
		{"$set", bsonx.Document(*docDetails)},
	}

	// search if client details for position already exists
	filterClientDetails := bsonx.Doc{
		{models.ClientDetailsClientPositionName,
			bsonx.Int32(docDetails.Lookup(models.ClientDetailsClientPositionName).Int32())},
	}

	// insert or update client details
	var t bsonx.Doc
	opts := &options.FindOneAndUpdateOptions{}
	opts.SetUpsert(true)
	res := d.db.Collection(ColNameClientDetails).FindOneAndUpdate(d.ctx, filterClientDetails, newDetails, opts)
	resErr := res.Decode(&t)
	if resErr != nil && resErr != mongo.ErrNoDocuments {
		return errors.New(fmt.Sprintf("%s %v", ErrorClientDetailsSave, resErr))
	}
	return nil
}

// Save client commitment to ClientCommitment collection
func (d *DbMongo) SaveClientCommitment(commitment models.ClientCommitment) error {
	// get document representation of client details
	docCommitment, docErr := models.GetDocumentFromModel(commitment)
	if docErr != nil {
		return errors.New(fmt.Sprintf("%s %v", BadDataClientCommitmentModel, docErr))
	}

	newCommitment := bsonx.Doc{
		{"$set", bsonx.Document(*docCommitment)},
	}

	// search if client details for position already exists
	filterClientCommitment := bsonx.Doc{
		{models.ClientCommitmentClientPositionName,
			bsonx.Int32(docCommitment.Lookup(models.ClientCommitmentClientPositionName).Int32())},
	}

	// insert or update client details
	var t bsonx.Doc
	opts := &options.FindOneAndUpdateOptions{}
	opts.SetUpsert(true)
	res := d.db.Collection(ColNameClientCommitment).FindOneAndUpdate(d.ctx, filterClientCommitment, newCommitment, opts)
	resErr := res.Decode(&t)
	if resErr != nil && resErr != mongo.ErrNoDocuments {
		return errors.New(fmt.Sprintf("%s %v", ErrorClientCommitmentSave, resErr))
	}
	return nil
}

// Get latest ClientDetails document
func (d *DbMongo) GetClientDetails() ([]models.ClientDetails, error) {
	// sort by client position
	sortFilter := bsonx.Doc{{models.ClientDetailsClientPositionName, bsonx.Int32(1)}}
	res, resErr := d.db.Collection(ColNameClientDetails).Find(d.ctx, bsonx.Doc{}, &options.FindOptions{Sort: sortFilter})
	if resErr != nil {
		return []models.ClientDetails{},
			errors.New(fmt.Sprintf("%s %v", ErrorClientDetailsGet, resErr))
	}

	// iterate through details
	var details []models.ClientDetails
	for res.Next(d.ctx) {
		var detailsDoc bsonx.Doc
		if err := res.Decode(&detailsDoc); err != nil {
			return []models.ClientDetails{},
				errors.New(fmt.Sprintf("%s %v", BadDataClientDetailsCol, err))
		}
		detailsModel := &models.ClientDetails{}
		modelErr := models.GetModelFromDocument(&detailsDoc, detailsModel)
		if modelErr != nil {
			return []models.ClientDetails{}, errors.New(fmt.Sprintf("%s %v", BadDataClientDetailsCol, modelErr))
		}
		details = append(details, *detailsModel)
	}
	if err := res.Err(); err != nil {
		return []models.ClientDetails{}, errors.New(fmt.Sprintf("%s %v", BadDataClientDetailsCol, err))
	}
	return details, nil
}

// Get Attestation collection document count
func (d *DbMongo) getAttestationCount(confirmed ...bool) (int64, error) {
	// set optional confirmed filter
	confirmedFilter := bsonx.Doc{}
	if len(confirmed) > 0 {
		confirmedFilter = bsonx.Doc{{models.AttestationConfirmedName, bsonx.Boolean(confirmed[0])}}
	}
	// find latest attestation count
	opts := options.CountOptions{}
	opts.SetLimit(1)
	count, countErr := d.db.Collection(ColNameAttestation).CountDocuments(d.ctx, confirmedFilter, &opts)
	if countErr != nil {
		return 0, errors.New(fmt.Sprintf("%s %v", ErrorAttestationGet, countErr))
	}

	return count, nil
}

// Get Attestation entry from collection and return merkle_root field
func (d *DbMongo) GetLatestAttestationMerkleRoot(confirmed bool) (string, error) {
	// first check if attestation has any documents
	count, countErr := d.getAttestationCount(confirmed)
	if countErr != nil {
		return "", countErr
	} else if count == 0 { // no attestations yet
		return "", nil
	}

	// filter by inserted date and confirmed to get latest attestation from Attestation collection
	sortFilter := bsonx.Doc{{models.AttestationInsertedAtName, bsonx.Int32(-1)}}
	confirmedFilter := bsonx.Doc{{models.AttestationConfirmedName, bsonx.Boolean(confirmed)}}

	var attestationDoc bsonx.Doc
	resErr := d.db.Collection(ColNameAttestation).FindOne(d.ctx,
		confirmedFilter, &options.FindOneOptions{Sort: sortFilter}).Decode(&attestationDoc)
	if resErr != nil {
		return "", errors.New(fmt.Sprintf("%s %v", ErrorAttestationGet, resErr))
	}
	return attestationDoc.Lookup(models.AttestationMerkleRootName).StringValue(), nil
}

// Return Commitment from MerkleCommitment commitments for attestation with given txid hash
func (d *DbMongo) getAttestationMerkleRoot(txid chainhash.Hash) (string, error) {
	// first check if attestation has any documents
	count, countErr := d.getAttestationCount()
	if countErr != nil {
		return "", countErr
	} else if count == 0 { // no attestations yet
		return "", nil
	}

	// get merke_root from Attestation collection for attestation txid provided
	filterAttestation := bsonx.Doc{
		{models.AttestationTxidName, bsonx.String(txid.String())},
	}

	var attestationDoc bsonx.Doc
	resErr := d.db.Collection(ColNameAttestation).FindOne(d.ctx, filterAttestation).Decode(&attestationDoc)
	if resErr != nil {
		if resErr == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", errors.New(fmt.Sprintf("%s %v", ErrorAttestationGet, resErr))
	}
	return attestationDoc.Lookup(models.CommitmentMerkleRootName).StringValue(), nil
}

// Return Commitment from MerkleCommitment commitments for attestation with given txid hash
func (d *DbMongo) GetAttestationMerkleCommitments(txid chainhash.Hash) ([]models.CommitmentMerkleCommitment, error) {
	// get merkle root of attestation
	merkleRoot, rootErr := d.getAttestationMerkleRoot(txid)
	if rootErr != nil {
		return []models.CommitmentMerkleCommitment{}, rootErr
	} else if merkleRoot == "" {
		return []models.CommitmentMerkleCommitment{}, nil
	}

	// filter MerkleCommitment collection by merkle_root and sort for client position
	sortFilter := bsonx.Doc{{models.CommitmentClientPositionName, bsonx.Int32(1)}}
	filterMerkleRoot := bsonx.Doc{{models.CommitmentMerkleRootName, bsonx.String(merkleRoot)}}
	res, resErr := d.db.Collection(ColNameMerkleCommitment).Find(d.ctx, filterMerkleRoot, &options.FindOptions{Sort: sortFilter})
	if resErr != nil {
		return []models.CommitmentMerkleCommitment{},
			errors.New(fmt.Sprintf("%s %v", ErrorMerkleCommitmentGet, resErr))
	}

	// fetch commitments
	var merkleCommitments []models.CommitmentMerkleCommitment
	for res.Next(d.ctx) {
		var commitmentDoc bsonx.Doc
		if err := res.Decode(&commitmentDoc); err != nil {
			log.Infof("%s\n", BadDataMerkleCommitmentCol)
			return []models.CommitmentMerkleCommitment{}, err
		}
		// decode document result to Commitment model and get hash
		commitmentModel := &models.CommitmentMerkleCommitment{}
		modelErr := models.GetModelFromDocument(&commitmentDoc, commitmentModel)
		if modelErr != nil {
			log.Infof("%s\n", BadDataMerkleCommitmentCol)
			return []models.CommitmentMerkleCommitment{}, modelErr
		}
		merkleCommitments = append(merkleCommitments, *commitmentModel)
	}
	if err := res.Err(); err != nil {
		return []models.CommitmentMerkleCommitment{},
			errors.New(fmt.Sprintf("%s %v", BadDataMerkleCommitmentCol, err))
	}
	return merkleCommitments, nil
}

// Return latest commitments from MerkleCommitment collection
func (d *DbMongo) GetClientCommitments() ([]models.ClientCommitment, error) {

	// sort by client position to get correct commitment order
	sortFilter := bsonx.Doc{{models.ClientCommitmentClientPositionName, bsonx.Int32(1)}}
	res, resErr := d.db.Collection(ColNameClientCommitment).Find(d.ctx, bsonx.Doc{}, &options.FindOptions{Sort: sortFilter})
	if resErr != nil {
		return []models.ClientCommitment{},
			errors.New(fmt.Sprintf("%s %v", ErrorClientCommitmentGet, resErr))
	}

	// iterate through commitments
	var latestCommitments []models.ClientCommitment
	for res.Next(d.ctx) {
		var commitmentDoc bsonx.Doc
		if err := res.Decode(&commitmentDoc); err != nil {
			return []models.ClientCommitment{},
				errors.New(fmt.Sprintf("%s %v", BadDataClientCommitmentCol, err))
		}
		commitmentModel := &models.ClientCommitment{}
		modelErr := models.GetModelFromDocument(&commitmentDoc, commitmentModel)
		if modelErr != nil {
			return []models.ClientCommitment{}, errors.New(fmt.Sprintf("%s %v", BadDataClientCommitmentCol, modelErr))
		}
		latestCommitments = append(latestCommitments, *commitmentModel)
	}
	if err := res.Err(); err != nil {
		return []models.ClientCommitment{}, errors.New(fmt.Sprintf("%s %v", BadDataClientCommitmentCol, err))
	}
	return latestCommitments, nil
}
