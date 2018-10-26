package server

import (
	"mainstay/requestapi"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Response handlers for requests send to Server

// VerifyBlockResponse handles response to whether a Block (height or hash) has been attested
func (s *Server) ResponseVerifyBlock(req requestapi.Request) requestapi.VerifyBlockResponse {
	verifyReq := req.(requestapi.ServerVerifyBlockRequest)

	resp := requestapi.VerifyBlockResponse{}
	var height int32
	if len(verifyReq.Id) == 64 {
		hash, err := chainhash.NewHashFromStr(verifyReq.Id)
		if err != nil {
			resp.SetResponseError("Invalid request for block hash: " + verifyReq.Id)
			resp.Attested = false
			return resp
		}
		blockheight, err := s.sideClient.GetBlockHeight(hash)
		if err != nil {
			resp.SetResponseError("Invalid request block does not exist: " + verifyReq.Id)
			resp.Attested = false
			return resp
		}
		height = blockheight
	} else {
		h, err := strconv.Atoi(verifyReq.Id)
		if err != nil {
			resp.SetResponseError("Invalid request for block: " + verifyReq.Id)
			resp.Attested = false
			return resp
		}
		height = int32(h)
	}
	resp.Attested = s.latestHeight >= height
	return resp
}

// BestVerifyBlockResponse handles reponse to Best (latest) Block attested
func (s *Server) ResponseBestBlock(req requestapi.Request) requestapi.BestBlockResponse {
	resp := requestapi.BestBlockResponse{}
	resp.BlockHash = s.latestAttestation.AttestedHash.String()
	return resp
}

// BestBlockHeightResponse handles reponse to Best (latest) Block height attested
func (s *Server) ResponseBestBlockHeight(req requestapi.Request) requestapi.BestBlockHeightResponse {
	resp := requestapi.BestBlockHeightResponse{}
	resp.BlockHeight = s.latestHeight
	return resp
}

// LatestAttestation handles reponse to Latest Attestation Transaction id
func (s *Server) ResponseLatestAttestation(req requestapi.Request) requestapi.LatestAttestationResponse {
	resp := requestapi.LatestAttestationResponse{}
	resp.TxHash = s.latestAttestation.Txid.String()
	return resp
}

// ResponseCommitmentSend handles response to client POST requests sending new commitments
func (s *Server) ResponseCommitmentSend(req requestapi.Request) requestapi.CommitmentSendResponse {
	commitmentReq := req.(requestapi.ServerCommitmentSendRequest)

	resp := requestapi.CommitmentSendResponse{}

	for _, key := range s.commitmentKeys {
		if commitmentReq.ClientId == key {

			hash, errHash := chainhash.NewHashFromStr(commitmentReq.Hash)
			if errHash != nil {
				resp.SetResponseError(errHash.Error() + "\nInvalid commitment hash: " + commitmentReq.Hash)
				resp.Verified = false
				return resp
			}

			height, errHeight := strconv.Atoi(commitmentReq.Height)
			if errHeight != nil {
				resp.SetResponseError(errHeight.Error() + "\nInvalid commitment height: " + commitmentReq.Height)
				resp.Verified = false
				return resp
			}

			// update latest commitment
			s.latestCommitment = *hash
			s.latestHeight = int32(height)

			resp.Verified = true
			return resp
		}
	}

	resp.SetResponseError("Invalid client identity: " + commitmentReq.ClientId)
	resp.Verified = false
	return resp
}

// LatestAttestation handles reponse to Latest Attestation Transaction id
func (s *Server) ResponseAttestationUpdate(req requestapi.Request) requestapi.AttestationUpdateResponse {
	updateReq := req.(requestapi.AttestationUpdateRequest)

	resp := requestapi.AttestationUpdateResponse{}
	resp.Updated = s.updateLatest(updateReq.Attestation)

	return resp
}

// LatestAttestation handles reponse to Latest Attestation Transaction id
func (s *Server) ResponseAttestationLatest(req requestapi.Request) requestapi.AttestationLatestResponse {
	resp := requestapi.AttestationLatestResponse{}
	resp.Attestation = s.latestAttestation

	return resp
}

// LatestAttestation handles reponse to Latest Attestation Transaction id
func (s *Server) ResponseAttestationCommitment(req requestapi.Request) requestapi.AttestastionCommitmentResponse {
	resp := requestapi.AttestastionCommitmentResponse{}
	resp.Commitment = s.latestCommitment

	return resp
}
