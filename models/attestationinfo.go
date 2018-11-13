package models

// struct for db AttestationInfo
type AttestationInfo struct {
	Txid      string  `bson:"txid"`
	Blockhash string  `bson:"blockhash"`
	Fee       float64 `bson:"fee"`
	Time      int64   `bson:"time"`
}

// AttestationInfo field names
const (
	ATTESTATION_INFO_TXID_NAME      = "txid"
	ATTESTATION_INFO_BLOCKHASH_NAME = "blockhash"
	ATTESTATION_INFO_FEE_NAME       = "fee"
	ATTESTATION_INFO_TIME_NAME      = "time"
)
