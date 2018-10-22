/*
Package attestation implements the MainStay protocol.

Implemented using an Attestation Service structure that runs multiple concurrent processes:
    - AttestClient that handles generating attestations and maintaining communication a bitcoin wallet
    - AttestServer that handles responding to API requests
    - A Requests channel to receive requests from requestapi
*/
package attestation

import (
	"bytes"
	"context"
	"log"
	confpkg "mainstay/config"
	"mainstay/messengers"
	"mainstay/models"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	zmq "github.com/pebbe/zmq4"
)

// Waiting time between attestations and/or attestation confirmation attempts
const ATTEST_WAIT_TIME = 10

// AttestationService structure
// Encapsulates Attest Server, Attest Client
// and a channel for reading requests and writing responses
type AttestService struct {
	ctx         context.Context
	wg          *sync.WaitGroup
	config      *confpkg.Config
	attester    *AttestClient
	server      *AttestServer
	channel     *models.Channel
	publisher   *messengers.PublisherZmq
	subscribers []*messengers.SubscriberZmq
}

var latestAttestation *Attestation // hold latest state
var attestDelay time.Duration      // initially 0 delay

var poller *zmq.Poller // poller to add all subscriber/publisher sockets

// NewAttestService returns a pointer to an AttestService instance
// Initiates Attest Client and Attest Server
func NewAttestService(ctx context.Context, wg *sync.WaitGroup, channel *models.Channel, config *confpkg.Config) *AttestService {
	if len(config.InitTX()) != 64 {
		log.Fatal("Incorrect txid size")
	}
	attester := NewAttestClient(config)

	latestAttestation = NewAttestation(chainhash.Hash{}, chainhash.Hash{}, ASTATE_INIT)
	server := NewAttestServer(config.OceanClient(), *latestAttestation, config.InitTX())

	// Initialise publisher for sending new hashes and txs
	// and subscribers to receive sig responses
	poller = zmq.NewPoller()
	publisher := messengers.NewPublisherZmq(confpkg.MAIN_PUBLISHER_PORT, poller)
	var subscribers []*messengers.SubscriberZmq
	subtopics := []string{confpkg.TOPIC_SIGS}
	for _, nodeaddr := range config.MultisigNodes() {
		subscribers = append(subscribers, messengers.NewSubscriberZmq(nodeaddr, subtopics, poller))
	}
	attestDelay = 30 * time.Second // add some delay for subscribers to have time to set up

	return &AttestService{ctx, wg, config, attester, server, channel, publisher, subscribers}
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

// Main attestation method. States:
// ASTATE_INIT
// - Check if there are unconfirmed or unspent transactions in the client
// - Update server with latest attestation information
// - If no transaction found wait, else initiate new attestation
// ASTATE_UNCONFIRMED
// - Check for unconfirmed transactions in the mempool of the main client
// - If confirmed, initiate new attestation
// ASTATE_CONFIRMED, ASTATE_NEW_ATTESTATION
// - Either attestation failed or we are initiation the attestation chain
// - Generate new hash for attestation and publish it to other signers
// ASTATE_COLLECTING_PUBKEYS
// - Waiting for tweaked keys from signers
// - Collect keys and generate address to pay the previous unspent to
// - Create unsigned attestation transaction and publish it to other signers
// - If successful await for sigs
// ASTATE_COLLECTING_SIGS
// - Waiting for signatures from signers
// - Collect signatures, combine them and sign the attestation transaction
// - If successful propagate the transaction and set state to unconfirmed
func (s *AttestService) doAttestation() {
	switch latestAttestation.state {
	case ASTATE_INIT:
		log.Println("*AttestService* INIT ATTESTATION CLIENT AND SERVER")
		// find the state of the attestation
		// when a DB is put in place, these information will be collected from there
		unconfirmed, unconfirmedTx := s.attester.getUnconfirmedTx()
		if unconfirmed { // check mempool for unconfirmed - added check in case something gets rejected
			*latestAttestation = unconfirmedTx
			s.server.UpdateLatest(*latestAttestation)
		} else {
			success, txunspent := s.attester.findLastUnspent()
			if success {
				txunspentHash, _ := chainhash.NewHashFromStr(txunspent.TxID)
				s.server.UpdateLatest(*NewAttestation(*txunspentHash, s.attester.getTxAttestedHash(*txunspentHash), ASTATE_CONFIRMED))
				latestAttestation.state = ASTATE_NEW_ATTESTATION
			} else {
				log.Println("*AttestService* No unconfirmed (mempool) or unspent tx found. Sleeping...")
				attestDelay = time.Duration(ATTEST_WAIT_TIME*10) * time.Second // add wait time
			}
		}
	case ASTATE_UNCONFIRMED:
		log.Printf("*AttestService* AWAITING CONFIRMATION \ntxid: (%s)\nhash: (%s)\n", latestAttestation.txid.String(), latestAttestation.attestedHash.String())
		newTx, err := s.config.MainClient().GetTransaction(&latestAttestation.txid)
		if err != nil {
			log.Fatal(err)
		}
		if newTx.BlockHash != "" {
			latestAttestation.latestTime = time.Now()
			latestAttestation.state = ASTATE_CONFIRMED
			log.Printf("********** Attestation confirmed for txid: (%s)\n", latestAttestation.txid.String())
			s.server.UpdateLatest(*latestAttestation)
			log.Printf("********** Updated latest attested height %d\n", s.server.latestHeight)
			attestDelay = time.Duration(0) * time.Second
		}
	case ASTATE_CONFIRMED, ASTATE_NEW_ATTESTATION:
		attestDelay = time.Duration(ATTEST_WAIT_TIME) * time.Second // add wait time

		unconfirmed, unconfirmedTx := s.attester.getUnconfirmedTx()
		if unconfirmed { // check mempool for unconfirmed - added check in case something gets rejected
			*latestAttestation = unconfirmedTx
			log.Printf("*AttestService* Waiting for confirmation of\ntxid: (%s)\nhash: (%s)\n", latestAttestation.txid.String(), latestAttestation.attestedHash.String())
		} else {
			log.Println("*AttestService* NEW ATTESTATION")
			clientListener := NewListener(s.config.OceanClient())
			hash := clientListener.GetNextHash()
			log.Printf("********** hash: %s\n", hash.String())
			if hash == latestAttestation.attestedHash { // skip attestation if same client hash
				log.Printf("********** Skipping attestation - Client hash already attested")
				return
			}
			// publish new attestation hash
			s.publisher.SendMessage(hash.CloneBytes(), confpkg.TOPIC_NEW_HASH)
			latestAttestation = NewAttestation(chainhash.Hash{}, hash, ASTATE_COLLECTING_PUBKEYS) // update attestation state
		}
	case ASTATE_COLLECTING_PUBKEYS:
		log.Printf("*AttestService* COLLECTING PUBKEYS\n")
		attestDelay = time.Duration(ATTEST_WAIT_TIME) * time.Second // add wait time

		key := s.attester.GetNextAttestationKey(latestAttestation.attestedHash)
		paytoaddr, script := s.attester.GetNextAttestationAddr(key, latestAttestation.attestedHash)

		success, txunspent := s.attester.findLastUnspent()
		if success {
			latestAttestation.tx = *s.attester.createAttestation(paytoaddr, txunspent, false)
			latestAttestation.txunspent = txunspent
			latestAttestation.redeemScript = script
			log.Printf("********** pre-sign txid: %s\n", latestAttestation.tx.TxHash().String())

			// publish pre signed transaction
			var txbytes bytes.Buffer
			latestAttestation.tx.Serialize(&txbytes)
			s.publisher.SendMessage(txbytes.Bytes(), confpkg.TOPIC_NEW_TX)

			latestAttestation.state = ASTATE_COLLECTING_SIGS // update attestation state
		} else {
			log.Fatal("Attestation unsuccessful - No valid unspent found")
		}
	case ASTATE_COLLECTING_SIGS:
		log.Printf("*AttestService* COLLECTING SIGS\n")
		attestDelay = time.Duration(ATTEST_WAIT_TIME) * time.Second // add wait time
		// Read sigs using subscribers
		var sigs [][]byte
		sockets, _ := poller.Poll(-1)
		for _, socket := range sockets {
			for _, sub := range s.subscribers {
				if sub.Socket() == socket.Socket {
					_, msg := sub.ReadMessage()
					sigs = append(sigs, msg)
				}
			}
		}
		txid := s.attester.signAndSendAttestation(&latestAttestation.tx, sigs, s.server.latest.attestedHash)
		log.Printf("********** Attestation committed for txid: (%s)\n", txid)
		latestAttestation.txid = txid
		latestAttestation.state = ASTATE_UNCONFIRMED
	}
}
