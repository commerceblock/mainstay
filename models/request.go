package models

const NAME_BEST_BLOCK = "BestBlock"
const NAME_BEST_BLOCK_HEIGHT = "BestBlockHeight"
const NAME_BLOCK = "Block"
const NAME_COMMITMENT_SEND = "CommitmentSend"
const NAME_INDEX = "Index"
const NAME_LATEST_ATTESTATION = "LatestAttestation"
const NAME_SERVER_VERIFY = "HandleServerVerify"
const NAME_TRANSACTION = "Transaction"

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
