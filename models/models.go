// Package models provides structs for request/reponse service communications.
package models

const GET = "GET"
const POST = "POST"

const ROUTE_BEST_BLOCK = "BestBlock"
const ROUTE_BEST_BLOCK_HEIGHT = "BestBlockHeight"
const ROUTE_BLOCK = "Block"
const ROUTE_COMMITMENT_SEND = "CommitmentSend"
const ROUTE_INDEX = "Index"
const ROUTE_LATEST_ATTESTATION = "LatestAttestation"
const ROUTE_SERVER_VERIFY = "HandleServerVerify"
const ROUTE_TRANSACTION = "Transaction"

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
