/*
Package attestation implements the MainStay server

Implemented using an Server structure that runs a main process Server that handles
responding to requests from Attestation service and storing latest Attestations / Commitments
*/
package server

import (
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

// NewServer returns a pointer to an Server instance
func NewServer(config *config.Config, sideClient clients.SidechainClient) *Server {
	return &Server{*models.NewAttestationDefault(), (*models.Commitment)(nil), sideClient}
}

// Update latest attestation in the server
func (s *Server) UpdateLatestAttestation(attestation models.Attestation, confirmed bool) error {

	// db interface
	// if confirmed - else
	// err := db.Store(attestation)
	s.latestAttestation = attestation

	return nil
}

// Return latest attestation stored in the server
func (s *Server) GetLatestAttestation() (models.Attestation, error) {

	// db interface
	// attestation, err := db.GetLatestAttestation()

	return s.latestAttestation, nil
}

// Return latest commitment stored in the server
func (s *Server) GetLatestCommitment() (models.Commitment, error) {

	// dummy just for now
	s.updateCommitment()

	// db interface
	// commitment, err := db.GetLatestCommitment()

	return *s.latestCommitment, nil
}

// Return commitment for a particular attestation transaction id
func (s *Server) GetAttestationCommitment(txid chainhash.Hash) (models.Commitment, error) {

	// db interface
	// commitment, err := db.GetCommitmentForAttestation(txid)

	commitment, _ := models.NewCommitment([]chainhash.Hash{chainhash.Hash{}})
	return *commitment, nil
}

// TODO REMOVE: Update latest commitment hash
func (s *Server) updateCommitment() {
	hash, err := s.sideClient.GetBestBlockHash()
	if err != nil {
		log.Fatal(err)
	}
	s.latestCommitment, _ = models.NewCommitment([]chainhash.Hash{*hash})
}
