// Wallet transaction generation and general handling

package main

import (
	"log"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	"github.com/davecgh/go-spew/spew"
)

func newTransaction(tx btcjson.ListUnspentResult, client *rpcclient.Client) {

	inputs := []btcjson.TransactionInput{
					{Txid: tx.TxID, Vout: tx.Vout},
				}
	
	addr, err := client.GetNewAddress("")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Addr: %s", addr.String())

	amounts := map[btcutil.Address]btcutil.Amount{addr: btcutil.Amount(tx.Amount)}
	msgtx, err := client.CreateRawTransaction(inputs, amounts, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("New tx:\n%v", spew.Sdump(msgtx))

	signedmsgtx, issigned, err := client.SignRawTransaction(msgtx)
	if err != nil || !issigned {
		log.Fatal(err)
	}
	log.Printf("Signed tx:\n%v", spew.Sdump(signedmsgtx))

	// TODO: add fees manually

	txhash, err := client.SendRawTransaction(signedmsgtx, false)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("New tx hash %s", txhash.String())
}
