package server

// Response handlers for requests send to Server

// LatestAttestation handles reponse to Latest Attestation Transaction id
func (s *Server) ResponseAttestationUpdate(req Request) AttestationUpdateResponse {
	updateReq := req.(AttestationUpdateRequest)

	resp := AttestationUpdateResponse{}
	resp.Updated = s.updateLatest(updateReq.Attestation)

	return resp
}

// LatestAttestation handles reponse to Latest Attestation Transaction id
func (s *Server) ResponseAttestationLatest(req Request) AttestationLatestResponse {
	resp := AttestationLatestResponse{}
	resp.Attestation = s.latestAttestation

	return resp
}

// LatestAttestation handles reponse to Latest Attestation Transaction id
func (s *Server) ResponseAttestationCommitment(req Request) AttestastionCommitmentResponse {
	resp := AttestastionCommitmentResponse{}
	resp.Commitment = s.latestCommitment

	return resp
}
