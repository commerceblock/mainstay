package server

import (
	"context"
	"errors"
	"fmt"
	"log"

	"mainstay/config"
	"mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/options"
)

const (
	// collection names
	COL_NAME_ATTESTATION       = "Attestation"
	COL_NAME_MERKLE_COMMITMENT = "MerkleCommitment"
	COL_NAME_MERKLE_PROOF      = "MerkleProof"
	COL_NAME_LATEST_COMMITMENT = "LatestCommitment"

	// error messages
	ERROR_MONGO_CLIENT  = "could not create mongoDB client"
	ERROR_MONGO_CONNECT = "could not connect to mongoDB client"
	ERROR_MONGO_PING    = "could not ping mongoDB database"

	ERROR_ATTESTATION_SAVE       = "could not save attestation"
	ERROR_MERKLE_COMMITMENT_SAVE = "could not save merkle commitment"
	ERROR_MERKLE_PROOF_SAVE      = "could not save merkle proof"

	ERROR_ATTESTATION_GET       = "could not get attestation"
	ERROR_MERKLE_COMMITMENT_GET = "could not get merkle commitment"
	ERROR_MERKLE_PROOF_GET      = "could not get merkle proof"
	ERROR_LATEST_COMMITMENT_GET = "could not get latest commitment"

	BAD_DATA_LATEST_COMMITMENT_COL = "bad data in latest commitment collection"
	BAD_DATA_MERKLE_COMMITMENT_COL = "bad data in merkle commitment collection"

	BAD_DATA_ATTESTATION_MODEL       = "bad data in attestation model"
	BAD_DATA_MERKLE_COMMITMENT_MODEL = "bad data in merkle commitment model"
	BAD_DATA_MERKLE_PROOF_MODEL      = "bad data in merkle proof model"
)

// Method to connect to mongo database through config
func dbConnect(ctx context.Context, dbConnectivity config.DbConnectivity) (*mongo.Database, error) {
	// get this from config
	uri := fmt.Sprintf(`mongodb://%s:%s@%s:%s/%s`,
		dbConnectivity.User,
		dbConnectivity.Password,
		dbConnectivity.Host,
		dbConnectivity.Port,
		dbConnectivity.Name,
	)

	client, err := mongo.NewClient(uri)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s %v", ERROR_MONGO_CLIENT, err))
	}

	err = client.Connect(ctx) // start background client routine
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s %v", ERROR_MONGO_CONNECT, err))
	}

	err = client.Ping(ctx, nil) // use Ping to check if mongod is running
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s %v", ERROR_MONGO_PING, err))
	}

	return client.Database(dbConnectivity.Name), nil
}

// DbMongo struct
type DbMongo struct {
	ctx            context.Context
	dbConnectivity config.DbConnectivity
	db             *mongo.Database
}

// Return new DbMongo instance
func NewDbMongo(ctx context.Context, dbConnectivity config.DbConnectivity) *DbMongo {
	db, errConnect := dbConnect(ctx, dbConnectivity)
	if errConnect != nil {
		log.Fatal(errConnect)
	}

	return &DbMongo{ctx, dbConnectivity, db}
}

// Save latest attestation to the Attestation collection
func (d *DbMongo) saveAttestation(attestation models.Attestation) error {

	// get document representation of Attestation object
	docAttestation, docErr := models.GetDocumentFromModel(attestation)
	if docErr != nil {
		return errors.New(fmt.Sprintf("%s %v", BAD_DATA_ATTESTATION_MODEL, docErr))
	}

	newAttestation := bson.NewDocument(
		bson.EC.SubDocument("$set", docAttestation),
	)

	// search if attestation already exists
	filterAttestation := bson.NewDocument(
		bson.EC.String(models.ATTESTATION_TXID_NAME, docAttestation.Lookup(models.ATTESTATION_TXID_NAME).StringValue()),
		bson.EC.String(models.ATTESTATION_MERKLE_ROOT_NAME, docAttestation.Lookup(models.ATTESTATION_MERKLE_ROOT_NAME).StringValue()),
	)

	// insert or update attestation
	t := bson.NewDocument()
	opts := &options.FindOneAndUpdateOptions{}
	opts.SetUpsert(true)
	res := d.db.Collection(COL_NAME_ATTESTATION).FindOneAndUpdate(d.ctx, filterAttestation, newAttestation, opts)
	resErr := res.Decode(t)
	if resErr != nil && resErr != mongo.ErrNoDocuments {
		return errors.New(fmt.Sprintf("%s %v", ERROR_ATTESTATION_SAVE, resErr))
	}

	return nil
}

// Save merkle commitments to the MerkleCommitment collection
func (d *DbMongo) saveMerkleCommitments(commitments []models.CommitmentMerkleCommitment) error {
	for pos := range commitments {
		// get document representation of each commitment
		// get document representation of Attestation object
		docCommitment, docErr := models.GetDocumentFromModel(commitments[pos])
		if docErr != nil {
			return errors.New(fmt.Sprintf("%s %v", BAD_DATA_MERKLE_COMMITMENT_MODEL, docErr))
		}

		newCommitment := bson.NewDocument(
			bson.EC.SubDocument("$set", docCommitment),
		)

		// search if merkle commitment already exists
		filterMerkleCommitment := bson.NewDocument(
			bson.EC.String(models.COMMITMENT_MERKLE_ROOT_NAME,
				docCommitment.Lookup(models.COMMITMENT_MERKLE_ROOT_NAME).StringValue()),
			bson.EC.Int32(models.COMMITMENT_CLIENT_POSITION_NAME,
				docCommitment.Lookup(models.COMMITMENT_CLIENT_POSITION_NAME).Int32()),
		)

		// insert or update merkle commitment
		t := bson.NewDocument()
		opts := &options.FindOneAndUpdateOptions{}
		opts.SetUpsert(true)
		res := d.db.Collection(COL_NAME_MERKLE_COMMITMENT).FindOneAndUpdate(d.ctx, filterMerkleCommitment, newCommitment, opts)
		resErr := res.Decode(t)
		if resErr != nil && resErr != mongo.ErrNoDocuments {
			return errors.New(fmt.Sprintf("%s %v", ERROR_MERKLE_COMMITMENT_SAVE, resErr))
		}
	}
	return nil
}

// Save merkle proofs to the MerkleProof collection
func (d *DbMongo) saveMerkleProofs(proofs []models.CommitmentMerkleProof) error {
	for pos := range proofs {
		// get document representation of merkle proof
		docProof, docErr := models.GetDocumentFromModel(proofs[pos])
		if docErr != nil {
			return errors.New(fmt.Sprintf("%s %v", BAD_DATA_MERKLE_PROOF_MODEL, docErr))
		}

		newProof := bson.NewDocument(
			bson.EC.SubDocument("$set", docProof),
		)

		// search if merkle proof already exists
		filterMerkleProof := bson.NewDocument(
			bson.EC.String(models.PROOF_MERKLE_ROOT_NAME,
				docProof.Lookup(models.PROOF_MERKLE_ROOT_NAME).StringValue()),
			bson.EC.Int32(models.PROOF_CLIENT_POSITION_NAME,
				docProof.Lookup(models.PROOF_CLIENT_POSITION_NAME).Int32()),
		)

		// insert or update merkle proof
		t := bson.NewDocument()
		opts := &options.FindOneAndUpdateOptions{}
		opts.SetUpsert(true)
		res := d.db.Collection(COL_NAME_MERKLE_PROOF).FindOneAndUpdate(d.ctx, filterMerkleProof, newProof, opts)
		resErr := res.Decode(t)
		if resErr != nil && resErr != mongo.ErrNoDocuments {
			return errors.New(fmt.Sprintf("%s %v", ERROR_MERKLE_PROOF_SAVE, resErr))
		}
	}
	return nil
}

// Get Attestation collection document count
func (d *DbMongo) getLatestAttestationCount() (int64, error) {
	// find latest attestation count
	opts := options.CountOptions{}
	opts.SetLimit(1)
	count, countErr := d.db.Collection(COL_NAME_ATTESTATION).Count(d.ctx, nil, &opts)
	if countErr != nil {
		return 0, errors.New(fmt.Sprintf("%s %v", ERROR_ATTESTATION_GET, countErr))
	}

	return count, nil
}

// Get Attestation entry from collection and return merkle_root field
func (d *DbMongo) getLatestAttestationMerkleRoot() (string, error) {
	// first check if attestation has any documents
	count, countErr := d.getLatestAttestationCount()
	if countErr != nil {
		return "", countErr
	} else if count == 0 { // no attestations yet
		return "", nil
	}

	// filter by inserted date and confirmed to get latest attestation from Attestation collection
	sortFilter := bson.NewDocument(bson.EC.Int32(models.ATTESTATION_INSERTED_AT_NAME, -1))
	confirmedFilter := bson.NewDocument(bson.EC.Boolean(models.ATTESTATION_CONFIRMED_NAME, true))

	attestationDoc := bson.NewDocument()
	resErr := d.db.Collection(COL_NAME_ATTESTATION).FindOne(d.ctx,
		confirmedFilter, &options.FindOneOptions{Sort: sortFilter}).Decode(attestationDoc)
	if resErr != nil {
		return "", errors.New(fmt.Sprintf("%s %v", ERROR_ATTESTATION_GET, resErr))
	}
	return attestationDoc.Lookup(models.ATTESTATION_MERKLE_ROOT_NAME).StringValue(), nil
}

// Return Commitment from MerkleCommitment commitments for attestation with given txid hash
func (d *DbMongo) getAttestationMerkleRoot(txid chainhash.Hash) (string, error) {
	// first check if attestation has any documents
	count, countErr := d.getLatestAttestationCount()
	if countErr != nil {
		return "", countErr
	} else if count == 0 { // no attestations yet
		return "", nil
	}

	// get merke_root from Attestation collection for attestation txid provided
	filterAttestation := bson.NewDocument(bson.EC.String(models.ATTESTATION_TXID_NAME, txid.String()))

	attestationDoc := bson.NewDocument()
	resErr := d.db.Collection(COL_NAME_ATTESTATION).FindOne(d.ctx, filterAttestation).Decode(attestationDoc)
	if resErr != nil {
		return "", errors.New(fmt.Sprintf("%s %v", ERROR_ATTESTATION_GET, resErr))
	}
	return attestationDoc.Lookup(models.COMMITMENT_MERKLE_ROOT_NAME).StringValue(), nil
}

// Return Commitment from MerkleCommitment commitments for attestation with given txid hash
func (d *DbMongo) getAttestationMerkleCommitments(txid chainhash.Hash) ([]models.CommitmentMerkleCommitment, error) {
	// get merkle root of attestation
	merkleRoot, rootErr := d.getAttestationMerkleRoot(txid)
	if rootErr != nil {
		return []models.CommitmentMerkleCommitment{}, rootErr
	} else if merkleRoot == "" {
		return []models.CommitmentMerkleCommitment{}, nil
	}

	// filter MerkleCommitment collection by merkle_root and sort for client position
	sortFilter := bson.NewDocument(bson.EC.Int32(models.COMMITMENT_CLIENT_POSITION_NAME, 1))
	filterMerkleRoot := bson.NewDocument(bson.EC.String(models.COMMITMENT_MERKLE_ROOT_NAME, merkleRoot))
	res, resErr := d.db.Collection(COL_NAME_MERKLE_COMMITMENT).Find(d.ctx, filterMerkleRoot, &options.FindOptions{Sort: sortFilter})
	if resErr != nil {
		return []models.CommitmentMerkleCommitment{},
			errors.New(fmt.Sprintf("%s %v", ERROR_MERKLE_COMMITMENT_GET, resErr))
	}

	// fetch commitments
	var merkleCommitments []models.CommitmentMerkleCommitment
	for res.Next(d.ctx) {
		commitmentDoc := bson.NewDocument()
		if err := res.Decode(commitmentDoc); err != nil {
			fmt.Printf("%s\n", BAD_DATA_MERKLE_COMMITMENT_COL)
			return []models.CommitmentMerkleCommitment{}, err
		}
		// decode document result to Commitment model and get hash
		commitmentModel := &models.CommitmentMerkleCommitment{}
		modelErr := models.GetModelFromDocument(commitmentDoc, commitmentModel)
		if modelErr != nil {
			fmt.Printf("%s\n", BAD_DATA_MERKLE_COMMITMENT_COL)
			return []models.CommitmentMerkleCommitment{}, modelErr
		}
		merkleCommitments = append(merkleCommitments, *commitmentModel)
	}
	if err := res.Err(); err != nil {
		return []models.CommitmentMerkleCommitment{},
			errors.New(fmt.Sprintf("%s %v", BAD_DATA_MERKLE_COMMITMENT_COL, err))
	}
	return merkleCommitments, nil
}

// Return latest commitments from MerkleCommitment collection
func (d *DbMongo) getLatestCommitments() ([]models.LatestCommitment, error) {

	// sort by client position to get correct commitment order
	sortFilter := bson.NewDocument(bson.EC.Int32(models.LATEST_COMMITMENT_CLIENT_POSITION_NAME, 1))
	res, resErr := d.db.Collection(COL_NAME_LATEST_COMMITMENT).Find(d.ctx, bson.NewDocument(), &options.FindOptions{Sort: sortFilter})
	if resErr != nil {
		return []models.LatestCommitment{},
			errors.New(fmt.Sprintf("%s %v", ERROR_LATEST_COMMITMENT_GET, resErr))
	}

	// iterate through commitments
	var latestCommitments []models.LatestCommitment
	for res.Next(d.ctx) {
		commitmentDoc := bson.NewDocument()
		if err := res.Decode(commitmentDoc); err != nil {
			fmt.Printf("%s\n", BAD_DATA_LATEST_COMMITMENT_COL)
			return []models.LatestCommitment{}, err
		}
		commitmentModel := &models.LatestCommitment{}
		modelErr := models.GetModelFromDocument(commitmentDoc, commitmentModel)
		if modelErr != nil {
			return []models.LatestCommitment{}, errors.New(fmt.Sprintf("%s %v", BAD_DATA_LATEST_COMMITMENT_COL, modelErr))
		}
		latestCommitments = append(latestCommitments, *commitmentModel)
	}
	if err := res.Err(); err != nil {
		return []models.LatestCommitment{}, errors.New(fmt.Sprintf("%s %v", BAD_DATA_LATEST_COMMITMENT_COL, err))
	}
	return latestCommitments, nil
}
