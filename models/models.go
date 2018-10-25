// Package models provides structs for request/reponse service communications.
package models

type BestBlockHeightResponse struct {
	Response
	BlockHeight int32 `json:"blockheight"`
}

type BestBlockResponse struct {
	Response
	BlockHash string `json:"blockhash"`
}

type BlockResponse struct {
	Response
	Attested bool `json:"attested"`
}

type CommintmentSend struct {
	Response
	Verified bool `json:"verified"`
}

type LatestAttestationResponse struct {
	Response
	TxHash string `json:"txhash"`
}

type TransactionResponse struct {
	Response
	Attested bool `json:"attested"`
}
