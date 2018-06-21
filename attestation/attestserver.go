// Attestation Server storing attestation subchain and responding to requests

package attestation

import (
    "log"
    "time"
    "github.com/btcsuite/btcd/rpcclient"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
    "ocean-attestation/models"
)

type AttestServer struct {
    latest          Attestation
    latestHeight    int32
    sideClient      *rpcclient.Client
}

func NewAttestServer(rpcSide *rpcclient.Client, latest Attestation, tx0 string, genesis chainhash.Hash) *AttestServer{
    tx0hash, err := chainhash.NewHashFromStr(tx0)
    if err != nil {
        log.Fatal("*AttestServer* Incorrect tx hash for initial transaction")
    }
    return &AttestServer{Attestation{*tx0hash, genesis, true, time.Now()}, 0, rpcSide}
}

// Keep the latest attested transaction for handling requests
func (s *AttestServer) UpdateLatest(tx Attestation) {
    s.latest = tx
    latestheader, err := s.sideClient.GetBlockHeaderVerbose(&s.latest.attestedHash)
    if err != nil {
        log.Printf("**AttestServer** No client hash on confirmed tx - Happens on init, should fix soon")
    } else {
        s.latestHeight = latestheader.Height
    }
}

// Respond to requests by the request service
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
