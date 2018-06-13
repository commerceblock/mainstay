// Wallet transaction generation and general handling

package main

import (
    "log"
    "github.com/btcsuite/btcd/btcjson"
    "github.com/btcsuite/btcd/rpcclient"
    "github.com/btcsuite/btcutil"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
)

const FEE_PER_BYTE = 20 // satoshis

type AttestClient struct {
    mainClient  *rpcclient.Client
    sideClient  *rpcclient.Client
    pk0         string
    txid0       string
}

func NewAttestClient(rpcMain *rpcclient.Client, rpcSide *rpcclient.Client, pk string, tx string) *AttestClient {
    if (len(pk) != 64) {
        log.Fatal("*AttestClient* Incorrect key size")
    }
    return &AttestClient{rpcMain, rpcSide, pk, tx}
}

// Get next attestation address by tweaking initial private key with current sidechain block hash
func (w *AttestClient) getNextAttestationAddr() (chainhash.Hash, btcutil.Address) {
    addr, err := w.mainClient.GetNewAddress("")
    if err != nil {
        log.Fatal(err)
    }

    hash, err := w.sideClient.GetBestBlockHash()
    if err != nil {
        log.Fatal(err)
    }

    return *hash, addr
}

// Generate a new transaction paying to the tweaked address, add fees and send the transaction through the wallet client
func (w *AttestClient) sendAttestation(paytoaddr btcutil.Address, txunspent btcjson.ListUnspentResult) (chainhash.Hash) {
    inputs := []btcjson.TransactionInput{{Txid: txunspent.TxID, Vout: txunspent.Vout},}

    amounts := map[btcutil.Address]btcutil.Amount{paytoaddr: btcutil.Amount(txunspent.Amount*100000000)}
    msgtx, err := w.mainClient.CreateRawTransaction(inputs, amounts, nil)
    if err != nil {
        log.Fatal(err)
    }

    fee := int64(FEE_PER_BYTE * msgtx.SerializeSize())
    msgtx.TxOut[0].Value -= fee
    log.Printf("*AttestClient* Tx fee %d", fee)

    signedmsgtx, issigned, err := w.mainClient.SignRawTransaction(msgtx)
    if err != nil || !issigned {
        log.Fatal(err)
    }

    txhash, err := w.mainClient.SendRawTransaction(signedmsgtx, false)
    if err != nil {
        log.Fatal(err)
    }
    return *txhash
}

// Verify that an unspent vout is on the tip of the subchain attestations
func (w *AttestClient) verifyTxOnSubchain(txid chainhash.Hash) bool {
    if (txid.String() == w.txid0) { // genesis transaction
        return true
    } else { //might be better to store subchain on init and no need to parse all transactions every time
        txraw, err := w.mainClient.GetRawTransaction(&txid)
        if err != nil {
            return false
        }

        prevtxid := txraw.MsgTx().TxIn[0].PreviousOutPoint.Hash
        return w.verifyTxOnSubchain(prevtxid)
    }
    return false
}

// Find the latest unspent vout that is on the tip of subchain attestations
func (w *AttestClient) findLastUnspent() (bool, btcjson.ListUnspentResult) {
    unspent, err := w.mainClient.ListUnspent()
    if err != nil {
        log.Fatal(err)
    }
    if (len(unspent) > 0) {
        for _, vout := range unspent {
            txhash, err := chainhash.NewHashFromStr(vout.TxID)
            if err != nil {
                log.Fatal(err)
            }
            if (w.verifyTxOnSubchain(*txhash)) { //theoretically only one unspent vout, but check anyway
                return true, vout
            }
        }
    }
    return false, btcjson.ListUnspentResult{}
}
