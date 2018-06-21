// Response handlers for attestation server

package attestation

import (
    "strconv"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
    "ocean-attestation/models"
)

func (s *AttestServer) BlockResponse(req models.Request) models.BlockResponse {
    res := models.Response{req, ""}
    var height int32
    if (len(req.Id) == 64) {
        hash, err := chainhash.NewHashFromStr(req.Id)
        if err != nil {
            res.Error = "Invalid request for block hash: " + req.Id
            return models.BlockResponse{res, false}
        }
        header, err := s.sideClient.GetBlockHeaderVerbose(hash)
        if err != nil {
            res.Error = "Invalid request block does not exist: " + req.Id
            return models.BlockResponse{res, false}
        }
        height = header.Height
    } else {
        h, err := strconv.Atoi(req.Id)
        if err != nil {
            res.Error = "Invalid request for block: " + req.Id
            return models.BlockResponse{res, false}
        }
        height = int32(h)
    }

    return models.BlockResponse{res, s.latestHeight >= height}
}

func (s *AttestServer) TransactionResponse(req models.Request) models.TransactionResponse {
    res := models.Response{req, ""}

    hash, err := chainhash.NewHashFromStr(req.Id)
    if err != nil {
        res.Error = "Invalid request for transaction: " + req.Id
        return models.TransactionResponse{res, false}
    }
    tx, err := s.sideClient.GetRawTransactionVerbose(hash)
    if err != nil {
        res.Error = "Invalid request transaction does not exist: " + req.Id
        return models.TransactionResponse{res, false}
    }
    txhash, _ := chainhash.NewHashFromStr(tx.BlockHash)
    header, err := s.sideClient.GetBlockHeaderVerbose(txhash)
    if err != nil {
        res.Error = "Invalid request transaction does not exist: " + req.Id
        return models.TransactionResponse{res, false}
    }
    return models.TransactionResponse{res, s.latestHeight >= header.Height}
}

func (s *AttestServer) BestBlockResponse(req models.Request) models.BestBlockResponse {
    return models.BestBlockResponse{models.Response{req, ""}, s.latest.attestedHash.String()}
}

func (s *AttestServer) BestBlockHeightResponse(req models.Request) models.BestBlockHeightResponse {
    return models.BestBlockHeightResponse{models.Response{req, ""}, s.latestHeight}
}

func (s *AttestServer) LatestAttestation(req models.Request) models.LatestAttestationResponse {
    return models.LatestAttestationResponse{models.Response{req, ""}, s.latest.txid.String()}
}
