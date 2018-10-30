package server

import (
	"mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Response implementations
// This are used by Server to reply to Requests send by either request service or attestation service
// These types are used in Server and ServerHandlers
// BaseResponse is used for errors
// Specialised responses can be implemented by implementing ResponseError()

// Response interface
type Response interface {
	ResponseError() string
}

// BaseResponse - only error specified
type BaseResponse struct {
	error string `json:"error"`
}

// implement interface
func (r BaseResponse) ResponseError() string {
	return r.error
}

// set response error
func (r *BaseResponse) SetResponseError(err string) {
	r.error = err
}

// AttestationUpdateResponse
type AttestationUpdateResponse struct {
	error   string `json:"error"`
	Updated bool   `json:"updated"`
}

// implement interface
func (u AttestationUpdateResponse) ResponseError() string {
	return u.error
}

// set response error
func (u *AttestationUpdateResponse) SetResponseError(err string) {
	u.error = err
}

// CommitmentSendResponse
type AttestationLatestResponse struct {
	error       string             `json:"error"`
	Attestation models.Attestation `json:"attestation"`
}

// implement interface
func (l AttestationLatestResponse) ResponseError() string {
	return l.error
}

// set response error
func (l *AttestationLatestResponse) SetResponseError(err string) {
	l.error = err
}

// CommitmentSendResponse
type AttestastionCommitmentResponse struct {
	error      string         `json:"error"`
	Commitment chainhash.Hash `json:"commitment"`
}

// implement interface
func (r AttestastionCommitmentResponse) ResponseError() string {
	return r.error
}

// set response error
func (r *AttestastionCommitmentResponse) SetResponseError(err string) {
	r.error = err
}
