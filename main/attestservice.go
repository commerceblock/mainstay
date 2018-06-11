package main

import (
    "log"
    "time"
    "sync"
    "context"
    "github.com/btcsuite/btcd/rpcclient"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
)

type AttestService struct {
    ctx             context.Context
    wg              *sync.WaitGroup
    mainClient      *rpcclient.Client
    attestClient    *AttestClient
    latestTxid      string
    awaitingConfirmation bool
}

func NewAttestService(ctx context.Context, wg *sync.WaitGroup, rpcMain *rpcclient.Client, rpcSide *rpcclient.Client, tx string, pk string) *AttestService{
    if (len(tx) != 64) {
        log.Fatal("Incorrect txid size")
    }
    attest := NewAttestClient(rpcMain, rpcSide, pk, tx)
    return &AttestService{ctx, wg, rpcMain, attest, "", false}
}

func (s *AttestService) Run() {
    defer s.wg.Done()

    s.getUnconfirmedTx()

    for {
        select {
            case <-s.ctx.Done():
                log.Println("Shutting down attestation service...")
                return
            // case receive confirmation request
                //
            default:
                s.doAttestation()
                time.Sleep(time.Second * 5)
        }
    }
}

// Find any previously unconfirmed transactions and start attestation from there
func (s *AttestService) getUnconfirmedTx() {
    log.Println("Looking for unconfirmed transactions")
    mempool, err := s.mainClient.GetRawMempool()
    if err != nil {
        log.Fatal(err)
    }

    for _, hash := range mempool {
        if (s.attestClient.verifyTxOnSubchain(hash.String())) {
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
func (s *AttestService) doAttestation() {
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
        success, txunspent := s.attestClient.findLastUnspent()
        if (success) {
            _ , paytoaddr := s.attestClient.getNextAttestationAddr()
            s.latestTxid = s.attestClient.sendAttestation(paytoaddr, txunspent)
            s.awaitingConfirmation = true
            log.Printf("Attestation committed")
        } else {
            log.Printf("Attestation unsuccessful")
        }
    }
}
