package attestation

import (
	"mainstay/models"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Response handlers for requests send to AttestServer

// BlockResponse handles response to whether a Block (heigh or hash) has been attested
func (s *AttestServer) BlockResponse(req models.RequestGet_s) models.BlockResponse {
	res := models.Response{""}
	var height int32
	if len(req.Id) == 64 {
		hash, err := chainhash.NewHashFromStr(req.Id)
		if err != nil {
			res.Error = "Invalid request for block hash: " + req.Id
			return models.BlockResponse{res, false}
		}
		blockheight, err := s.sideClient.GetBlockHeight(hash)
		if err != nil {
			res.Error = "Invalid request block does not exist: " + req.Id
			return models.BlockResponse{res, false}
		}
		height = blockheight
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

// TransactionResponse handles response to whether a specific Transaction id has been attested
func (s *AttestServer) TransactionResponse(req models.RequestGet_s) models.TransactionResponse {
	res := models.Response{""}
	hash, err := chainhash.NewHashFromStr(req.Id)
	if err != nil {
		res.Error = "Invalid request for transaction: " + req.Id
		return models.TransactionResponse{res, false}
	}
	txblockhash, err := s.sideClient.GetTxBlockHash(hash)
	if err != nil {
		res.Error = "Invalid request transaction does not exist: " + req.Id
		return models.TransactionResponse{res, false}
	}
	txhash, _ := chainhash.NewHashFromStr(txblockhash)
	blockheight, err := s.sideClient.GetBlockHeight(txhash)
	if err != nil {
		res.Error = "Invalid request transaction does not exist: " + req.Id
		return models.TransactionResponse{res, false}
	}
	return models.TransactionResponse{res, s.latestHeight >= blockheight}
}

// BestBlockResponse handles reponse to Best (latest) Block attested
func (s *AttestServer) BestBlockResponse(req models.RequestGet_s) models.BestBlockResponse {
	return models.BestBlockResponse{models.Response{""}, s.latest.attestedHash.String()}
}

// BestBlockHeightResponse handles reponse to Best (latest) Block height attested
func (s *AttestServer) BestBlockHeightResponse(req models.RequestGet_s) models.BestBlockHeightResponse {
	return models.BestBlockHeightResponse{models.Response{""}, s.latestHeight}
}

// LatestAttestation handles reponse to Latest Attestation Transaction id
func (s *AttestServer) LatestAttestation(req models.RequestGet_s) models.LatestAttestationResponse {
	return models.LatestAttestationResponse{models.Response{""}, s.latest.txid.String()}
}
