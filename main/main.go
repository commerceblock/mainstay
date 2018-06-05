package main

import (
	"log"
	"ocean-attestation/conf"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/davecgh/go-spew/spew"
)

const TX_ID_0 	= "fa4e6d5c18a8b60999fa2f798ed94023b12f7e138647d9edcd0cf1d965a6dcba"
const PK_0		= "21037dd3a0a572179f61ea625765759bacb353d0387db60a06ae1997b44422444b88ac"

var mainClient *rpcclient.Client
var oceanClient *rpcclient.Client

func initialise() {
	mainClient 	= conf.GetRPC("main")
	oceanClient = conf.GetRPC("ocean")
}

func main() {
	initialise()

	// Get the latest unspent and generate a new transaction
	unspent, err := mainClient.ListUnspent()
	if err != nil {
		log.Fatal(err)
	}
	tx := unspent[0]
	log.Printf("First utxo:\n%v", spew.Sdump(tx))
	newTransaction(tx, mainClient)
	
	defer mainClient.Shutdown()
	defer oceanClient.Shutdown()
}
