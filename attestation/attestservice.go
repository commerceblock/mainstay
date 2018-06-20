// Attestation service routine

package attestation

import (
    "log"
    "sync"
    "context"
    "time"
    "github.com/btcsuite/btcd/rpcclient"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
    "github.com/btcsuite/btcd/chaincfg"
    "ocean-attestation/models"
)

type AttestService struct {
    ctx             context.Context
    wg              *sync.WaitGroup
    mainClient      *rpcclient.Client
    attester        *AttestClient
    server          *AttestServer
    channel         *models.Channel
}

var latestTx *Attestation

func NewAttestService(ctx context.Context, wg *sync.WaitGroup, channel *models.Channel, rpcMain *rpcclient.Client, rpcSide *rpcclient.Client, cfgMain *chaincfg.Params, tx0 string, pk0 string) *AttestService{
    if (len(tx0) != 64) {
        log.Fatal("Incorrect txid size")
    }
    attester := NewAttestClient(rpcMain, rpcSide, cfgMain, pk0, tx0)

    genesis, err := rpcSide.GetBlockHash(0)
    if err!=nil {
        log.Fatal(err)
    }
    latestTx = &Attestation{chainhash.Hash{}, chainhash.Hash{}, true, time.Now()}
    server := NewAttestServer(rpcSide, *latestTx, tx0, *genesis)

    return &AttestService{ctx, wg, rpcMain, attester, server, channel}
}

func (s *AttestService) Run() {
    defer s.wg.Done()

    log.Printf("*AttestService* genesis tx %s\n", s.attester.txid0)
    log.Println("*AttestService* Looking for unconfirmed transactions")
    s.attester.getUnconfirmedTx(latestTx)

    s.wg.Add(1)
    go func() { //Waiting for requests from the request service and pass to server for response
        defer s.wg.Done()
        for {
            select {
                case <-s.ctx.Done():
                    return
                case req := <-s.channel.Requests:
                    s.channel.Responses <- s.server.Respond(req)
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
                log.Printf("*AttestService* Attestation confirmed for txid: (%s)\n", latestTx.txid.String())
                s.server.UpdateLatest(*latestTx)
                log.Printf("**AttestServer** Updated latest attested height %d\n", s.server.latestHeight)
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
