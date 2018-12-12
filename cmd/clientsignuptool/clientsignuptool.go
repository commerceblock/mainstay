// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

// Client signup tool

import (
	"context"
	"fmt"
	"log"
	"os"

	"mainstay/config"
	"mainstay/models"
	"mainstay/server"

	"github.com/btcsuite/btcutil"
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

// read client details and get client position
func clientPosition() int32 {
	// Read existing clients and get next available client position
	fmt.Println("existing clients")
	details, errDb := dbMongo.GetClientDetails()
	if errDb != nil {
		log.Fatal(errDb)
	}
	var maxClientPosition int32
	if len(details) == 0 {
		fmt.Println("no existing client positions")
		return 0
	} else {
		for _, client := range details {
			if client.ClientPosition > maxClientPosition {
				maxClientPosition = client.ClientPosition
			}
			fmt.Printf("client_position: %d pubkey: %s\n", client.ClientPosition, client.Pubkey)
		}
	}
	fmt.Println()
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

	nextClientPosition := clientPosition()
	fmt.Printf("next available position: %d\n", nextClientPosition)
	fmt.Println()

	// Insert client pubkey details and verify
	fmt.Println("*********************************************")
	fmt.Println("************ Client Pubkey Address info *************")
	fmt.Println("*********************************************")
	fmt.Println()
	fmt.Print("Insert pubkey hash address: ")
	var addr string
	fmt.Scanln(&addr)
	_, errAddr := btcutil.DecodeAddress(addr, mainConfig.MainChainCfg()) // encode to address
	if errAddr != nil {
		log.Fatal(errAddr)
	}
	fmt.Println("addr verified")
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
	newClientDetails := models.ClientDetails{ClientPosition: nextClientPosition, AuthToken: uuid.String(), Pubkey: addr}
	saveErr := dbMongo.SaveClientDetails(newClientDetails)
	if saveErr != nil {
		log.Fatal(saveErr)
	}
	fmt.Println("NEW CLIENT DETAILS")
	fmt.Printf("client_position: %d\n", newClientDetails.ClientPosition)
	fmt.Printf("auth_token: %s\n", newClientDetails.AuthToken)
	fmt.Printf("pubkey: %s\n", newClientDetails.Pubkey)
	fmt.Println()
	clientPosition()
}
