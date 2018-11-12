// Token generator
package main

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
