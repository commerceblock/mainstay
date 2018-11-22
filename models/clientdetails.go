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
	CLIENT_DETAILS_CLIENT_POSITION_NAME = "client_position"
	CLIENT_DETAILS_AUTH_TOKEN_NAME      = "auth_token"
	CLIENT_DETAILS_PUBKEY_NAME          = "pubkey"
)
