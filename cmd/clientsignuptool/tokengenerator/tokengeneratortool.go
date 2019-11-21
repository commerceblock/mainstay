// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

// Token generator

import (
	"github.com/satori/go.uuid"
	"mainstay/log"
)

// main
func main() {
	// New auth token ID for client
	log.Infoln()
	log.Infoln("***** Client Auth Token identification ******")
	uuid, err := uuid.NewV4()
	if err != nil {
		log.Error(err)
	}
	log.Infof("new-uuid: %s\n", uuid.String())
	log.Infoln()
}
