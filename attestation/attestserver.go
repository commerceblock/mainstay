package attestation

import (
	"context"
	"log"
	"sync"

	"mainstay/clients"
	"mainstay/models"
)

// AttestServer structure
// Stores information on the latest attestation
// Responds to external API requests handled by RequestApi
type AttestServer struct {
	ctx          context.Context
	wg           *sync.WaitGroup
	requestChan  chan models.RequestWithResponseChannel
	latestChan   chan models.RequestWithResponseChannel
	updateChan   chan Attestation
	latest       Attestation
	latestHeight int32
	sideClient   clients.SidechainClient
}

// NewAttestServer returns a pointer to an AttestServer instance
func NewAttestServer(ctx context.Context, wg *sync.WaitGroup, rpcSide clients.SidechainClient, latest Attestation) *AttestServer {
	reqChan := make(chan models.RequestWithResponseChannel)
	latestChan := make(chan models.RequestWithResponseChannel)
	updChan := make(chan Attestation)
	return &AttestServer{ctx, wg, reqChan, latestChan, updChan, latest, 0, rpcSide}
}

// Update information on the latest attestation and sidechain height
func (s *AttestServer) UpdateLatest(tx Attestation) {
	s.latest = tx
	latestheight, err := s.sideClient.GetBlockHeight(&s.latest.attestedHash)
	if err != nil {
		log.Printf("**AttestServer** No client hash on confirmed tx")
	} else {
		s.latestHeight = latestheight
	}
}

// Respond returns appropriate response based on request type
func (s *AttestServer) Respond(req models.Request) interface{} {
	switch req.Name {
	case "Block":
		return s.BlockResponse(req)
	case "BestBlock":
		return s.BestBlockResponse(req)
	case "BestBlockHeight":
		return s.BestBlockHeightResponse(req)
	case "LatestAttestation":
		return s.LatestAttestation(req)
	case "Transaction":
		return s.TransactionResponse(req)
	default:
		return models.Response{req, "**AttestServer** Non supported request type"}
	}
}

// Main attest server method listening to remote requests and updates
func (s *AttestServer) Run() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case req := <-s.requestChan:
			req.Response <- s.Respond(req.Request)
		case latest := <-s.latestChan:
			latest.Response <- s.latest
		case update := <-s.updateChan:
			s.UpdateLatest(update)
		}
	}
}
