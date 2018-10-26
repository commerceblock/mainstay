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

	attestationServiceChannel chan requestapi.RequestWithInterfaceChannel
	requestServiceChannel     chan requestapi.RequestWithResponseChannel

	latestAttestation models.Attestation
	latestCommitment  chainhash.Hash
	commitmentKeys    []string

	// to remove soon
	latestHeight int32
	sideClient   clients.SidechainClient
}

// NewServer returns a pointer to an Server instance
func NewServer(ctx context.Context, wg *sync.WaitGroup, config *config.Config) *Server {
	attChan := make(chan requestapi.RequestWithInterfaceChannel)
	reqChan := make(chan requestapi.RequestWithResponseChannel)
	return &Server{ctx, wg, attChan, reqChan, *models.NewAttestationDefault(), chainhash.Hash{}, config.ClientKeys(), 0, config.OceanClient()}
}

// Return channel for communcation with attestation service
func (s *Server) AttestationServiceChannel() chan requestapi.RequestWithInterfaceChannel {
	return s.attestationServiceChannel
}

// Return channel for communication with request api
func (s *Server) RequestServiceChannel() chan requestapi.RequestWithResponseChannel {
	return s.requestServiceChannel
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

// Respond returns appropriate response based on request type
func (s *Server) respond(req requestapi.Request) requestapi.Response {
	switch req.RequestType() {
	case requestapi.SERVER_VERIFY_BLOCK:
		return s.ResponseVerifyBlock(req)
	case requestapi.SERVER_BEST_BLOCK:
		return s.ResponseBestBlock(req)
	case requestapi.SERVER_BEST_BLOCK_HEIGHT:
		return s.ResponseBestBlockHeight(req)
	case requestapi.SERVER_LATEST_ATTESTATION:
		return s.ResponseLatestAttestation(req)
	case requestapi.SERVER_COMMITMENT_SEND:
		return s.ResponseCommitmentSend(req)
	default:
		baseResp := requestapi.BaseResponse{}
		baseResp.SetResponseError("**Server** Non supported request type " + req.RequestType())
		return baseResp
	}
}

// Attestation Respond returns appropriate response based on request type
func (s *Server) respondAttestation(req requestapi.Request) interface{} {
	switch req.RequestType() {
	case requestapi.ATTESTATION_UPDATE:
		updateReq := req.(requestapi.AttestationUpdateRequest)
		return s.updateLatest(updateReq.Attestation)
	case requestapi.ATTESTATION_LATEST:
		return s.latestAttestation
	case requestapi.ATTESTATION_COMMITMENT:
		s.updateCommitment() // TODO: Remove - proper commitment tool implemented
		return s.latestCommitment
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
		case req := <-s.requestServiceChannel: // request service requests
			req.Response <- s.respond(req.Request)
		case req := <-s.attestationServiceChannel: // attestation service requests
			req.Response <- s.respondAttestation(req.Request)
		}
	}
}
