package requestapi

const (
	SERVER_INDEX              = "Index"
	SERVER_BEST_BLOCK         = "BestBlock"
	SERVER_BEST_BLOCK_HEIGHT  = "BestBlockHeight"
	SERVER_VERIFY_BLOCK       = "VerifyBlock"
	SERVER_LATEST_ATTESTATION = "LatestAttestation"
	SERVER_COMMITMENT_SEND    = "CommitmentSend"
)

type Request interface {
    RequestType() (string)
}

// base request type
type BaseRequest struct {
   requestType   string `json:"type"`
}

func (b BaseRequest) RequestType() string {
    return b.requestType
}

func (b *BaseRequest) SetRequestType(requestType string) {
    b.requestType = requestType
}

// request type for SERVER_VERIFY_BLOCK
type ServerVerifyBlockRequest struct {
   requestType   string `json:"requestType"`
   Id         string `json:"id"`
}

func (v ServerVerifyBlockRequest) RequestType() string {
    return v.requestType
}

func (v *ServerVerifyBlockRequest) SetRequestType(requestType string) {
    v.requestType = requestType
}

// request type for SERVER_COMMITMENT_SEND
type ServerCommitmentSendRequest struct {
    requestType string `json:"requestType"`
    ClientId string `json:"clientId"`
    Hash     string `json:"hash"`
    Height   string `json:"height"`
}

func (c ServerCommitmentSendRequest) RequestType() string {
    return c.requestType
}

func (c *ServerCommitmentSendRequest) SetRequestType(requestType string) {
    c.requestType = requestType
}
