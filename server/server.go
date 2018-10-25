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

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Server structure
// Stores information on the latest attestation
// Responds to external API requests handled by RequestApi
type Server struct {
	ctx           context.Context
	wg            *sync.WaitGroup
	updateChan    chan models.Attestation
	requestChan   chan models.RequestWithResponseChannel
	getLatestChan chan models.RequestWithResponseChannel
	getNextChan   chan models.RequestWithResponseChannel

	latestAttestation models.Attestation
	latestCommitment  chainhash.Hash
	commitmentKeys    []string

	// to remove soon
	latestHeight int32
	sideClient   clients.SidechainClient
}

// NewServer returns a pointer to an Server instance
func NewServer(ctx context.Context, wg *sync.WaitGroup, config *config.Config) *Server {
	uChan := make(chan models.Attestation)
	rChan := make(chan models.RequestWithResponseChannel)
	latestChan := make(chan models.RequestWithResponseChannel)
	nextChan := make(chan models.RequestWithResponseChannel)
	return &Server{ctx, wg, uChan, rChan, latestChan, nextChan, *models.NewAttestationDefault(), chainhash.Hash{}, config.ClientKeys(), 0, config.OceanClient()}
}

// Return request channel to allow request service to push client requests
func (s *Server) RequestChan() chan models.RequestWithResponseChannel {
	return s.requestChan
}

// Return update channel to allow service to push latest attestation updates
func (s *Server) UpdateChan() chan models.Attestation {
	return s.updateChan
}

// Return request channel to allow service to make requests for latest attestation
func (s *Server) GetLatestChan() chan models.RequestWithResponseChannel {
	return s.getLatestChan
}

// Return latest channel to allow service to make request for latest commitment hash
func (s *Server) GetNextChan() chan models.RequestWithResponseChannel {
	return s.getNextChan
}

// Update information on the latest attestation
func (s *Server) updateLatest(tx models.Attestation) {
	s.latestAttestation = tx
	latestheight, err := s.sideClient.GetBlockHeight(&s.latestAttestation.AttestedHash)
	if err != nil {
		log.Printf("**Server** No client hash on confirmed tx")
	} else {
		s.latestHeight = latestheight
	}
}

// Update latest commitment hash
func (s *Server) updateCommitment() {
	hash, err := s.sideClient.GetBestBlockHash()
	if err != nil {
		log.Fatal(err)
	}
	s.latestCommitment = *hash
}

// Verify incoming client request
func (s *Server) verifyCommitment(req models.Request) models.CommitmentSendResponse {

	res := models.Response{""}

	for _, key := range s.commitmentKeys {
		if req.Id == key {

			// TODO
			// updateCommitment ()

			return models.CommitmentSendResponse{res, true}
		}
	}

	res.Error = "Invalid client identity: " + req.Id
	return models.CommitmentSendResponse{res, false}
}

// Respond returns appropriate response based on request type
func (s *Server) respond(req models.RequestGet_s) interface{} {
	switch req.Name {
	case models.REQUEST_VERIFY_BLOCK:
		return s.VerifyBlockResponse(req)
	case models.REQUEST_BEST_BLOCK:
		return s.BestBlockResponse(req)
	case models.REQUEST_BEST_BLOCK_HEIGHT:
		return s.BestBlockHeightResponse(req)
	case models.REQUEST_LATEST_ATTESTATION:
		return s.LatestAttestation(req)
	default:
		return models.Response{"**Server** Non supported request type"}
	}
}

// Main attest server method listening to remote requests and updates
func (s *Server) Run() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case req := <-s.requestChan: // api requests
			req.Response <- s.respond(req.Request)

			// case req := <s.requestChan_POST: // commitment requests
			//     req.Response <- s.verifyCommitment(req.Request)

		case update := <-s.updateChan: // attestation service updates
			s.updateLatest(update)
		case latest := <-s.getLatestChan: // latest attestation request
			latest.Response <- s.latestAttestation
		case next := <-s.getNextChan: // next attestation hash request
			s.updateCommitment()
			next.Response <- s.latestCommitment
		}
	}
}
