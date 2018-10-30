package server

import "mainstay/models"

// Request implementations
// These are used by AttestationService and RequestService to send
// requests to the main Server of the system
// The const names below are used to differentiate between different request types
// BaseRequest can be used for requests that pass no arguments
// Specialised requests can be implemented by implementing RequestType()

const (
	ATTESTATION_UPDATE     = "AttestationUpdate"
	ATTESTATION_LATEST     = "AttestationLatest"
	ATTESTATION_COMMITMENT = "AttestationCommitment"
)

// Request interface
type Request interface {
	RequestType() string
}

// BaseRequest struct - only requestType specified
type BaseRequest struct {
	requestType string
}

// implement interface
func (b BaseRequest) RequestType() string {
	return b.requestType
}

// set request type
func (b *BaseRequest) SetRequestType(requestType string) {
	b.requestType = requestType
}

// AttestationUpdateRequest for ATTESTATION_UPDATE
type AttestationUpdateRequest struct {
	requestType string
	Attestation models.Attestation
}

// implement interface
func (u AttestationUpdateRequest) RequestType() string {
	return u.requestType
}

// set request type
func (u *AttestationUpdateRequest) SetRequestType(requestType string) {
	u.requestType = requestType
}
