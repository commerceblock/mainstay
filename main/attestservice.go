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

type AttestService struct {
    ctx             context.Context
    wg              *sync.WaitGroup
    mainClient      *rpcclient.Client
    attester        *AttestClient
    requests        chan Request
    latestTxid      string
    isWaitingConfirm    bool
    waitingStartTime    time.Time
}

func NewAttestService(ctx context.Context, wg *sync.WaitGroup, reqs chan Request, rpcMain *rpcclient.Client, rpcSide *rpcclient.Client, tx string, pk string) *AttestService{
    if (len(tx) != 64) {
        log.Fatal("Incorrect txid size")
    }
    attest := NewAttestClient(rpcMain, rpcSide, pk, tx)
    return &AttestService{ctx, wg, rpcMain, attest, reqs, "", false, time.Now()}
}

func (s *AttestService) Run() {
    defer s.wg.Done()

    s.getUnconfirmedTx()

    s.wg.Add(1)
    go func() { //Waiting for requests from the confirmation service
        defer s.wg.Done()
        for {
            select {
                case <-s.ctx.Done():
                    return
                case req := <-s.requests:
                    log.Printf("*attest* -- request -- %s\n", req)
                    req.Attested = true
                    s.requests <- req
            }
        }
    }()

    for { //Doing attestations and waiting for transaction confirmation
        select {
            case <-s.ctx.Done():
                log.Println("Shutting down attestation service...")
                return
            default:
                s.doAttestation()
        }
    }
}

// Find any previously unconfirmed transactions and start attestation from there
func (s *AttestService) getUnconfirmedTx() {
    log.Println("*attest* Looking for unconfirmed transactions")
    mempool, err := s.mainClient.GetRawMempool()
    if err != nil {
        log.Fatal(err)
    }

    for _, hash := range mempool {
        if (s.attester.verifyTxOnSubchain(hash.String())) {
            s.latestTxid = hash.String()
            s.isWaitingConfirm = true
            s.waitingStartTime = time.Now()
            log.Printf("*attest* Still waiting for confirmation of:\ntxid: (%s)\n", s.latestTxid)
            break
        }
    }
}

// Main attestation method. Two main functionalities
// - If we are not waiting for any confirmations get the last unspent vout, verify that
//  it is on the subchain and create/send a new transaction through the main client
// - If transaction has been sent, wait for confirmation on the next generated block
func (s *AttestService) doAttestation() {
    if (s.isWaitingConfirm) {
        if (time.Since(s.waitingStartTime).Seconds() > 60) { // Only check for confirmation every 60 seconds
            log.Printf("*attest* Waiting for confirmation of:\ntxid: (%s)\n", s.latestTxid)
            txhash, err := chainhash.NewHashFromStr(s.latestTxid)
            if err != nil {
                log.Fatal(err)
            }

            newTx, err := s.mainClient.GetTransaction(txhash)
            if err != nil {
                log.Fatal(err)
            }
            if (newTx.BlockHash != "") {
                s.isWaitingConfirm = false
                log.Printf("*attest* Attestation %s confirmed\n", s.latestTxid)
            } else {
                s.waitingStartTime = time.Now()
            }
        }
    } else {
        log.Println("*attest* Attempting attestation")
        success, txunspent := s.attester.findLastUnspent()
        if (success) {
            _ , paytoaddr := s.attester.getNextAttestationAddr()
            s.latestTxid = s.attester.sendAttestation(paytoaddr, txunspent)
            log.Printf("*attest* New tx hash %s\n", s.latestTxid)
            s.isWaitingConfirm = true
            s.waitingStartTime = time.Now()
            log.Printf("*attest* Attestation committed - Waiting for confirmation")
        } else {
            log.Printf("*attest* Attestation unsuccessful")
        }
    }
}
