// Response handlers for attestation server

package main

import (
    "strconv"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
)

func (s *AttestServer) BlockResponse(req Request) BlockResponse {
    res := Response{req, ""}
    var height int32
    if (len(req.Id) == 64) {
        hash, err := chainhash.NewHashFromStr(req.Id)
        if err != nil {
            res.Error = "Invalid request for block hash: " + req.Id
            return BlockResponse{res, false}
        }
        header, err := s.sideClient.GetBlockHeaderVerbose(hash)
        if err != nil {
            res.Error = "Invalid request block does not exist: " + req.Id
            return BlockResponse{res, false}
        }
        height = header.Height
    } else {
        h, err := strconv.Atoi(req.Id)
        if err != nil {
            res.Error = "Invalid request for block: " + req.Id
            return BlockResponse{res, false}
        }
        height = int32(h)
    }

    return BlockResponse{res, s.latestHeight >= height}
}

func (s *AttestServer) TransactionResponse(req Request) TransactionResponse {
    res := Response{req, ""}

    hash, err := chainhash.NewHashFromStr(req.Id)
    if err != nil {
        res.Error = "Invalid request for transaction: " + req.Id
        return TransactionResponse{res, false}
    }
    tx, err := s.sideClient.GetRawTransactionVerbose(hash)
    if err != nil {
        res.Error = "Invalid request transaction does not exist: " + req.Id
        return TransactionResponse{res, false}
    }
    txhash, _ := chainhash.NewHashFromStr(tx.BlockHash)
    header, err := s.sideClient.GetBlockHeaderVerbose(txhash)
    if err != nil {
        res.Error = "Invalid request transaction does not exist: " + req.Id
        return TransactionResponse{res, false}
    }
    return TransactionResponse{res, s.latestHeight >= header.Height}
}

func (s *AttestServer) BestBlockResponse(req Request) BestBlockResponse {
    return BestBlockResponse{Response{req, ""}, s.latest.clientHash.String()}
}

func (s *AttestServer) LatestAttestation(req Request) LatestAttestationResponse {
    return LatestAttestationResponse{Response{req, ""}, s.latest.txid.String()}
}
