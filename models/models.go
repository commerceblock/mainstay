// Requests passed between attestation and request services

package models

type Request struct {
    Name      string    `json:"name"`
    Id        string    `json:"id"`
}

type Response struct {
    Request   Request
    Error     string    `json:"error"`
}
type BlockResponse struct {
    Response
    Attested    bool    `json:"attested"`
}
type BestBlockResponse struct {
    Response
    BlockHash   string    `json:"blockhash"`
}
type TransactionResponse struct {
    Response
    Attested    bool    `json:"attested"`
}
type LatestAttestationResponse struct {
    Response
    TxHash      string    `json:"txhash"`
}

type Channel struct {
    Requests    chan Request
    Responses   chan interface{}
}

func NewChannel() *Channel {
    channel := &Channel{}
    channel.Requests = make(chan Request)
    channel.Responses = make(chan interface{})
    return channel
}
