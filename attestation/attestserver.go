package attestation

import (
	"context"
	"log"
	"sync"

	"mainstay/clients"
	"mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// AttestServer structure
// Stores information on the latest attestation
// Responds to external API requests handled by RequestApi
type AttestServer struct {
	ctx           context.Context
	wg            *sync.WaitGroup
	updateChan    chan Attestation
	requestChan   chan models.RequestWithResponseChannel
	getLatestChan chan models.RequestWithResponseChannel
	getNextChan   chan models.RequestWithResponseChannel

	latest       Attestation
	latestHeight int32
	sideClient   clients.SidechainClient
}

// NewAttestServer returns a pointer to an AttestServer instance
func NewAttestServer(ctx context.Context, wg *sync.WaitGroup, rpcSide clients.SidechainClient, latest Attestation) *AttestServer {
	uChan := make(chan Attestation)
	rChan := make(chan models.RequestWithResponseChannel)
	latestChan := make(chan models.RequestWithResponseChannel)
	nextChan := make(chan models.RequestWithResponseChannel)
	return &AttestServer{ctx, wg, uChan, rChan, latestChan, nextChan, latest, 0, rpcSide}
}

// Update information on the latest attestation and sidechain height
func (s *AttestServer) updateLatest(tx Attestation) {
	s.latest = tx
	latestheight, err := s.sideClient.GetBlockHeight(&s.latest.attestedHash)
	if err != nil {
		log.Printf("**AttestServer** No client hash on confirmed tx")
	} else {
		s.latestHeight = latestheight
	}
}

// Return latest hash
func (s *AttestServer) getNextHash() chainhash.Hash {
	hash, err := s.sideClient.GetBestBlockHash()
	if err != nil {
		log.Fatal(err)
	}
	return *hash
}

// Verify incoming client request
func (s *AttestServer) verifyCommitment(req models.Request) interface{} {

	// read hash from request
	// read id
	// verify id
	// update latest
	// return ACK

	return nil
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
	case "Commitment":
		return s.verifyCommitment(req)
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
		case req := <-s.requestChan: // API requests
			req.Response <- s.Respond(req.Request)
		case update := <-s.updateChan: // service updates
			s.updateLatest(update)
		case latest := <-s.getLatestChan: // latest attestation
			latest.Response <- s.latest
		case next := <-s.getNextChan: // next attestation hash
			next.Response <- s.getNextHash()
		}
	}
}
