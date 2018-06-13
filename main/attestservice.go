// Attestation service routine

package main

import (
    "log"
    "sync"
    "context"
    "time"
    "github.com/btcsuite/btcd/rpcclient"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
)

const genesis = "54e0d36bf404dc0a8cea4a080f8b485404c9001149cdd863e9350fa7f05fac53" //current testnet genesis

type AttestService struct {
    ctx             context.Context
    wg              *sync.WaitGroup
    mainClient      *rpcclient.Client
    attester        *AttestClient
    server          *AttestServer
    requests        chan Request
}

var latestTx *Attestation

func NewAttestService(ctx context.Context, wg *sync.WaitGroup, reqs chan Request, rpcMain *rpcclient.Client, rpcSide *rpcclient.Client, tx0 string, pk0 string) *AttestService{
    if (len(tx0) != 64) {
        log.Fatal("Incorrect txid size")
    }
    attester := NewAttestClient(rpcMain, rpcSide, pk0, tx0)

    latestTx = &Attestation{chainhash.Hash{}, chainhash.Hash{}, true, time.Now()}
    server := NewAttestServer(rpcSide, latestTx, tx0, genesis)

    return &AttestService{ctx, wg, rpcMain, attester, server, reqs}
}

func (s *AttestService) Run() {
    defer s.wg.Done()

    s.getUnconfirmedTx()

    s.wg.Add(1)
    go func() { //Waiting for requests from the request service and pass to server for response
        defer s.wg.Done()
        for {
            select {
                case <-s.ctx.Done():
                    return
                case req := <-s.requests:
                    s.requests <- s.server.Respond(req)
            }
        }
    }()

    for { //Doing attestations using attestation client and waiting for transaction confirmation
        select {
            case <-s.ctx.Done():
                log.Println("Shutting down Attestation Service...")
                return
            default:
                s.doAttestation()
        }
    }
}

// Find any previously unconfirmed transactions and start attestation from there
func (s *AttestService) getUnconfirmedTx() {
    log.Println("*AttestService* Looking for unconfirmed transactions")
    mempool, err := s.mainClient.GetRawMempool()
    if err != nil {
        log.Fatal(err)
    }

    for _, hash := range mempool {
        if (s.attester.verifyTxOnSubchain(*hash)) {
            latestTx = &Attestation{*hash, chainhash.Hash{}, false, time.Now()}
            log.Printf("*AttestService* Still waiting for confirmation of txid: (%s)\n", latestTx.txid)
            break
        }
    }
}

// Main attestation method. Two main functionalities
// - If we are not waiting for any confirmations get the last unspent vout, verify that
//  it is on the subchain and create/send a new transaction through the main client
// - If transaction has been sent, wait for confirmation on the next generated block
func (s *AttestService) doAttestation() {
    if (!latestTx.confirmed) {
        if (time.Since(latestTx.latestTime).Seconds() > 60) { // Only check for confirmation every 60 seconds
            log.Printf("*AttestService* Waiting for confirmation of txid: (%s)\n", latestTx.txid.String())
            newTx, err := s.mainClient.GetTransaction(&latestTx.txid)
            if err != nil {
                log.Fatal(err)
            }
            latestTx.latestTime = time.Now()
            if (newTx.BlockHash != "") {
                latestTx.confirmed = true
                s.server.AddConfirmed(*latestTx)
                log.Printf("*AttestService* Attestation confirmed for txid: (%s)\n", latestTx.txid.String())
            }
        }
    } else {
        log.Println("*AttestService* Attempting attestation")
        success, txunspent := s.attester.findLastUnspent()
        if (success) {
            hash, paytoaddr := s.attester.getNextAttestationAddr()
            txid := s.attester.sendAttestation(paytoaddr, txunspent)
            latestTx = &Attestation{txid, hash, false, time.Now()}
            log.Printf("*AttestService* Attestation committed for txid: (%s)\n", txid)
        } else {
            log.Printf("*AttestService* Attestation unsuccessful")
        }
    }
}
