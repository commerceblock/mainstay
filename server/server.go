/*
Package attestation implements the MainStay server

Implemented using an Server structure that runs a main process:
    - Server that handles responding to API requests, client requests and service requests
*/
package server

import (
	"context"
	"log"
	"sync"

	"mainstay/clients"
	"mainstay/config"
	"mainstay/models"
	"mainstay/requestapi"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Server structure
// Stores information on the latest attestation
// Responds to external API requests handled by RequestApi
type Server struct {
	ctx context.Context
	wg  *sync.WaitGroup

	attestationServiceChannel chan requestapi.RequestWithResponseChannel

	latestAttestation models.Attestation
	latestCommitment  chainhash.Hash
	commitmentKeys    []string

	// to remove soon
	latestHeight int32
	sideClient   clients.SidechainClient
}

// NewServer returns a pointer to an Server instance
func NewServer(ctx context.Context, wg *sync.WaitGroup, config *config.Config) *Server {
	attChan := make(chan requestapi.RequestWithResponseChannel)
	return &Server{ctx, wg, attChan, *models.NewAttestationDefault(), chainhash.Hash{}, config.ClientKeys(), 0, config.OceanClient()}
}

// Return channel for communcation with attestation service
func (s *Server) AttestationServiceChannel() chan requestapi.RequestWithResponseChannel {
	return s.attestationServiceChannel
}

// Update information on the latest attestation
func (s *Server) updateLatest(tx models.Attestation) bool {
	s.latestAttestation = tx

	// TODO: REMOVE - height will be embedded in the Commitment model
	latestheight, err := s.sideClient.GetBlockHeight(&s.latestAttestation.AttestedHash)
	if err != nil {
		log.Printf("**Server** No client hash on confirmed tx")
	} else {
		s.latestHeight = latestheight
	}

	return true
}

// Update latest commitment hash
func (s *Server) updateCommitment() {
	hash, err := s.sideClient.GetBestBlockHash()
	if err != nil {
		log.Fatal(err)
	}
	s.latestCommitment = *hash
}

// Attestation Respond returns appropriate response based on request type
func (s *Server) respondAttestation(req requestapi.Request) requestapi.Response {
	switch req.RequestType() {
	case requestapi.ATTESTATION_UPDATE:
		return s.ResponseAttestationUpdate(req)
	case requestapi.ATTESTATION_LATEST:
		return s.ResponseAttestationLatest(req)
	case requestapi.ATTESTATION_COMMITMENT:
		s.updateCommitment() // TODO: Remove - proper commitment tool implemented
		return s.ResponseAttestationCommitment(req)
	default:
		baseResp := requestapi.BaseResponse{}
		baseResp.SetResponseError("**Server** Non supported request type " + req.RequestType())
		return baseResp
	}
}

// Main attest server method listening to remote requests and updates
func (s *Server) Run() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case req := <-s.attestationServiceChannel: // attestation service requests
			req.Response <- s.respondAttestation(req.Request)
		}
	}
}
