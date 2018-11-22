// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

// Token generator

import (
	"fmt"
	"log"

	"github.com/satori/go.uuid"
)

// main
func main() {
	// New auth token ID for client
	fmt.Println()
	fmt.Println("***** Client Auth Token identification ******")
	uuid, err := uuid.NewV4()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("new-uuid: %s\n", uuid.String())
	fmt.Println()
}
