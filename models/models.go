// Package models provides structs for request/reponse service communications.
package models

type Request struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type Response struct {
	Request Request
	Error   string `json:"error"`
}
type BlockResponse struct {
	Response
	Attested bool `json:"attested"`
}
type BestBlockResponse struct {
	Response
	BlockHash string `json:"blockhash"`
}
type BestBlockHeightResponse struct {
	Response
	BlockHeight int32 `json:"blockheight"`
}
type TransactionResponse struct {
	Response
	Attested bool `json:"attested"`
}
type LatestAttestationResponse struct {
	Response
	TxHash string `json:"txhash"`
}


type RequestWithResponseChannel struct {
    Request  Request
    Response chan interface{}
}

type Channel struct {
	Requests  chan Request
	Responses chan interface{}
}

func NewChannel() *Channel {
	channel := &Channel{}
	channel.Requests = make(chan Request)
	channel.Responses = make(chan interface{})
	return channel
}
