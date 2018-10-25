package models

type Response struct {
	Error string `json:"error"`
}

type BestBlockHeightResponse struct {
	Response
	BlockHeight int32 `json:"blockheight"`
}

type BestBlockResponse struct {
	Response
	BlockHash string `json:"blockhash"`
}

type VerifyBlockResponse struct {
	Response
	Attested bool `json:"attested"`
}

type LatestAttestationResponse struct {
	Response
	TxHash string `json:"txhash"`
}

type CommitmentSendResponse struct {
	Response
	Verified bool `json:"verified"`
}
