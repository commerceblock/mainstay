package models

// struct for db AttestationInfo
type AttestationInfo struct {
	Txid      string `bson:"txid"`
	Blockhash string `bson:"blockhash"`
	Amount    int64  `bson:"amount"`
	Time      int64  `bson:"time"`
}

// AttestationInfo field names
const (
	ATTESTATION_INFO_TXID_NAME      = "txid"
	ATTESTATION_INFO_BLOCKHASH_NAME = "blockhash"
	ATTESTATION_INFO_AMOUNT_NAME    = "amount"
	ATTESTATION_INFO_TIME_NAME      = "time"
)
