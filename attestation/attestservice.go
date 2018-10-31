/*
Package attestation implements the MainStay protocol.

Implemented using an Attestation Service structure that runs a main processes that
handles generating attestations and maintaining communication to a bitcoin wallet
*/
package attestation

import (
	"bytes"
	"context"
	"errors"
	"log"
	"sync"
	"time"

	confpkg "mainstay/config"
	"mainstay/messengers"
	"mainstay/models"
	"mainstay/server"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	zmq "github.com/pebbe/zmq4"
)

// Attestation state type
type AttestationState int

// Attestation states
const (
	ASTATE_NEW_ATTESTATION    AttestationState = 0
	ASTATE_UNCONFIRMED        AttestationState = 1
	ASTATE_CONFIRMED          AttestationState = 2
	ASTATE_COLLECTING_PUBKEYS AttestationState = 3
	ASTATE_COLLECTING_SIGS    AttestationState = 4
	ASTATE_FAILURE            AttestationState = 10
	ASTATE_INIT               AttestationState = 100
)

// Waiting time between attestations and/or attestation confirmation attempts
const ATTEST_WAIT_TIME = 60

// AttestationService structure
// Encapsulates Attest Client and connectivity
// to a Server for updates and requests
type AttestService struct {
	ctx         context.Context
	wg          *sync.WaitGroup
	config      *confpkg.Config
	attester    *AttestClient
	server      *server.Server
	publisher   *messengers.PublisherZmq
	subscribers []*messengers.SubscriberZmq
	state       AttestationState
}

var latestAttestation *models.Attestation // hold latest state
var attestDelay time.Duration             // initially 0 delay

var poller *zmq.Poller // poller to add all subscriber/publisher sockets

// NewAttestService returns a pointer to an AttestService instance
// Initiates Attest Client and Attest Server
func NewAttestService(ctx context.Context, wg *sync.WaitGroup, server *server.Server, config *confpkg.Config) *AttestService {
	if len(config.InitTX()) != 64 {
		log.Fatal("Incorrect txid size")
	}
	attester := NewAttestClient(config)

	latestAttestation = models.NewAttestationDefault()

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

	return &AttestService{ctx, wg, config, attester, server, publisher, subscribers, ASTATE_INIT}
}

// Run Attest Service
func (s *AttestService) Run() {
	defer s.wg.Done()

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

//Main attestation service method - cycles through AttestationStates
func (s *AttestService) doAttestation() {
	switch s.state {

	// ASTATE_INIT, ASTATE_FAILURE
	// - Check if there are unconfirmed or unspent transactions in the client
	// - Update server with latest attestation information
	// - If no transaction found wait, else initiate new attestation
	case ASTATE_INIT, ASTATE_FAILURE:
		log.Println("*AttestService* INIT ATTESTATION CLIENT AND SERVER")
		// find the state of the attestation
		// when a DB is put in place, these information will be collected from there
		unconfirmed, unconfirmedTxid, unconfirmedErr := s.attester.getUnconfirmedTx()
		if s.failureState(unconfirmedErr) {
			return
		} else if unconfirmed { // check mempool for unconfirmed - added check in case something gets rejected
			commitment, commitmentErr := s.server.GetAttestationCommitment(unconfirmedTxid)
			if s.failureState(commitmentErr) {
				return
			}
			latestAttestation = models.NewAttestation(unconfirmedTxid, commitment)
			s.state = ASTATE_UNCONFIRMED
		} else {
			success, txunspent, unspentErr := s.attester.findLastUnspent()
			if s.failureState(unspentErr) {
				return
			} else if success {
				txunspentHash, _ := chainhash.NewHashFromStr(txunspent.TxID)
				commitment, commitmentErr := s.server.GetAttestationCommitment(*txunspentHash)
				if s.failureState(commitmentErr) {
					return
				}
				latestAttestation = models.NewAttestation(*txunspentHash, commitment)
				s.state = ASTATE_CONFIRMED

				s.updateServer(*latestAttestation) // update server with latest attestation
				confirmedHash := latestAttestation.CommitmentHash()
				s.publisher.SendMessage((&confirmedHash).CloneBytes(), confpkg.TOPIC_CONFIRMED_HASH) // update clients
			} else {
				log.Println("*AttestService* No unconfirmed (mempool) or unspent tx found. Sleeping...")
				attestDelay = time.Duration(ATTEST_WAIT_TIME*10) * time.Second // add wait time
			}
		}

	// ASTATE_UNCONFIRMED
	// - Check for unconfirmed transactions in the mempool of the main client
	// - If confirmed, initiate new attestation
	case ASTATE_UNCONFIRMED:
		log.Printf("*AttestService* AWAITING CONFIRMATION \ntxid: (%s)\nhash: (%s)\n", latestAttestation.Txid.String(), latestAttestation.CommitmentHash().String())
		newTx, err := s.config.MainClient().GetTransaction(&latestAttestation.Txid)
		if s.failureState(err) {
			return
		}
		if newTx.BlockHash != "" {
			log.Printf("********** Attestation confirmed for txid: (%s)\n", latestAttestation.Txid.String())
			attestDelay = time.Duration(0) * time.Second

			s.state = ASTATE_CONFIRMED

			s.updateServer(*latestAttestation) // update server with latest attestation

			confirmedHash := latestAttestation.CommitmentHash()
			s.publisher.SendMessage((&confirmedHash).CloneBytes(), confpkg.TOPIC_CONFIRMED_HASH) //update clients
		}

	// ASTATE_CONFIRMED, ASTATE_NEW_ATTESTATION
	// - Either attestation failed or we are initiation the attestation chain
	// - Generate new hash for attestation and publish it to other signers
	case ASTATE_CONFIRMED, ASTATE_NEW_ATTESTATION:
		attestDelay = time.Duration(ATTEST_WAIT_TIME) * time.Second // add wait time

		unconfirmed, unconfirmedTxid, unconfirmedErr := s.attester.getUnconfirmedTx()
		if s.failureState(unconfirmedErr) {
			return
		} else if unconfirmed { // check mempool for unconfirmed - added check in case something gets rejected
			commitment, commitmentErr := s.server.GetAttestationCommitment(unconfirmedTxid)
			if s.failureState(commitmentErr) {
				return
			}
			latestAttestation = models.NewAttestation(unconfirmedTxid, commitment)
			s.state = ASTATE_UNCONFIRMED

			log.Printf("*AttestService* Waiting for confirmation of\ntxid: (%s)\nhash: (%s)\n", latestAttestation.Txid.String(), latestAttestation.CommitmentHash().String())
		} else {
			log.Println("*AttestService* NEW ATTESTATION")

			// get latest commitment hash from server
			latestCommitment, latestErr := s.server.GetLatestCommitment()
			if s.failureState(latestErr) {
				return
			}
			latestCommitmentHash := latestCommitment.GetCommitmentHash()

			log.Printf("********** hash: %s\n", latestCommitmentHash.String())
			if latestCommitmentHash == latestAttestation.CommitmentHash() { // skip attestation if same client hash
				log.Printf("********** Skipping attestation - Client hash already attested")
				return
			}
			// publish new attestation hash
			s.publisher.SendMessage((&latestCommitmentHash).CloneBytes(), confpkg.TOPIC_NEW_HASH)
			latestAttestation = models.NewAttestation(chainhash.Hash{}, &latestCommitment)
			s.state = ASTATE_COLLECTING_PUBKEYS // update attestation state
		}

	// ASTATE_COLLECTING_PUBKEYS
	// - Waiting for tweaked keys from signers
	// - Collect keys and generate address to pay the previous unspent to
	// - Create unsigned attestation transaction and publish it to other signers
	// - If successful await for sigs
	case ASTATE_COLLECTING_PUBKEYS:
		log.Println("*AttestService* COLLECTING PUBKEYS")
		attestDelay = time.Duration(ATTEST_WAIT_TIME) * time.Second // add wait time

		key, keyErr := s.attester.GetNextAttestationKey(latestAttestation.CommitmentHash())
		if s.failureState(keyErr) {
			return
		}

		paytoaddr, _ := s.attester.GetNextAttestationAddr(key, latestAttestation.CommitmentHash())
		importErr := s.attester.ImportAttestationAddr(paytoaddr)
		if s.failureState(importErr) {
			return
		}

		success, txunspent, unspentErr := s.attester.findLastUnspent()
		if s.failureState(unspentErr) {
			return
		} else if success {
			var createErr error
			var newTx *wire.MsgTx
			newTx, createErr = s.attester.createAttestation(paytoaddr, txunspent, false)
			latestAttestation.Tx = *newTx
			if s.failureState(createErr) {
				return
			}

			log.Printf("********** pre-sign txid: %s\n", latestAttestation.Tx.TxHash().String())

			// publish pre signed transaction
			var txbytes bytes.Buffer
			latestAttestation.Tx.Serialize(&txbytes)
			s.publisher.SendMessage(txbytes.Bytes(), confpkg.TOPIC_NEW_TX)

			s.state = ASTATE_COLLECTING_SIGS // update attestation state
		} else {
			s.failureState(errors.New("Attestation unsuccessful - No valid unspent found"))
			return
		}

	// ASTATE_COLLECTING_SIGS
	// - Waiting for signatures from signers
	// - Collect signatures, combine them and sign the attestation transaction
	// - If successful propagate the transaction and set state to unconfirmed
	case ASTATE_COLLECTING_SIGS:
		log.Println("*AttestService* COLLECTING SIGS")
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

		// get last attestation commitment from server
		latest, latestErr := s.server.GetLatestAttestation()
		if s.failureState(latestErr) {
			return
		}

		txid, attestationErr := s.attester.signAndSendAttestation(&latestAttestation.Tx, sigs, latest.CommitmentHash())
		if s.failureState(attestationErr) {
			return
		}

		log.Printf("********** Attestation committed for txid: (%s)\n", txid)
		latestAttestation.Txid = txid
		s.state = ASTATE_UNCONFIRMED
	}
}

// Method to update server with latest attestation
func (s *AttestService) updateServer(attestation models.Attestation) {
	//s.server.UpdateChan() <- *latestAttestation // send server update just to make sure it's up to date
	log.Println("*AttestService* Updating server with latest attestation")

	errUpdate := s.server.UpdateLatestAttestation(attestation)

	if errUpdate != nil {
		log.Fatal(errors.New("Server update failed"))
	} else {
		log.Println("*AttestService* Server update successful")
	}
}

// Set to error state and provide debugging information
func (s *AttestService) failureState(err error) bool {
	if err != nil {
		log.Println("*AttestService* ATTESTATION FAILURE")
		log.Println(err)
		s.state = ASTATE_FAILURE
		return true
	}
	return false
}
