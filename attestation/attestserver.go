package attestation

import (
	"log"
	"mainstay/clients"
	"mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// AttestServer structure
// Stores information on the latest attestation
// Responds to external API requests handled by RequestApi
type AttestServer struct {
	latest       Attestation
	latestHeight int32
	sideClient   clients.SidechainClient
}

// NewAttestServer returns a pointer to an AttestServer instance
func NewAttestServer(rpcSide clients.SidechainClient, latest Attestation, tx0 string) *AttestServer {
	tx0hash, err := chainhash.NewHashFromStr(tx0)
	if err != nil {
		log.Fatal("*AttestServer* Incorrect tx hash for initial transaction")
	}
	return &AttestServer{*NewAttestation(*tx0hash, chainhash.Hash{}, ASTATE_CONFIRMED), 0, rpcSide}
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
func (s *AttestServer) Respond(req models.RequestGet_s) interface{} {
	switch req.Name {
	case models.ROUTE_BLOCK:
		return s.BlockResponse(req)
	case models.ROUTE_BEST_BLOCK:
		return s.BestBlockResponse(req)
	case models.ROUTE_BEST_BLOCK_HEIGHT:
		return s.BestBlockHeightResponse(req)
	case models.ROUTE_LATEST_ATTESTATION:
		return s.LatestAttestation(req)
	case models.ROUTE_TRANSACTION:
		return s.TransactionResponse(req)
	default:
		return models.Response{"**AttestServer** Non supported request type"}
	}
}
