// Wallet transaction generation and general handling

package main

import (
    "log"
    "github.com/btcsuite/btcd/btcjson"
    "github.com/btcsuite/btcd/rpcclient"
    "github.com/btcsuite/btcutil"
    "github.com/davecgh/go-spew/spew"
)

func newTransaction(tx btcjson.ListUnspentResult, client *rpcclient.Client) (string){
    addr, err := client.GetNewAddress("")
    if err != nil {
        log.Fatal(err)
    }

    inputs := []btcjson.TransactionInput{
                    {Txid: tx.TxID, Vout: tx.Vout},
                }

    // estimate fee ?
    var fee btcutil.Amount = 1000 // in satoshi

    amounts := map[btcutil.Address]btcutil.Amount{addr: btcutil.Amount(tx.Amount*100000000)-fee}
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

    hash, err := client.SendRawTransaction(signedmsgtx, false)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("New tx hash %s\n", hash.String())

    return hash.String()
}
