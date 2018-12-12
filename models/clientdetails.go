// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package models

import (
	_ "github.com/mongodb/mongo-go-driver/bson"
)

// struct for db ClientDetails
type ClientDetails struct {
	ClientPosition int32  `bson:"client_position"`
	AuthToken      string `bson:"auth_token"`
	Pubkey         string `bson:"pubkey"`
}

// ClientDetails field names
const (
	ClientDetailsClientPositionName = "client_position"
	ClientDetailsAuthTokenName      = "auth_token"
	ClientDetailsPubkeyName         = "pubkey"
)
