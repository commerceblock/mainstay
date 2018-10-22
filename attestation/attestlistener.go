package attestation

import (
	"log"
    "context"
    "sync"
	"mainstay/clients"
    "mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Attest listener struct
type AttestListener struct {
    ctx         context.Context
    wg          *sync.WaitGroup
	client      clients.SidechainClient
    requestChan chan models.RequestWithResponseChannel
    latestChan  chan models.RequestWithResponseChannel
}

// Return new AttestListener instance with channels for receiving client
// requests and latest attestation querying
func NewAttestListener(ctx context.Context, wg *sync.WaitGroup, client clients.SidechainClient) *AttestListener {
    reqChan := make(chan models.RequestWithResponseChannel)
    latestChan := make(chan models.RequestWithResponseChannel)
    return &AttestListener{ctx, wg, client, reqChan, latestChan}
}

// Verify incoming client request
func (l *AttestListener) verifyRequest(req models.Request) interface {} {

    // read hash from request
    // read id
    // verify id
    // update latest
    // return ACK

    return nil
}

// Return latest hash
func (l *AttestListener) getNextHash() chainhash.Hash {
    hash, err := l.client.GetBestBlockHash()
    if err != nil {
        log.Fatal(err)
    }
    return *hash
}

// Main method listening to incoming requests and replying
func (l *AttestListener) Run() {
    defer l.wg.Done()

    for {
        select {
        case req := <- l.requestChan:
            req.Response <- l.verifyRequest(req.Request)
        case latest := <- l.latestChan:
            latest.Response <- l.getNextHash()
        case <-l.ctx.Done():
            return
        }
    }
}
