/*
Package attestation implements the MainStay protocol.

Implemented using an Attestation Service structure that runs multiple concurrent processes:
    - AttestClient that handles generating attestations and maintaining communication a bitcoin wallet
    - AttestServer that handles responding to API requests
    - A Requests channel to receive requests from requestapi
*/
package attestation

import (
	"context"
	"log"
	"mainstay/clients"
	"mainstay/models"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
)

// Waiting time between attestations and/or attestation confirmation attempts
const ATTEST_WAIT_TIME = 60

// AttestationService structure
// Encapsulates Attest Server, Attest Client
// and a channel for reading requests and writing responses
type AttestService struct {
	ctx        context.Context
	wg         *sync.WaitGroup
	mainClient *rpcclient.Client
	attester   *AttestClient
	server     *AttestServer
	channel    *models.Channel
}

var latestTx *Attestation
var attestDelay time.Duration // initially 0 delay

// NewAttestService returns a pointer to an AttestService instance
// Initiates Attest Client and Attest Server
func NewAttestService(ctx context.Context, wg *sync.WaitGroup, channel *models.Channel, rpcMain *rpcclient.Client, rpcSide clients.SidechainClient, cfgMain *chaincfg.Params, tx0 string) *AttestService {
	if len(tx0) != 64 {
		log.Fatal("Incorrect txid size")
	}
	attester := NewAttestClient(rpcMain, rpcSide, cfgMain, tx0)

	genesisHash, err := rpcSide.GetBlockHash(0)
	if err != nil {
		log.Fatal(err)
	}
	latestTx = &Attestation{chainhash.Hash{}, chainhash.Hash{}, true, time.Now()}
	server := NewAttestServer(rpcSide, *latestTx, tx0, *genesisHash)

	return &AttestService{ctx, wg, rpcMain, attester, server, channel}
}

// Run Attest Service
func (s *AttestService) Run() {
	defer s.wg.Done()

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
		timer := time.NewTimer(attestDelay)
		select {
		case <-s.ctx.Done():
			log.Println("Shutting down Attestation Service...")
			return
		case <-timer.C:
			s.doAttestation()
		}
	}
}

// Main attestation method. Three different states:
// - Check for unconfirmed transactions in the mempool of the main client
// - Attempt to do an attestation by searching for the last unspent and generating
//  a new TX where this unspent is the vin and the latest client hash is attested
// - Wait for confirmation of the attestation on the next main client block and
//  update information on the latest attestation when confirmation is received
func (s *AttestService) doAttestation() {
	if !latestTx.confirmed { // wait for confirmation
		log.Printf("*AttestService* Waiting for confirmation of\ntxid: (%s)\nhash: (%s)\n", latestTx.txid.String(), latestTx.attestedHash.String())
		newTx, err := s.mainClient.GetTransaction(&latestTx.txid)
		if err != nil {
			log.Fatal(err)
		}
		if newTx.BlockHash != "" {
			latestTx.latestTime = time.Now()
			latestTx.confirmed = true
			log.Printf("*AttestService* Attestation confirmed for txid: (%s)\n", latestTx.txid.String())
			s.server.UpdateLatest(*latestTx)
			log.Printf("**AttestServer** Updated latest attested height %d\n", s.server.latestHeight)
			attestDelay = time.Duration(0) * time.Second
		}
	} else {
		attestDelay = time.Duration(ATTEST_WAIT_TIME) * time.Second
		unconfirmed, unconfirmedTx := s.attester.getUnconfirmedTx()
		if unconfirmed { // check mempool for unconfirmed - added check in case something gets rejected
			*latestTx = unconfirmedTx
			log.Printf("*AttestService* Waiting for confirmation of\ntxid: (%s)\nhash: (%s)\n", latestTx.txid.String(), latestTx.attestedHash.String())
		} else {
			log.Println("*AttestService* Attempting attestation")
			success, txunspent := s.attester.findLastUnspent()
			if success {
				hash, paytoaddr := s.attester.getNextAttestationAddr()
				if hash == latestTx.attestedHash { // skip attestation if same client hash
					log.Printf("*AttestService* Skipping attestation - Client hash already attested")
					return
				}
				txid := s.attester.sendAttestation(paytoaddr, txunspent, false /* don't sure default fee */)
				latestTx = &Attestation{txid, hash, false, time.Now()}
				log.Printf("*AttestService* Attestation committed for txid: (%s)\n", txid)
			} else {
				log.Fatal("*AttestService* Attestation unsuccessful - No valid unspent found")
			}
		}
	}
}
