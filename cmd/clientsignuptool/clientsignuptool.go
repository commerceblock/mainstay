// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

// Client signup tool

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"mainstay/config"
	"mainstay/models"
	"mainstay/server"

	"github.com/btcsuite/btcd/btcec"
	"github.com/satori/go.uuid"
)

const ConfPath = "/src/mainstay/cmd/clientsignuptool/conf.json"

var (
	mainConfig *config.Config
	dbMongo    *server.DbMongo
)

// init
func init() {
	confFile, confErr := config.GetConfFile(os.Getenv("GOPATH") + ConfPath)
	if confErr != nil {
		log.Fatal(confErr)
	}
	var mainConfigErr error
	mainConfig, mainConfigErr = config.NewConfig(confFile)
	if mainConfigErr != nil {
		log.Fatal(mainConfigErr)
	}
}

// print client details
func printClientDetails() {
	// Read existing clients and get next available client position
	fmt.Println("existing clients")
	details, errDb := dbMongo.GetClientDetails()
	if errDb != nil {
		log.Fatal(errDb)
	}
	if len(details) == 0 {
		fmt.Println("no existing client positions")
		return
	}
	for _, client := range details {
		fmt.Printf("client_position: %d pubkey: %s name: %s\n",
			client.ClientPosition, client.Pubkey, client.ClientName)
	}
	fmt.Println()
}

// read client details and get client position
func clientPosition() int32 {
	// Read existing clients and get next available client position
	details, errDb := dbMongo.GetClientDetails()
	if errDb != nil {
		log.Fatal(errDb)
	}
	var maxClientPosition int32
	if len(details) == 0 {
		return 0
	}
	for _, client := range details {
		if client.ClientPosition > maxClientPosition {
			maxClientPosition = client.ClientPosition
		}
	}
	return maxClientPosition + 1
}

// main
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbMongo = server.NewDbMongo(ctx, mainConfig.DbConfig())

	fmt.Println()
	fmt.Println("*********************************************")
	fmt.Println("************ Client Signup Tool *************")
	fmt.Println("*********************************************")
	fmt.Println()
	printClientDetails()

	nextClientPosition := clientPosition()
	fmt.Printf("next available position: %d\n", nextClientPosition)
	fmt.Println()

	// Insert client pubkey details and verify
	fmt.Println("*********************************************")
	fmt.Println("************ Client Pubkey Info *************")
	fmt.Println("*********************************************")
	fmt.Println()
	fmt.Print("Insert pubkey: ")
	var pubKey string
	fmt.Scanln(&pubKey)
	pubKeyBytes, pubKeyBytesErr := hex.DecodeString(pubKey)
	if pubKeyBytesErr != nil {
		log.Fatal(pubKeyBytesErr)
	}
	_, errPub := btcec.ParsePubKey(pubKeyBytes, btcec.S256())
	if errPub != nil {
		log.Fatal(errPub)
	}
	fmt.Println("pubkey verified")
	fmt.Println()

	// New auth token ID for client
	fmt.Println("*********************************************")
	fmt.Println("***** Client Auth Token identification ******")
	fmt.Println("*********************************************")
	fmt.Println()
	uuid, err := uuid.NewV4()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("new-uuid: %s\n", uuid.String())
	fmt.Println()

	// Create new client details
	fmt.Println("*********************************************")
	fmt.Println("*********** Inserting New Client ************")
	fmt.Println("*********************************************")
	fmt.Println()
	fmt.Print("Insert client name: ")
	var clientName string
	fmt.Scanln(&clientName)
	newClientDetails := models.ClientDetails{
		ClientPosition: nextClientPosition,
		AuthToken:      uuid.String(),
		Pubkey:         pubKey,
		ClientName:     clientName}
	saveErr := dbMongo.SaveClientDetails(newClientDetails)
	if saveErr != nil {
		log.Fatal(saveErr)
	}
	fmt.Println("NEW CLIENT DETAILS")
	fmt.Printf("client_position: %d\n", newClientDetails.ClientPosition)
	fmt.Printf("auth_token: %s\n", newClientDetails.AuthToken)
	fmt.Printf("pubkey: %s\n", newClientDetails.Pubkey)
	fmt.Println()
	printClientDetails()
}
