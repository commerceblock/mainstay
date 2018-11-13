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

// Attestation Service is the main processes that handles generating
// attestations and maintaining communication to a bitcoin wallet

// Attestation state type
type AttestationState int

// Attestation states
const (
	ASTATE_ERROR              AttestationState = -1
	ASTATE_INIT               AttestationState = 0
	ASTATE_NEXT_COMMITMENT    AttestationState = 1
	ASTATE_NEW_ATTESTATION    AttestationState = 2
	ASTATE_SIGN_ATTESTATION   AttestationState = 3
	ASTATE_SEND_ATTESTATION   AttestationState = 4
	ASTATE_AWAIT_CONFIRMATION AttestationState = 5
	ASTATE_HANDLE_UNCONFIRMED AttestationState = 6
)

// error consts
const (
	ERROR_UNSPENT_NOT_FOUND = "No valid unspent found"
)

// waiting time schedules
const (
	// fixed waiting time between states
	ATIME_FIXED = 5 * time.Second

	// waiting time for sigs to arrive from multisig nodes
	ATIME_SIGS = 1 * time.Minute

	// waiting time between attemps to check if an attestation has been confirmed
	ATIME_CONFIRMATION = 15 * time.Minute

	// waiting time between consecutive attestations after one was confirmed
	ATIME_NEW_ATTESTATION = 60 * time.Minute

	// waiting time until we handle an attestation that has not been confirmed
	// usually by increasing the fee of the previous transcation to speed up confirmation
	ATIME_HANDLE_UNCONFIRMED = 60 * time.Minute
)

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
	attestation *models.Attestation
	errorState  error
}

var attestDelay time.Duration // delay between states
var confirmTime time.Time     // delay untill confirmation

var poller *zmq.Poller // poller to add all subscriber/publisher sockets

// NewAttestService returns a pointer to an AttestService instance
// Initiates Attest Client and Attest Server
func NewAttestService(ctx context.Context, wg *sync.WaitGroup, server *server.Server, config *confpkg.Config) *AttestService {
	// Check init txid validity
	_, errInitTx := chainhash.NewHashFromStr(config.InitTX())
	if errInitTx != nil {
		log.Fatalf("Incorrect initial transaction id %s\n", config.InitTX())
	}

	// initiate attestation client
	attester := NewAttestClient(config)

	// Initialise publisher for sending new hashes and txs
	// and subscribers to receive sig responses
	poller = zmq.NewPoller()
	publisher := messengers.NewPublisherZmq(confpkg.MAIN_PUBLISHER_PORT, poller)
	var subscribers []*messengers.SubscriberZmq
	subtopics := []string{confpkg.TOPIC_SIGS}
	for _, nodeaddr := range config.MultisigNodes() {
		subscribers = append(subscribers, messengers.NewSubscriberZmq(nodeaddr, subtopics, poller))
	}

	return &AttestService{ctx, wg, config, attester, server, publisher, subscribers, ASTATE_INIT, models.NewAttestationDefault(), nil}
}

// Run Attest Service
func (s *AttestService) Run() {
	defer s.wg.Done()

	attestDelay = 30 * time.Second // add some delay for subscribers to have time to set up

	for { //Doing attestations using attestation client and waiting for transaction confirmation
		timer := time.NewTimer(attestDelay)
		select {
		case <-s.ctx.Done():
			log.Println("Shutting down Attestation Service...")
			return
		case <-timer.C:
			s.doAttestation()

			// for testing
			// attestDelay = 1*time.Second
		}
	}
}

// ASTATE_ERROR
// - Print error state and re-initiate attestation
func (s *AttestService) doStateError() {
	log.Println("*AttestService* ATTESTATION SERVICE FAILURE")
	log.Println(s.errorState)
	s.state = ASTATE_INIT
}

// ASTATE_INIT
// - Check if there are unconfirmed or unspent transactions in the client
// - Update server with latest attestation information
// - If no transaction found wait, else initiate new attestation
// - If no attestation found, check last unconfirmed from db
func (s *AttestService) doStateInit() {
	log.Println("*AttestService* INITIATING ATTESTATION PROCESS")

	// find the state of the attestation
	unconfirmed, unconfirmedTxid, unconfirmedErr := s.attester.getUnconfirmedTx()
	if s.setFailure(unconfirmedErr) {
		return // will rebound to init
	} else if unconfirmed { // check mempool for unconfirmed - added check in case something gets rejected
		commitment, commitmentErr := s.server.GetAttestationCommitment(unconfirmedTxid)
		if s.setFailure(commitmentErr) {
			return // will rebound to init
		}
		s.attestation = models.NewAttestation(unconfirmedTxid, &commitment) // initialise attestation
		s.state = ASTATE_AWAIT_CONFIRMATION                                 // update attestation state
	} else {
		success, unspent, unspentErr := s.attester.findLastUnspent()
		if s.setFailure(unspentErr) {
			return // will rebound to init
		} else if success {
			unspentTxid, _ := chainhash.NewHashFromStr(unspent.TxID)
			commitment, commitmentErr := s.server.GetAttestationCommitment(*unspentTxid)
			if s.setFailure(commitmentErr) {
				return // will rebound to init
			} else if (commitment.GetCommitmentHash() != chainhash.Hash{}) {
				s.attestation = models.NewAttestation(*unspentTxid, &commitment)
				// update server with latest confirmed attestation
				s.attestation.Confirmed = true
				errUpdate := s.server.UpdateLatestAttestation(*s.attestation)
				if s.setFailure(errUpdate) {
					return // will rebound to init
				}
			} else {
				s.attestation = models.NewAttestationDefault()
			}
			confirmedHash := s.attestation.CommitmentHash()
			s.publisher.SendMessage((&confirmedHash).CloneBytes(), confpkg.TOPIC_CONFIRMED_HASH) // update clients

			s.state = ASTATE_NEXT_COMMITMENT // update attestation state
		} else {
			// no unspent so there must be a transaction waiting confirmation not on the mempool
			// check server for latest unconfirmed attestation
			lastCommitmentHash, latestErr := s.server.GetLatestAttestationCommitmentHash(false)
			if s.setFailure(latestErr) {
				return // will rebound to init
			}
			commitment, commitmentErr := s.server.GetAttestationCommitment(lastCommitmentHash)
			if s.setFailure(commitmentErr) {
				return // will rebound to init
			}
			s.attestation = models.NewAttestation(lastCommitmentHash, &commitment) // initialise attestation
			s.state = ASTATE_AWAIT_CONFIRMATION                                    // update attestation state
		}
	}
}

// ASTATE_NEXT_COMMITMENT
// - Get latest commitment from server
// - Check if commitment has already been attested
// - Send commitment to client signers
// - Initialise new attestation
func (s *AttestService) doStateNextCommitment() {
	log.Println("*AttestService* NEW ATTESTATION COMMITMENT")

	// get latest commitment hash from server
	latestCommitment, latestErr := s.server.GetClientCommitment()
	if s.setFailure(latestErr) {
		return // will rebound to init
	}
	latestCommitmentHash := latestCommitment.GetCommitmentHash()

	// check if commitment has already been attested
	log.Printf("********** received commitment hash: %s\n", latestCommitmentHash.String())
	if latestCommitmentHash == s.attestation.CommitmentHash() {
		log.Printf("********** Skipping attestation - Client commitment already attested")
		return // will remain at the same state
	}

	// publish new commitment hash to clients
	s.publisher.SendMessage((&latestCommitmentHash).CloneBytes(), confpkg.TOPIC_NEW_HASH)

	// initialise new attestation with commitment
	s.attestation = models.NewAttestationDefault()
	s.attestation.SetCommitment(&latestCommitment)

	s.state = ASTATE_NEW_ATTESTATION // update attestation state
}

// ASTATE_NEW_ATTESTATION
// - Generate new pay to address for attestation transaction using client commitment
// - Create new unsigned transaction using the last unspent
// - Publish unsigned transaction to signer clients
// - add ATIME_SIGS waiting time
func (s *AttestService) doStateNewAttestation() {
	log.Println("*AttestService* NEW ATTESTATION")

	// Get key and address for next attestation using client commitment
	key, keyErr := s.attester.GetNextAttestationKey(s.attestation.CommitmentHash())
	if s.setFailure(keyErr) {
		return // will rebound to init
	}
	paytoaddr, _ := s.attester.GetNextAttestationAddr(key, s.attestation.CommitmentHash())
	importErr := s.attester.ImportAttestationAddr(paytoaddr)
	if s.setFailure(importErr) {
		return // will rebound to init
	}
	log.Printf("********** pay-to addr: %s\n", paytoaddr.String())

	// Generate new unsigned attestation transaction from last unspent
	success, txunspent, unspentErr := s.attester.findLastUnspent()
	if s.setFailure(unspentErr) {
		return // will rebound to init
	} else if success {
		var createErr error
		var newTx *wire.MsgTx
		newTx, createErr = s.attester.createAttestation(paytoaddr, txunspent, false)
		s.attestation.Tx = *newTx
		if s.setFailure(createErr) {
			return // will rebound to init
		}

		log.Printf("********** pre-sign txid: %s\n", s.attestation.Tx.TxHash().String())

		// publish pre signed transaction
		var txbytes bytes.Buffer
		s.attestation.Tx.Serialize(&txbytes)
		s.publisher.SendMessage(txbytes.Bytes(), confpkg.TOPIC_NEW_TX)

		s.state = ASTATE_SIGN_ATTESTATION // update attestation state
		attestDelay = ATIME_SIGS          // add sigs waiting time
		log.Printf("********** sleeping for: %s ...\n", attestDelay.String())
	} else {
		s.setFailure(errors.New(ERROR_UNSPENT_NOT_FOUND))
		return // will rebound to init
	}
}

// ASTATE_SIGN_ATTESTATION
// - Collect signatures from client signers
// - Combine signatures them and sign the attestation transaction
func (s *AttestService) doStateSignAttestation() {
	log.Println("*AttestService* SIGN ATTESTATION")

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
	log.Printf("********** received %d signatures\n", len(sigs))

	// get last confirmed commitment from server
	lastCommitmentHash, latestErr := s.server.GetLatestAttestationCommitmentHash()
	if s.setFailure(latestErr) {
		return // will rebound to init
	}

	// sign attestation with combined sigs and last commitment
	signedTx, signErr := s.attester.signAttestation(&s.attestation.Tx, sigs, lastCommitmentHash)
	if s.setFailure(signErr) {
		return // will rebound to init
	}
	s.attestation.Tx = *signedTx
	s.attestation.Txid = s.attestation.Tx.TxHash()

	s.state = ASTATE_SEND_ATTESTATION // update attestation state
}

// ASTATE_SEND_ATTESTATION
// - Store unconfirmed attestation to server prior to sending
// - Send attestation transaction through the client to the network
// - add ATIME_CONFIRMATION waiting time
// - start time for confirmation time
func (s *AttestService) doStateSendAttestation() {
	log.Println("*AttestService* SEND ATTESTATION")

	// update server with latest unconfirmed attestation, in case the service fails
	errUpdate := s.server.UpdateLatestAttestation(*s.attestation)
	if s.setFailure(errUpdate) {
		return // will rebound to init
	}

	// sign attestation with combined signatures and send through client to network
	txid, attestationErr := s.attester.sendAttestation(&s.attestation.Tx)
	if s.setFailure(attestationErr) {
		return // will rebound to init
	}
	s.attestation.Txid = txid
	log.Printf("********** attestation transaction committed with txid: (%s)\n", txid)

	s.state = ASTATE_AWAIT_CONFIRMATION // update attestation state
	attestDelay = ATIME_CONFIRMATION    // add confirmation waiting time
	log.Printf("********** sleeping for: %s ...\n", attestDelay.String())
	confirmTime = time.Now() // set time for awaiting confirmation
}

// ASTATE_AWAIT_CONFIRMATION
// - Check if the attestation transaction has been confirmed in the main network
// - If confirmed, initiate new attestation, update server and signer clients
// - Check if ATIME_HANDLE_UNCONFIRMED has elapsed since attestation was sent
// - add ATIME_NEW_ATTESTATION if confirmed or ATIME_CONFIRMATION if not to waiting time
func (s *AttestService) doStateAwaitConfirmation() {
	log.Printf("*AttestService* AWAITING CONFIRMATION \ntxid: (%s)\ncommitment: (%s)\n", s.attestation.Txid.String(), s.attestation.CommitmentHash().String())

	// if attestation has been unconfirmed for too long
	// set to handle unconfirmed state
	if time.Since(confirmTime) > ATIME_HANDLE_UNCONFIRMED {
		s.state = ASTATE_HANDLE_UNCONFIRMED
		return
	}

	newTx, err := s.config.MainClient().GetTransaction(&s.attestation.Txid)
	if s.setFailure(err) {
		return // will rebound to init
	}

	if newTx.BlockHash != "" {
		log.Printf("********** attestation confirmed with txid: (%s)\n", s.attestation.Txid.String())

		// update server with latest confirmed attestation
		s.attestation.Confirmed = true
		errUpdate := s.server.UpdateLatestAttestation(*s.attestation)
		if s.setFailure(errUpdate) {
			return // will rebound to init
		}

		confirmedHash := s.attestation.CommitmentHash()
		s.publisher.SendMessage((&confirmedHash).CloneBytes(), confpkg.TOPIC_CONFIRMED_HASH) //update clients

		s.state = ASTATE_NEXT_COMMITMENT // update attestation state

		attestDelay = ATIME_NEW_ATTESTATION - time.Since(confirmTime) // add new attestation waiting time - subtract waiting time
	} else {
		attestDelay = ATIME_CONFIRMATION // add confirmation waiting time
	}
	log.Printf("********** sleeping for: %s ...\n", attestDelay.String())
}

// ASTATE_HANDLE_UNCONFIRMED
// - Handle attestations that have been unconfirmed for too long
func (s *AttestService) doStateHandleUnconfirmed() {
	log.Println("*AttestService* HANDLE UNCONFIRMED")

	return
}

//Main attestation service method - cycles through AttestationStates
func (s *AttestService) doAttestation() {

	// fixed waiting time between states specific states might
	// re-write this to set specific waiting times
	attestDelay = ATIME_FIXED

	switch s.state {

	case ASTATE_ERROR:
		s.doStateError()

	case ASTATE_INIT:
		s.doStateInit()

	case ASTATE_NEXT_COMMITMENT:
		s.doStateNextCommitment()

	case ASTATE_NEW_ATTESTATION:
		s.doStateNewAttestation()

	case ASTATE_SIGN_ATTESTATION:
		s.doStateSignAttestation()

	case ASTATE_SEND_ATTESTATION:
		s.doStateSendAttestation()

	case ASTATE_AWAIT_CONFIRMATION:
		s.doStateAwaitConfirmation()

	case ASTATE_HANDLE_UNCONFIRMED:
		s.doStateHandleUnconfirmed()
	}
}

// Check if there is an error and set error state
func (s *AttestService) setFailure(err error) bool {
	if err != nil {
		s.errorState = err
		s.state = ASTATE_ERROR
		return true
	}
	return false
}
