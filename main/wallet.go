// Wallet transaction generation and general handling

package main

import (
    "log"
    "github.com/btcsuite/btcd/btcjson"
    "github.com/btcsuite/btcd/rpcclient"
    "github.com/btcsuite/btcutil"
)

const FEE_PER_BYTE = 20 // satoshis

func newTransaction(tx btcjson.ListUnspentResult, client *rpcclient.Client) (string){
    addr, err := client.GetNewAddress("")
    if err != nil {
        log.Fatal(err)
    }

    inputs := []btcjson.TransactionInput{{Txid: tx.TxID, Vout: tx.Vout},}

    amounts := map[btcutil.Address]btcutil.Amount{addr: btcutil.Amount(tx.Amount*100000000)}
    msgtx, err := client.CreateRawTransaction(inputs, amounts, nil)
    if err != nil {
        log.Fatal(err)
    }

    fee := int64(FEE_PER_BYTE * msgtx.SerializeSize())
    msgtx.TxOut[0].Value -= fee
    log.Printf("Tx fee %d", fee)

    signedmsgtx, issigned, err := client.SignRawTransaction(msgtx)
    if err != nil || !issigned {
        log.Fatal(err)
    }

    txhash, err := client.SendRawTransaction(signedmsgtx, false)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("New tx hash %s\n", txhash.String())

    return txhash.String()
}
