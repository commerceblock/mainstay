package server

import (
	"context"
	"log"

	"mainstay/config"
	"mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Server structure
// Stores information on the latest attestation and commitment
// Methods to get latest state by attestation service
type Server struct {
	dbInterface Db
}

// NewServer returns a pointer to an Server instance
func NewServer(ctx context.Context, config *config.Config) *Server {

	dbInterface, dbErr := NewDbMongo(ctx)
	if dbErr != nil {
		log.Fatal(dbErr)
	}

	return &Server{dbInterface}
}

// Update latest attestation in the server
func (s *Server) UpdateLatestAttestation(attestation models.Attestation) error {

	errSave := s.dbInterface.saveAttestation(attestation)
	if errSave != nil {
		return errSave
	}
	commitment, errCommitment := attestation.Commitment()
	if errCommitment != nil {
		return errCommitment
	}
	errSave = s.dbInterface.saveCommitment(*commitment)
	if errSave != nil {
		return errSave
	}
	return nil
}

// Return latest attestation stored in the server
func (s *Server) GetLatestAttestedCommitmentHash() (chainhash.Hash, error) {

	//db interface
	return s.dbInterface.getLatestAttestedCommitmentHash()
}

// Return latest commitment stored in the server
func (s *Server) GetLatestCommitment() (models.Commitment, error) {

	// db interface
	return s.dbInterface.getLatestCommitment()
}

// Return commitment for a particular attestation transaction id
func (s *Server) GetAttestationCommitment(attestationTxid chainhash.Hash) (models.Commitment, error) {

	// db interface
	_, _ = s.dbInterface.getAttestationCommitment(attestationTxid)

	return models.Commitment{}, nil
}
