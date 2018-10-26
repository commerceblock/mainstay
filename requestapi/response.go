package requestapi

type Response interface {
	ResponseError() (string)
}

type BaseResponse struct {
    error string `json:"error"`
}

func (r BaseResponse) ResponseError() string {
    return r.error
}

func (r *BaseResponse) SetResponseError(err string) {
    r.error = err
}

type BestBlockHeightResponse struct {
    error string `json:"error"`
	BlockHeight int32 `json:"blockheight"`
}

func (r BestBlockHeightResponse) ResponseError() string {
    return r.error
}

func (r *BestBlockHeightResponse) SetResponseError(err string) {
    r.error = err
}

type BestBlockResponse struct {
    error string `json:"error"`
	BlockHash string `json:"blockhash"`
}

func (r BestBlockResponse) ResponseError() string {
    return r.error
}

func (r *BestBlockResponse) SetResponseError(err string) {
    r.error = err
}

type VerifyBlockResponse struct {
    error string `json:"error"`
	Attested bool `json:"attested"`
}

func (r VerifyBlockResponse) ResponseError() string {
    return r.error
}

func (r *VerifyBlockResponse) SetResponseError(err string) {
    r.error = err
}

type LatestAttestationResponse struct {
    error string `json:"error"`
	TxHash string `json:"txhash"`
}

func (r LatestAttestationResponse) ResponseError() string {
    return r.error
}

func (r *LatestAttestationResponse) SetResponseError(err string) {
    r.error = err
}

type CommitmentSendResponse struct {
    error string `json:"error"`
	Verified bool `json:"verified"`
}

func (r CommitmentSendResponse) ResponseError() string {
    return r.error
}

func (r *CommitmentSendResponse) SetResponseError(err string) {
    r.error = err
}
