package server

import (
	"context"
	"log"

	"mainstay/clients"
	"mainstay/config"
	"mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Server structure
// Stores information on the latest attestation and commitment
// Methods to get latest state by attestation service
type Server struct {
	latestAttestation models.Attestation
	latestCommitment  *models.Commitment

	// to remove soon
	sideClient clients.SidechainClient
}

var dbInterface DbMongo

// NewServer returns a pointer to an Server instance
func NewServer(ctx context.Context, config *config.Config, sideClient clients.SidechainClient) *Server {

	var dbErr error
	dbInterface, dbErr = NewDbMongo(ctx)
	if dbErr != nil {
		log.Fatal(dbErr)
	}

	return &Server{*models.NewAttestationDefault(), (*models.Commitment)(nil), sideClient}
}

// Update latest attestation in the server
func (s *Server) UpdateLatestAttestation(attestation models.Attestation) error {

	s.latestAttestation = attestation // to remove

	errSave := dbInterface.saveAttestation(attestation)
	if errSave != nil {
		return errSave
	}
	commitment, errCommitment := attestation.Commitment()
	if errCommitment != nil {
		return errCommitment
	}
	errSave = dbInterface.saveCommitment(*commitment)
	if errSave != nil {
		return errSave
	}
	return nil
}

// Return latest attestation stored in the server
func (s *Server) GetLatestAttestedCommitmentHash() (chainhash.Hash, error) {

	//db interface
	return dbInterface.getLatestAttestedCommitmentHash()
}

// Return latest commitment stored in the server
func (s *Server) GetLatestCommitment() (models.Commitment, error) {

	// db interface
	return dbInterface.getLatestCommitment()
}

// Return commitment for a particular attestation transaction id
func (s *Server) GetAttestationCommitment(attestationTxid chainhash.Hash) (models.Commitment, error) {

	// db interface
	_, _ = dbInterface.getAttestationCommitment(attestationTxid)

	return models.Commitment{}, nil
}

// TODO REMOVE: Update latest commitment hash
func (s *Server) updateCommitment() {
	hash, err := s.sideClient.GetBestBlockHash()
	if err != nil {
		log.Fatal(err)
	}
	s.latestCommitment, _ = models.NewCommitment([]chainhash.Hash{*hash, *hash, *hash})
}
