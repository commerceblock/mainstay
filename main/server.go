package main

import (
    "log"
    "time"
    "sync"
    "context"
    "github.com/btcsuite/btcd/rpcclient"
    "github.com/btcsuite/btcd/btcjson"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
)

type Server struct {
    ctx         context.Context
    wg          *sync.WaitGroup
    mainClient  *rpcclient.Client
    ourClient   *rpcclient.Client
    txid0       string
    pk0         string
}

func NewServer(ctx context.Context, wg *sync.WaitGroup, rpcMain *rpcclient.Client, rpcOur *rpcclient.Client, tx string, pk string) *Server{
    if (len(tx) != 64) {
        log.Fatal("Incorrect txid size")
    }
    return &Server{ctx, wg, rpcMain, rpcOur, tx, pk}
}

func (s *Server) Run() {
    defer s.wg.Done()

    for {
        select {
            case <-s.ctx.Done():
                log.Println("Shutting down server...")
                return
            // case receive confirmation request
                //
            default:
                s.doAttestation()
                time.Sleep(time.Second * 5)
        }
    }
}

// Verify that an unspent vout is on the tip of the subchain attestations
func (s *Server) verifyUnspentOnSubchain(txid string) bool {
    if (txid == s.txid0) { // genesis transaction
        return true
    } else { //might be better to store subchain on init and no need to parse all transactions every time
        txhash, err := chainhash.NewHashFromStr(txid)
        if err != nil {
            log.Fatal(err)
        }

        txraw, err := s.mainClient.GetRawTransaction(txhash)
        if err != nil {
            return false
        }

        prevtxid := txraw.MsgTx().TxIn[0].PreviousOutPoint.Hash.String()
        return s.verifyUnspentOnSubchain(prevtxid)
    }
    return false
}

// Find the latest unspent vout that is on the tip of subchain attestations
func (s *Server) findLastUnspent() (bool, btcjson.ListUnspentResult) {
    unspent, err := s.mainClient.ListUnspent()
    if err != nil {
        log.Fatal(err)
    }
    if (len(unspent) > 0) {
        for _, vout := range unspent {
            if (s.verifyUnspentOnSubchain(vout.TxID) && //theoretically there should only be one unspent vout,
                vout.Amount > 50) { // scraping the fee change vouts for now
                return true, vout
            }
        }
    }
    return false, btcjson.ListUnspentResult{}
}

var latestTxid = ""
var latestAddr = ""
var awaitingConfirmation = false

func (s *Server) doAttestation() {
    if (awaitingConfirmation) {
        log.Printf("Waiting for confirmation of:\ntxid: (%s)\naddr: (%s)\n", latestTxid, latestAddr)

        txhash, err := chainhash.NewHashFromStr(latestTxid)
        if err != nil {
            log.Fatal(err)
        }

        newTx, err := s.mainClient.GetTransaction(txhash)
        if err != nil {
            log.Fatal(err)
        }
        if (newTx.BlockHash != "") {
            awaitingConfirmation = false
            log.Printf("Attestation %s confirmed\n", latestTxid)
        }
    } else {
        log.Println("Attempting attestation...")
        success, tx := s.findLastUnspent()
        if (success) {
            log.Printf("Attestation committed")
            latestTxid, latestAddr = newTransaction(tx, s.mainClient)
            awaitingConfirmation = true        
        } else {
            log.Printf("Attestation unsuccessful")
        }
    }
}
