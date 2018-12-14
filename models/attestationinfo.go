// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

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
	AttestationInfoTxidName      = "txid"
	AttestationInfoBlockhashName = "blockhash"
	AttestationInfoAmountName    = "amount"
	AttestationInfoTimeName      = "time"
)
