// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

// Client signup tool

import (
	"bufio"
	"context"
	"encoding/hex"
	"os"
	"fmt"

	"mainstay/config"
	"mainstay/db"
	"mainstay/models"
	"mainstay/log"

	"github.com/btcsuite/btcd/btcec"
	"github.com/satori/go.uuid"
)

const ConfPath = "/src/mainstay/cmd/clientsignuptool/conf.json"

var (
	mainConfig *config.Config
	dbMongo    *db.DbMongo
)

// init
func init() {
	confFile, confErr := config.GetConfFile(os.Getenv("GOPATH") + ConfPath)
	if confErr != nil {
		log.Error(confErr)
	}
	var mainConfigErr error
	mainConfig, mainConfigErr = config.NewConfig(confFile)
	if mainConfigErr != nil {
		log.Error(mainConfigErr)
	}
}

// print client details
func printClientDetails() {
	// Read existing clients and get next available client position
	log.Infoln("existing clients")
	details, errDb := dbMongo.GetClientDetails()
	if errDb != nil {
		log.Error(errDb)
	}
	if len(details) == 0 {
		log.Infoln("no existing client positions")
		return
	}
	for _, client := range details {
		log.Infof("client_position: %d pubkey: %s name: %s\n",
			client.ClientPosition, client.Pubkey, client.ClientName)
	}
	log.Infoln()
}

// read client details and get client position
func clientPosition() int32 {
	// Read existing clients and get next available client position
	details, errDb := dbMongo.GetClientDetails()
	if errDb != nil {
		log.Error(errDb)
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

	dbMongo = db.NewDbMongo(ctx, mainConfig.DbConfig())

	log.Infoln()
	log.Infoln("*********************************************")
	log.Infoln("************ Client Signup Tool *************")
	log.Infoln("*********************************************")
	log.Infoln()
	printClientDetails()

	nextClientPosition := clientPosition()
	log.Infof("next available position: %d\n", nextClientPosition)
	log.Infoln()

	// Insert client pubkey details and verify
	log.Infoln("*********************************************")
	log.Infoln("************ Client Pubkey Info *************")
	log.Infoln("*********************************************")
	log.Infoln()
	log.Info("Insert pubkey (optional): ")
	var pubKey string
	fmt.Scanln(&pubKey)
	if pubKey == "" {
		log.Infoln("no pubkey authentication")
	} else {
		pubKeyBytes, pubKeyBytesErr := hex.DecodeString(pubKey)
		if pubKeyBytesErr != nil {
			log.Error(pubKeyBytesErr)
		}
		_, errPub := btcec.ParsePubKey(pubKeyBytes, btcec.S256())
		if errPub != nil {
			log.Error(errPub)
		}
		log.Infoln("pubkey verified")
		log.Infoln()
	}

	// New auth token ID for client
	log.Infoln("*********************************************")
	log.Infoln("***** Client Auth Token identification ******")
	log.Infoln("*********************************************")
	log.Infoln()
	uuid, err := uuid.NewV4()
	if err != nil {
		log.Error(err)
	}
	log.Infof("new-uuid: %s\n", uuid.String())
	log.Infoln()

	// Create new client details
	log.Infoln("*********************************************")
	log.Infoln("*********** Inserting New Client ************")
	log.Infoln("*********************************************")
	log.Infoln()
	log.Info("Insert client name: ")

	// scan input client name
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	clientName := scanner.Text()

	newClientDetails := models.ClientDetails{
		ClientPosition: nextClientPosition,
		AuthToken:      uuid.String(),
		Pubkey:         pubKey,
		ClientName:     clientName}
	saveErr := dbMongo.SaveClientDetails(newClientDetails)
	if saveErr != nil {
		log.Error(saveErr)
	}
	log.Infoln("NEW CLIENT DETAILS")
	log.Infof("client_position: %d\n", newClientDetails.ClientPosition)
	log.Infof("auth_token: %s\n", newClientDetails.AuthToken)
	log.Infof("pubkey: %s\n", newClientDetails.Pubkey)
	log.Infoln()
	printClientDetails()
}
