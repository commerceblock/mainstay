package server

import (
	"mainstay/models"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Response handlers for requests send to Server

// VerifyBlockResponse handles response to whether a Block (heigh or hash) has been attested
func (s *Server) VerifyBlockResponse(req models.RequestGet_s) models.VerifyBlockResponse {
	res := models.Response{""}
	var height int32
	if len(req.Id) == 64 {
		hash, err := chainhash.NewHashFromStr(req.Id)
		if err != nil {
			res.Error = "Invalid request for block hash: " + req.Id
			return models.VerifyBlockResponse{res, false}
		}
		blockheight, err := s.sideClient.GetBlockHeight(hash)
		if err != nil {
			res.Error = "Invalid request block does not exist: " + req.Id
			return models.VerifyBlockResponse{res, false}
		}
		height = blockheight
	} else {
		h, err := strconv.Atoi(req.Id)
		if err != nil {
			res.Error = "Invalid request for block: " + req.Id
			return models.VerifyBlockResponse{res, false}
		}
		height = int32(h)
	}

	return models.VerifyBlockResponse{res, s.latestHeight >= height}
}

// BestVerifyBlockResponse handles reponse to Best (latest) Block attested
func (s *Server) BestBlockResponse(req models.RequestGet_s) models.BestBlockResponse {
	return models.BestBlockResponse{models.Response{""}, s.latestAttestation.AttestedHash.String()}
}

// BestBlockHeightResponse handles reponse to Best (latest) Block height attested
func (s *Server) BestBlockHeightResponse(req models.RequestGet_s) models.BestBlockHeightResponse {
	return models.BestBlockHeightResponse{models.Response{""}, s.latestHeight}
}

// LatestAttestation handles reponse to Latest Attestation Transaction id
func (s *Server) LatestAttestation(req models.RequestGet_s) models.LatestAttestationResponse {
	return models.LatestAttestationResponse{models.Response{""}, s.latestAttestation.Txid.String()}
}
