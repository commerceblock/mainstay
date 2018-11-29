package requestapi

// import "mainstay/models"

// // Request implementations
// // These are used by AttestationService and RequestService to send
// // requests to the main Server of the system
// // The const names below are used to differentiate between different request types
// // BaseRequest can be used for requests that pass no arguments
// // Specialised requests can be implemented by implementing RequestType()

// const (
// 	SERVER_INDEX              = "Index"
// 	SERVER_BEST_BLOCK         = "BestBlock"
// 	SERVER_BEST_BLOCK_HEIGHT  = "BestBlockHeight"
// 	SERVER_VERIFY_BLOCK       = "VerifyBlock"
// 	SERVER_LATEST_ATTESTATION = "LatestAttestation"
// 	SERVER_COMMITMENT_SEND    = "CommitmentSend"

// 	ATTESTATION_UPDATE     = "AttestationUpdate"
// 	ATTESTATION_LATEST     = "AttestationLatest"
// 	ATTESTATION_COMMITMENT = "AttestationCommitment"
// )

// // Request interface
// type Request interface {
// 	RequestType() string
// }

// // BaseRequest struct - only requestType specified
// type BaseRequest struct {
// 	requestType string
// }

// // implement interface
// func (b BaseRequest) RequestType() string {
// 	return b.requestType
// }

// // set request type
// func (b *BaseRequest) SetRequestType(requestType string) {
// 	b.requestType = requestType
// }

// // ServerVerifyBlockRequest for SERVER_VERIFY_BLOCK
// type ServerVerifyBlockRequest struct {
// 	requestType string
// 	Id          string
// }

// // implement interface
// func (v ServerVerifyBlockRequest) RequestType() string {
// 	return v.requestType
// }

// // set request type
// func (v *ServerVerifyBlockRequest) SetRequestType(requestType string) {
// 	v.requestType = requestType
// }

// // ServerCommitmentSendRequest for SERVER_COMMITMENT_SEND
// type ServerCommitmentSendRequest struct {
// 	requestType string
// 	ClientId    string
// 	Hash        string
// 	Height      string
// }

// // implement interface
// func (c ServerCommitmentSendRequest) RequestType() string {
// 	return c.requestType
// }

// // set request type
// func (c *ServerCommitmentSendRequest) SetRequestType(requestType string) {
// 	c.requestType = requestType
// }

// // AttestationUpdateRequest for ATTESTATION_UPDATE
// type AttestationUpdateRequest struct {
// 	requestType string
// 	Attestation models.Attestation
// }

// // implement interface
// func (u AttestationUpdateRequest) RequestType() string {
// 	return u.requestType
// }

// // set request type
// func (u *AttestationUpdateRequest) SetRequestType(requestType string) {
// 	u.requestType = requestType
// }

// // Channel implementation with channel for Requests and channel for Responses
// type Channel struct {
//     Requests  chan Request
//     Responses chan Response
// }

// // Return new Channel instance
// func NewChannel() *Channel {
//     channel := &Channel{}
//     channel.Requests = make(chan Request)
//     channel.Responses = make(chan Response)
//     return channel
// }

// // Channel struct used to pass response channel for responses along with request
// type RequestWithResponseChannel struct {
//     Request  Request
//     Response chan Response
// }

// // Response implementations
// // This are used by Server to reply to Requests send by either request service or attestation service
// // These types are used in Server and ServerHandlers
// // BaseResponse is used for errors
// // Specialised responses can be implemented by implementing ResponseError()

// // Response interface
// type Response interface {
//     ResponseError() string
// }

// // BaseResponse - only error specified
// type BaseResponse struct {
//     error string `json:"error"`
// }

// // implement interface
// func (r BaseResponse) ResponseError() string {
//     return r.error
// }

// // set response error
// func (r *BaseResponse) SetResponseError(err string) {
//     r.error = err
// }

// // BestBlockHeightResponse
// type BestBlockHeightResponse struct {
//     error       string `json:"error"`
//     BlockHeight int32  `json:"blockheight"`
// }

// // implement interface
// func (r BestBlockHeightResponse) ResponseError() string {
//     return r.error
// }

// // set response error
// func (r *BestBlockHeightResponse) SetResponseError(err string) {
//     r.error = err
// }

// // BestBlockResponse
// type BestBlockResponse struct {
//     error     string `json:"error"`
//     BlockHash string `json:"blockhash"`
// }

// // implement interface
// func (r BestBlockResponse) ResponseError() string {
//     return r.error
// }

// // set response error
// func (r *BestBlockResponse) SetResponseError(err string) {
//     r.error = err
// }

// // VerifyBlockResponse
// type VerifyBlockResponse struct {
//     error    string `json:"error"`
//     Attested bool   `json:"attested"`
// }

// // implement interface
// func (r VerifyBlockResponse) ResponseError() string {
//     return r.error
// }

// // set response error
// func (r *VerifyBlockResponse) SetResponseError(err string) {
//     r.error = err
// }

// // LatestAttestationResponse
// type LatestAttestationResponse struct {
//     error  string `json:"error"`
//     TxHash string `json:"txhash"`
// }

// // implement interface
// func (r LatestAttestationResponse) ResponseError() string {
//     return r.error
// }

// // set response error
// func (r *LatestAttestationResponse) SetResponseError(err string) {
//     r.error = err
// }

// // CommitmentSendResponse
// type CommitmentSendResponse struct {
//     error    string `json:"error"`
//     Verified bool   `json:"verified"`
// }

// // implement interface
// func (r CommitmentSendResponse) ResponseError() string {
//     return r.error
// }

// // set response error
// func (r *CommitmentSendResponse) SetResponseError(err string) {
//     r.error = err
// }

// // AttestationUpdateResponse
// type AttestationUpdateResponse struct {
//     error   string `json:"error"`
//     Updated bool   `json:"updated"`
// }

// // implement interface
// func (u AttestationUpdateResponse) ResponseError() string {
//     return u.error
// }

// // set response error
// func (u *AttestationUpdateResponse) SetResponseError(err string) {
//     u.error = err
// }

// // CommitmentSendResponse
// type AttestationLatestResponse struct {
//     error       string             `json:"error"`
//     Attestation models.Attestation `json:"attestation"`
// }

// // implement interface
// func (l AttestationLatestResponse) ResponseError() string {
//     return l.error
// }

// // set response error
// func (l *AttestationLatestResponse) SetResponseError(err string) {
//     l.error = err
// }

// // CommitmentSendResponse
// type AttestastionCommitmentResponse struct {
//     error      string         `json:"error"`
//     Commitment chainhash.Hash `json:"commitment"`
// }

// // implement interface
// func (r AttestastionCommitmentResponse) ResponseError() string {
//     return r.error
// }

// // set response error
// func (r *AttestastionCommitmentResponse) SetResponseError(err string) {
//     r.error = err
// }
