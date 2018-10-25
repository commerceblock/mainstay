package models

const (
	REQUEST_INDEX              = "Index"
	REQUEST_BEST_BLOCK         = "BestBlock"
	REQUEST_BEST_BLOCK_HEIGHT  = "BestBlockHeight"
	REQUEST_VERIFY_BLOCK       = "VerifyBlock"
	REQUEST_LATEST_ATTESTATION = "LatestAttestation"
	REQUEST_COMMITMENT_SEND    = "CommitmentSend"
)

type Request struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type RequestGet_s struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type RequestPost_s struct {
	Name     string `json:"name"`
	ClientId string `json:"clientId"`
	Hash     string `json:"hash"`
	Height   string `json:"height"`
}
