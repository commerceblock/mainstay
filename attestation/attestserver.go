package attestation

import (
    "log"
    "time"

    "ocean-attestation/models"
    "ocean-attestation/clients"

    "github.com/btcsuite/btcd/chaincfg/chainhash"
)

// AttestServer structure
// Stores information on the latest attestation
// Responds to external API requests handled by RequestApi
type AttestServer struct {
    latest          Attestation
    latestHeight    int32
    sideClient      clients.SidechainClient
}

// NewAttestServer returns a pointer to an AttestServer instance
func NewAttestServer(rpcSide clients.SidechainClient, latest Attestation, tx0 string, genesis chainhash.Hash) *AttestServer{
    tx0hash, err := chainhash.NewHashFromStr(tx0)
    if err != nil {
        log.Fatal("*AttestServer* Incorrect tx hash for initial transaction")
    }
    return &AttestServer{Attestation{*tx0hash, genesis, true, time.Now()}, 0, rpcSide}
}

// Update information on the latest attestation and sidechain height
func (s *AttestServer) UpdateLatest(tx Attestation) {
    s.latest = tx
    latestheight, err := s.sideClient.GetBlockHeight(&s.latest.attestedHash)
    if err != nil {
        log.Printf("**AttestServer** No client hash on confirmed tx - Happens on init, should fix soon")
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
