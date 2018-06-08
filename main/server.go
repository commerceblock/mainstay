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
    latestTxid  string
    awaitingConfirmation bool
}

func NewServer(ctx context.Context, wg *sync.WaitGroup, rpcMain *rpcclient.Client, rpcOur *rpcclient.Client, tx string, pk string) *Server{
    if (len(tx) != 64) {
        log.Fatal("Incorrect txid size")
    }
    return &Server{ctx, wg, rpcMain, rpcOur, tx, pk, "", false}
}

func (s *Server) Run() {
    defer s.wg.Done()

    s.getUnconfirmedTx()

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
func (s *Server) verifyTxOnSubchain(txid string) bool {
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
        return s.verifyTxOnSubchain(prevtxid)
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
            if (s.verifyTxOnSubchain(vout.TxID)) { //theoretically only one unspent vout, but check anyway
                return true, vout
            }
        }
    }
    return false, btcjson.ListUnspentResult{}
}

// Find any previously unconfirmed transactions and start attestation from there
func (s *Server) getUnconfirmedTx() {
    log.Println("Looking for unconfirmed transactions")
    mempool, err := s.mainClient.GetRawMempool()
    if err != nil {
        log.Fatal(err)
    }

    for _, hash := range mempool {
        if (s.verifyTxOnSubchain(hash.String())) {
            s.latestTxid = hash.String()
            s.awaitingConfirmation = true
            log.Printf("Still waiting for confirmation of:\ntxid: (%s)\n", s.latestTxid)
            break
        }
    }
}

// Main attestation method. Two main functionalities
// - If we are not waiting for any confirmations get the last unspent vout, verify that
//  it is on the subchain and create/send a new transaction through the main client
// - If transaction has been sent, wait for confirmation on the next generated block
func (s *Server) doAttestation() {
    if (s.awaitingConfirmation) {
        log.Printf("Waiting for confirmation of:\ntxid: (%s)\n", s.latestTxid)

        txhash, err := chainhash.NewHashFromStr(s.latestTxid)
        if err != nil {
            log.Fatal(err)
        }

        newTx, err := s.mainClient.GetTransaction(txhash)
        if err != nil {
            log.Fatal(err)
        }
        if (newTx.BlockHash != "") {
            s.awaitingConfirmation = false
            log.Printf("Attestation %s confirmed\n", s.latestTxid)
        }
    } else {
        log.Println("Attempting attestation...")
        success, tx := s.findLastUnspent()
        if (success) {
            s.latestTxid = newTransaction(tx, s.mainClient)
            s.awaitingConfirmation = true
            log.Printf("Attestation committed")
        } else {
            log.Printf("Attestation unsuccessful")
        }
    }
}
