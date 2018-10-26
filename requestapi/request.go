package requestapi

import "mainstay/models"

// Request implementations
// These are used by AttestationService and RequestService to send
// requests to the main Server of the system
// The const names below are used to differentiate between different request types
// BaseRequest can be used for requests that pass no arguments
// Specialised requests can be implemented by implementing RequestType()

const (
	SERVER_INDEX              = "Index"
	SERVER_BEST_BLOCK         = "BestBlock"
	SERVER_BEST_BLOCK_HEIGHT  = "BestBlockHeight"
	SERVER_VERIFY_BLOCK       = "VerifyBlock"
	SERVER_LATEST_ATTESTATION = "LatestAttestation"
	SERVER_COMMITMENT_SEND    = "CommitmentSend"

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


// ServerVerifyBlockRequest for SERVER_VERIFY_BLOCK
type ServerVerifyBlockRequest struct {
	requestType string
	Id          string
}

// implement interface
func (v ServerVerifyBlockRequest) RequestType() string {
	return v.requestType
}

// set request type
func (v *ServerVerifyBlockRequest) SetRequestType(requestType string) {
	v.requestType = requestType
}


// ServerCommitmentSendRequest for SERVER_COMMITMENT_SEND
type ServerCommitmentSendRequest struct {
	requestType string
	ClientId    string
	Hash        string
	Height      string
}

// implement interface
func (c ServerCommitmentSendRequest) RequestType() string {
	return c.requestType
}

// set request type
func (c *ServerCommitmentSendRequest) SetRequestType(requestType string) {
	c.requestType = requestType
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
