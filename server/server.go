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
	latestCommitment  chainhash.Hash

	// to remove soon
	sideClient clients.SidechainClient
}

// NewServer returns a pointer to an Server instance
func NewServer(config *config.Config, sideClient clients.SidechainClient) *Server {
	return &Server{*models.NewAttestationDefault(), chainhash.Hash{}, sideClient}
}

// Update latest attestation in the server
func (s *Server) UpdateLatestAttestation(tx models.Attestation) error {

	// db interface

	s.latestAttestation = tx
	return nil
}

// Return latest attestation stored in the server
func (s *Server) GetLatestAttestation() (models.Attestation, error) {

	// db interface

	return s.latestAttestation, nil
}

// Return latest commitment stored in the server
func (s *Server) GetLatestCommitment() (chainhash.Hash, error) {

	// dummy just for now
	s.updateCommitment()

	//db interface

	return s.latestCommitment, nil
}

// Return commitment for a particular attestation transaction id
func (s *Server) GetAttestationCommitment(txid chainhash.Hash) (chainhash.Hash, error) {

	// db interface

	return chainhash.Hash{}, nil
}

// TODO REMOVE: Update latest commitment hash
func (s *Server) updateCommitment() {
	hash, err := s.sideClient.GetBestBlockHash()
	if err != nil {
		log.Fatal(err)
	}
	s.latestCommitment = *hash
}
