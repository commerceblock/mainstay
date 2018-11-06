package server

import (
	"context"
	"fmt"
	"os"

	"mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/mongodb/mongo-go-driver/bson"
	_ "github.com/mongodb/mongo-go-driver/bson/objectid"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/findopt"
)

const (
	// collection names
	COL_NAME_ATTESTATION       = "Attestation"
	COL_NAME_MERKLE_COMMITMENT = "MerkleCommitment"
	COL_NAME_MERKLE_PROOF      = "MerkleProof"
	COL_NAME_LATEST_COMMITMENT = "LatestCommitment"

	// LatestCommitment field names
	LATEST_COMMITMENT_CLIENT_POSITION_NAME = "client_position"
	LATEST_COMMITMENT_COMMITMENT_NAME      = "commitment"
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

// Function to get bson Document from model interface that implements MarshalBSON
func getDocumentFromModel(model interface{}) (*bson.Document, error) {

	bytes, marshalErr := bson.Marshal(model)
	if marshalErr != nil {
		return nil, marshalErr
	}
	doc, docErr := bson.ReadDocument(bytes)
	if docErr != nil {
		return nil, docErr
	}
	return doc, nil
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

// Save latest attestation to the Attestation collection
func (d *DbMongo) saveAttestation(attestation models.Attestation) error {

	// get document representation of Attestation object
	docAttestation, docErr := getDocumentFromModel(attestation)
	if docErr != nil {
		return docErr
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
	res := d.db.Collection(COL_NAME_ATTESTATION).FindOneAndUpdate(d.ctx, filterAttestation, newAttestation, findopt.Upsert(true))
	resErr := res.Decode(t)
	if resErr != nil && resErr != mongo.ErrNoDocuments {
		fmt.Printf("couldn't save attestation: %v\n", resErr)
		return resErr
	}

	return nil
}

// Handle saving Commitment underlying components to the database
func (d *DbMongo) saveCommitment(commitment models.Commitment) error {
	// store merkle commitments
	merkleCommitments := commitment.GetMerkleCommitments()
	errSave := d.saveMerkleCommitments(merkleCommitments)
	if errSave != nil {
		return errSave
	}

	// store merkle proofs
	merkleProofs := commitment.GetMerkleProofs()
	errSave = d.saveMerkleProofs(merkleProofs)
	if errSave != nil {
		return errSave
	}

	return nil
}

// Save merkle commitments to the MerkleCommitment collection
func (d *DbMongo) saveMerkleCommitments(commitments []models.CommitmentMerkleCommitment) error {
	for pos := range commitments {
		// get document representation of each commitment
		// get document representation of Attestation object
		docCommitment, docErr := getDocumentFromModel(commitments[pos])
		if docErr != nil {
			return docErr
		}

		newCommitment := bson.NewDocument(
			bson.EC.SubDocument("$set", docCommitment),
		)

		// search if merkle commitment already exists
		filterMerkleCommitment := bson.NewDocument(
			bson.EC.String(models.COMMITMENT_MERKLE_ROOT_NAME, docCommitment.Lookup(models.COMMITMENT_MERKLE_ROOT_NAME).StringValue()),
			bson.EC.Int32(models.COMMITMENT_CLIENT_POSITION_NAME, docCommitment.Lookup(models.COMMITMENT_CLIENT_POSITION_NAME).Int32()),
		)

		// insert or update merkle commitment
		t := bson.NewDocument()
		res := d.db.Collection(COL_NAME_MERKLE_COMMITMENT).FindOneAndUpdate(d.ctx, filterMerkleCommitment, newCommitment, findopt.Upsert(true))
		resErr := res.Decode(t)
		if resErr != nil && resErr != mongo.ErrNoDocuments {
			fmt.Printf("couldn't save merkle commitment: %v\n", resErr)
			return resErr
		}
	}
	return nil
}

// Save merkle proofs to the MerkleProof collection
func (d *DbMongo) saveMerkleProofs(proofs []models.CommitmentMerkleProof) error {
	for pos := range proofs {
		// get document representation of merkle proof
		docProof, docErr := getDocumentFromModel(proofs[pos])
		if docErr != nil {
			return docErr
		}

		newProof := bson.NewDocument(
			bson.EC.SubDocument("$set", docProof),
		)

		// search if merkle proof already exists
		filterMerkleProof := bson.NewDocument(
			bson.EC.String(models.PROOF_MERKLE_ROOT_NAME, docProof.Lookup(models.PROOF_MERKLE_ROOT_NAME).StringValue()),
			bson.EC.Int32(models.PROOF_CLIENT_POSITION_NAME, docProof.Lookup(models.PROOF_CLIENT_POSITION_NAME).Int32()),
		)

		// insert or update merkle proof
		t := bson.NewDocument()
		res := d.db.Collection(COL_NAME_MERKLE_PROOF).FindOneAndUpdate(d.ctx, filterMerkleProof, newProof, findopt.Upsert(true))
		resErr := res.Decode(t)
		if resErr != nil && resErr != mongo.ErrNoDocuments {
			fmt.Printf("couldn't be created: %v\n", resErr)
			return resErr
		}
	}
	return nil
}

// Return latest attested commitment hash from Attestation collection
func (d *DbMongo) getLatestAttestedCommitmentHash() (chainhash.Hash, error) {

	// filter by inserted_at DESC to get latest attestation
	sortFilter := bson.NewDocument(bson.EC.Int32(models.ATTESTATION_INSERTED_AT_NAME, -1))

	// ADD FILTER CONFIRMED ONLY

	attestationDoc := bson.NewDocument()
	resErr := d.db.Collection(COL_NAME_ATTESTATION).FindOne(d.ctx, bson.NewDocument(), findopt.Sort(sortFilter)).Decode(attestationDoc)
	if resErr != nil {
		fmt.Printf("couldn't get latest attestation: %v\n", resErr)
		return chainhash.Hash{}, resErr
	}

	// get merkle root and return the hash
	merkle_root := attestationDoc.Lookup(models.ATTESTATION_MERKLE_ROOT_NAME).StringValue()
	commitmentHash, errHash := chainhash.NewHashFromStr(merkle_root)
	if errHash != nil {
		fmt.Printf("bad data in merkle_root column: %s\n", merkle_root)
		return chainhash.Hash{}, errHash
	}
	return *commitmentHash, nil
}

// Return latest Commitment constructed from commitments of MerkleCommitment collection
func (d *DbMongo) getLatestCommitment() (models.Commitment, error) {

	// sort by client position to get correct commitment order
	sortFilter := bson.NewDocument(bson.EC.Int32(LATEST_COMMITMENT_CLIENT_POSITION_NAME, 1))
	res, resErr := d.db.Collection(COL_NAME_LATEST_COMMITMENT).Find(d.ctx, bson.NewDocument(), findopt.Sort(sortFilter))
	if resErr != nil {
		fmt.Printf("couldn't get latest commitment: %v\n", resErr)
		return models.Commitment{}, resErr
	}

	// iterate through commitments
	var commitmentHashes []chainhash.Hash
	for res.Next(d.ctx) {
		commitmentDoc := bson.NewDocument()
		if err := res.Decode(commitmentDoc); err != nil {
			fmt.Printf("bad data in %s table: %s\n", COL_NAME_LATEST_COMMITMENT, res)
			return models.Commitment{}, err
		}
		commitment := commitmentDoc.Lookup(LATEST_COMMITMENT_COMMITMENT_NAME).StringValue()
		commitmentHash, errHash := chainhash.NewHashFromStr(commitment)
		if errHash != nil {
			fmt.Printf("bad data in commitment column: %s\n", commitment)
			return models.Commitment{}, errHash
		}
		commitmentHashes = append(commitmentHashes, *commitmentHash)
	}
	if err := res.Err(); err != nil {
		return models.Commitment{}, fmt.Errorf("could not decode data: %v", err)
	}

	// contruct Commitment from MerkleCommitment commitment hashes
	commitment, errCommitment := models.NewCommitment(commitmentHashes)
	if errCommitment != nil {
		return models.Commitment{}, errCommitment
	}
	return *commitment, nil
}

// Return Commitment from MerkleCommitment commitments for attestation with given txid hash
func (d *DbMongo) getAttestationCommitment(attestationTxid chainhash.Hash) (models.Commitment, error) {

	// get merke_root from Attestation collection for attestation txid provided
	filterAttestation := bson.NewDocument(bson.EC.String(models.ATTESTATION_TXID_NAME, attestationTxid.String()))
	attestationDoc := bson.NewDocument()
	resErr := d.db.Collection(COL_NAME_ATTESTATION).FindOne(d.ctx, filterAttestation).Decode(attestationDoc)
	if resErr != nil {
		fmt.Printf("couldn't get latest: %v\n", resErr)
		return models.Commitment{}, resErr
	}

	// filter MerkleCommitment collection by merkle_root and sort for client position
	sortFilter := bson.NewDocument(bson.EC.Int32(models.COMMITMENT_CLIENT_POSITION_NAME, 1))
	merkle_root := attestationDoc.Lookup(models.COMMITMENT_MERKLE_ROOT_NAME).StringValue()
	filterMerkleRoot := bson.NewDocument(bson.EC.String(models.COMMITMENT_MERKLE_ROOT_NAME, merkle_root))
	res, resErr := d.db.Collection(COL_NAME_MERKLE_COMMITMENT).Find(d.ctx, filterMerkleRoot, findopt.Sort(sortFilter))
	if resErr != nil {
		fmt.Printf("couldn't get latest: %v\n", resErr)
		return models.Commitment{}, resErr
	}

	// fetch commitments
	var commitmentHashes []chainhash.Hash
	for res.Next(d.ctx) {
		commitmentDoc := bson.NewDocument()
		if err := res.Decode(commitmentDoc); err != nil {
			fmt.Printf("bad data in %s table: %s\n", COL_NAME_MERKLE_COMMITMENT, res)
			return models.Commitment{}, err
		}
		commitment := commitmentDoc.Lookup(models.COMMITMENT_COMMITMENT_NAME).StringValue()
		commitmentHash, errHash := chainhash.NewHashFromStr(commitment)
		if errHash != nil {
			fmt.Printf("bad data in commitment column: %s\n", commitment)
			return models.Commitment{}, errHash
		}
		commitmentHashes = append(commitmentHashes, *commitmentHash)
	}
	if err := res.Err(); err != nil {
		return models.Commitment{}, fmt.Errorf("could not decode data: %v", err)
	}

	// construct Commitment from MerkleCommitment commitments
	commitment, errCommitment := models.NewCommitment(commitmentHashes)
	if errCommitment != nil {
		return models.Commitment{}, errCommitment
	}
	return *commitment, nil
}
