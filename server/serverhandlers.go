package server

import (
	"mainstay/requestapi"
)

// Response handlers for requests send to Server

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
