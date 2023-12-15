// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

// Token generator

import (
	"mainstay/log"

	uuid "github.com/satori/go.uuid"
)

// main
func main() {
	// New auth token ID for client
	log.Infoln()
	log.Infoln("***** Client Auth Token identification ******")
	uuid := uuid.NewV4()
	log.Infof("new-uuid: %s\n", uuid.String())
	log.Infoln()
}
