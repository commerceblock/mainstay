// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	"bytes"
	"context"
	"errors"
	"log"
	"sync"
	"time"

	confpkg "mainstay/config"
	"mainstay/models"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	_ "github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

// Attestation Service is the main processes that handles generating
// attestations and maintaining communication to a bitcoin wallet

// Attestation state type
type AttestationState int

// Attestation states
const (
	AStateError             AttestationState = -1
	AStateInit              AttestationState = 0
	AStateNextCommitment    AttestationState = 1
	AStateNewAttestation    AttestationState = 2
	AStateSignAttestation   AttestationState = 3
	AStatePreSendStore      AttestationState = 4
	AStateSendAttestation   AttestationState = 5
	AStateAwaitConfirmation AttestationState = 6
	AStateHandleUnconfirmed AttestationState = 7
)

// error / warning consts
const (
	ErroUnspentNotFound = "No valid unspent found"

	WarningInvalidATimeNewAttestationArg    = "Warning - Invalid new attestation time config value"
	WarningInvalidATimeHandleUnconfirmedArg = "Warning - Invalid handle unconfirmed time config value"
)

// waiting time schedules
const (
	// fixed waiting time between states
	ATimeFixed = 5 * time.Second

	// waiting time for sigs to arrive from multisig nodes
	ATimeSigs = 1 * time.Minute

	// waiting time to next attestation attempt when skipping already attested commitment
	ATimeSkip = 1 * time.Minute

	// waiting time between attemps to check if an attestation has been confirmed
	ATimeConfirmation = 15 * time.Minute

	// waiting time between consecutive attestations after one was confirmed
	DefaultATimeNewAttestation = 60 * time.Minute

	// waiting time until we handle an attestation that has not been confirmed
	// usually by increasing the fee of the previous transcation to speed up confirmation
	DefaultATimeHandleUnconfirmed = 60 * time.Minute
)

// AttestationService structure
// Encapsulates Attest Client and connectivity
// to a AttestServer for updates and requests
type AttestService struct {
	// context required for safe service cancellation
	ctx context.Context

	// waitgroup required to maintain all goroutines
	wg *sync.WaitGroup

	// service config
	config *confpkg.Config

	// client interface for attestation creation and key tweaking
	attester *AttestClient

	// server connection for querying and/or storing information
	server *AttestServer

	// interface to signers to send commitments/transactions and receive signatures
	signer AttestSigner

	// mainstain current attestation state, model and error state
	state       AttestationState
	attestation *models.Attestation
	errorState  error
	isRegtest   bool
}

var (
	atimeNewAttestation    time.Duration // delay between attestations - DEFAULTS to DefaultATimeNewAttestation
	atimeHandleUnconfirmed time.Duration // delay until handling unconfirmed - DEFAULTS to DefaultATimeHandleUnconfirmed

	attestDelay time.Duration // handle state delay
	confirmTime time.Time     // handle confirmation timing

	isFeeBumped bool // flag to keep track if the fee has already been bumped
)

// NewAttestService returns a pointer to an AttestService instance
// Initiates Attest Client and Attest AttestServer
func NewAttestService(ctx context.Context, wg *sync.WaitGroup, server *AttestServer, signer AttestSigner, config *confpkg.Config) *AttestService {
	// Check init txid validity
	_, errInitTx := chainhash.NewHashFromStr(config.InitTx())
	if errInitTx != nil {
		log.Fatalf("Incorrect initial transaction id %s\n", config.InitTx())
	}

	// initiate attestation client
	attester := NewAttestClient(config)
	isFeeBumped = false

	// initiate timing schedules
	atimeNewAttestation = DefaultATimeNewAttestation
	if config.TimingConfig().NewAttestationMinutes > 0 {
		atimeNewAttestation = time.Duration(config.TimingConfig().NewAttestationMinutes) * time.Minute
	} else {
		log.Printf("%s (%v)\n", WarningInvalidATimeNewAttestationArg, config.TimingConfig().NewAttestationMinutes)
	}
	log.Printf("Time new attestation set to: %v\n", atimeNewAttestation)
	atimeHandleUnconfirmed = DefaultATimeHandleUnconfirmed
	if config.TimingConfig().HandleUnconfirmedMinutes > 0 {
		atimeHandleUnconfirmed = time.Duration(config.TimingConfig().HandleUnconfirmedMinutes) * time.Minute
	} else {
		log.Printf("%s (%v)\n", WarningInvalidATimeHandleUnconfirmedArg, config.TimingConfig().HandleUnconfirmedMinutes)
	}
	log.Printf("Time handle unconfirmed set to: %v\n", atimeHandleUnconfirmed)

	return &AttestService{ctx, wg, config, attester, server, signer, AStateInit, models.NewAttestationDefault(), nil, config.Regtest()}
}

// Run Attest Service
func (s *AttestService) Run() {
	defer s.wg.Done()

	attestDelay = 10 * time.Second // add some delay for subscribers to have time to set up

	for { //Doing attestations using attestation client and waiting for transaction confirmation
		timer := time.NewTimer(attestDelay)
		select {
		case <-s.ctx.Done():
			log.Println("Shutting down Attestation Service...")
			return
		case <-timer.C:
			// do next attestation state
			s.doAttestation()

			// for testing - overwrite delay
			if s.isRegtest {
				attestDelay = 10 * time.Second
			}

			log.Printf("********** sleeping for: %s ...\n", attestDelay.String())
		}
	}
}

// AStateError
// - Print error state and re-initiate attestation
func (s *AttestService) doStateError() {
	log.Println("*AttestService* ATTESTATION SERVICE FAILURE")
	log.Println(s.errorState)
	s.state = AStateInit // update attestation state
}

// part of AStateInit
// handle case when an unconfirmed transactions is found in the mempool
// fetch attestation information and set service state to AStateAwaitConfirmation
func (s *AttestService) stateInitUnconfirmed(unconfirmedTxid chainhash.Hash) {
	commitment, commitmentErr := s.server.GetAttestationCommitment(unconfirmedTxid, false)
	if s.setFailure(commitmentErr) {
		return // will rebound to init
	}
	log.Printf("********** found unconfirmed attestation: %s\n", unconfirmedTxid.String())
	s.attestation = models.NewAttestation(unconfirmedTxid, &commitment) // initialise attestation
	rawTx, _ := s.config.MainClient().GetRawTransaction(&unconfirmedTxid)
	s.attestation.Tx = *rawTx.MsgTx() // set msgTx

	s.state = AStateAwaitConfirmation // update attestation state
	walletTx, _ := s.config.MainClient().GetTransaction(&unconfirmedTxid)
	confirmTime = time.Unix(walletTx.Time, 0)
}

// part of AStateInit
// handle case when an unspent transaction is found in the wallet
// if the unspent is a previous attestation, update database info
// initiate a new attestation and inform signers of commitment
func (s *AttestService) stateInitUnspent(unspent btcjson.ListUnspentResult) {
	unspentTxid, _ := chainhash.NewHashFromStr(unspent.TxID)
	commitment, commitmentErr := s.server.GetAttestationCommitment(*unspentTxid)
	if s.setFailure(commitmentErr) {
		return // will rebound to init
	} else if (commitment.GetCommitmentHash() != chainhash.Hash{}) {
		log.Printf("********** found confirmed attestation: %s\n", unspentTxid.String())
		s.attestation = models.NewAttestation(*unspentTxid, &commitment)
		// update server with latest confirmed attestation
		s.attestation.Confirmed = true
		rawTx, _ := s.config.MainClient().GetRawTransaction(unspentTxid)
		walletTx, _ := s.config.MainClient().GetTransaction(unspentTxid)
		s.attestation.Tx = *rawTx.MsgTx()  // set msgTx
		s.attestation.UpdateInfo(walletTx) // set tx info

		errUpdate := s.server.UpdateLatestAttestation(*s.attestation)
		if s.setFailure(errUpdate) {
			return // will rebound to init
		}

		s.attester.Fees.ResetFee(s.isRegtest) // reset client fees
		// set delay to the difference between atimeNewAttestation and time since last attestation
		lastDelay := time.Since(time.Unix(s.attestation.Info.Time, 0))
		if atimeNewAttestation > lastDelay {
			attestDelay = atimeNewAttestation - lastDelay
		}
	} else {
		log.Println("********** found unspent transaction, initiating staychain")
		s.attestation = models.NewAttestationDefault()
	}
	confirmedHash := s.attestation.CommitmentHash()
	s.signer.SendConfirmedHash((&confirmedHash).CloneBytes()) // update clients

	s.state = AStateNextCommitment // update attestation state
}

// part of AStateInit
// handles wallet failure when neither unconfirmed or unspent is found
// above case should happen very rarely but when it does, import
// both latest unconfirmed and confirmed attestation addresses to wallet
func (s *AttestService) stateInitWalletFailure() {

	log.Println("********** wallet failure")

	// get last confirmed commitment from server
	lastCommitmentHash, latestErr := s.server.GetLatestAttestationCommitmentHash()
	if s.setFailure(latestErr) {
		return // will rebound to init
	}

	// Get latest confirmed attestation address and re-import to wallet
	paytoaddr, _, addrErr := s.attester.GetNextAttestationAddr((*btcutil.WIF)(nil), lastCommitmentHash)
	if s.setFailure(addrErr) {
		return // will rebound to init
	}
	log.Printf("********** importing latest confirmed addr: %s ...\n", paytoaddr.String())
	importErr := s.attester.ImportAttestationAddr(paytoaddr)
	if s.setFailure(importErr) {
		return // will rebound to init
	}

	// get last unconfirmed commitment from server
	lastCommitmentHash, latestErr = s.server.GetLatestAttestationCommitmentHash(false)
	if s.setFailure(latestErr) {
		return // will rebound to init
	}

	// Get latest unconfirmed attestation address and re-import to wallet
	paytoaddr, _, addrErr = s.attester.GetNextAttestationAddr((*btcutil.WIF)(nil), lastCommitmentHash)
	if s.setFailure(addrErr) {
		return // will rebound to init
	}
	log.Printf("********** importing latest unconfirmed addr: %s ...\n", paytoaddr.String())
	importErr = s.attester.ImportAttestationAddr(paytoaddr)
	if s.setFailure(importErr) {
		return // will rebound to init
	}

	s.state = AStateInit // update attestation state
}

// AStateInit
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
		// handle init unconfirmed case
		s.stateInitUnconfirmed(unconfirmedTxid)
	} else {
		success, unspent, unspentErr := s.attester.findLastUnspent()
		if s.setFailure(unspentErr) {
			return // will rebound to init
		} else if success {
			// handle init unspent case
			s.stateInitUnspent(unspent)
		} else {
			// handle wallet failure case
			s.stateInitWalletFailure()
		}
	}
}

// AStateNextCommitment
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
		attestDelay = ATimeSkip // sleep
		return                  // will remain at the same state
	}

	// initialise new attestation with commitment
	s.attestation = models.NewAttestationDefault()
	s.attestation.SetCommitment(&latestCommitment)

	s.state = AStateNewAttestation // update attestation state
}

// AStateNewAttestation
// - Generate new pay to address for attestation transaction using client commitment
// - Create new unsigned transaction using the last unspent
// - If a topup unspent exists, add this to the new attestation
// - Publish unsigned transaction to signer clients
// - add ATimeSigs waiting time
func (s *AttestService) doStateNewAttestation() {
	log.Println("*AttestService* NEW ATTESTATION")

	// Get key and address for next attestation using client commitment
	key, keyErr := s.attester.GetNextAttestationKey(s.attestation.CommitmentHash())
	if s.setFailure(keyErr) {
		return // will rebound to init
	}
	paytoaddr, _, addrErr := s.attester.GetNextAttestationAddr(key, s.attestation.CommitmentHash())
	if s.setFailure(addrErr) {
		return // will rebound to init
	}
	log.Printf("********** importing pay-to addr: %s ...\n", paytoaddr.String())
	importErr := s.attester.ImportAttestationAddr(paytoaddr, false) // no rescan needed here
	if s.setFailure(importErr) {
		return // will rebound to init
	}

	// Generate new unsigned attestation transaction from last unspent
	success, unspent, unspentErr := s.attester.findLastUnspent()
	if s.setFailure(unspentErr) {
		return // will rebound to init
	} else if success {
		var unspentList []btcjson.ListUnspentResult
		unspentList = append(unspentList, unspent)

		// search for topup unspent and add if it exists
		topupFound, topupUnspent, topupUnspentErr := s.attester.findTopupUnspent()
		if s.setFailure(topupUnspentErr) {
			return // will rebound to init
		} else if topupFound {
			log.Printf("********** found topup unspent: %s\n", topupUnspent.TxID)
			unspentList = append(unspentList, topupUnspent)
		}

		// create attestation transaction for the list of unspents paying to addr generated
		newTx, createErr := s.attester.createAttestation(paytoaddr, unspentList)
		if s.setFailure(createErr) {
			return // will rebound to init
		}

		s.attestation.Tx = *newTx
		log.Printf("********** pre-sign txid: %s\n", s.attestation.Tx.TxHash().String())

		// get last confirmed commitment from server
		lastCommitmentHash, latestErr := s.server.GetLatestAttestationCommitmentHash()
		if s.setFailure(latestErr) {
			return // will rebound to init
		}

		// publish pre signed transaction
		txPreImages, getPreImagesErr := s.attester.getTransactionPreImages(lastCommitmentHash, newTx)
		if s.setFailure(getPreImagesErr) {
			return // will rebound to init
		}
		// get pre image bytes
		var txPreImageBytes [][]byte
		for _, txPreImage := range txPreImages {
			var txBytesBuffer bytes.Buffer
			txPreImage.Serialize(&txBytesBuffer)
			txPreImageBytes = append(txPreImageBytes, txBytesBuffer.Bytes())
		}
		s.signer.SendTxPreImages(txPreImageBytes)

		s.state = AStateSignAttestation // update attestation state
		attestDelay = ATimeSigs         // add sigs waiting time
	} else {
		s.setFailure(errors.New(ErroUnspentNotFound))
		return // will rebound to init
	}
}

// AStateSignAttestation
// - Collect signatures from client signers
// - Combine signatures them and sign the attestation transaction
func (s *AttestService) doStateSignAttestation() {
	log.Println("*AttestService* SIGN ATTESTATION")

	// Read sigs using subscribers
	sigs := s.signer.GetSigs()
	for sigForInput, _ := range sigs {
		log.Printf("********** received %d signatures for input %d \n",
			len(sigs[sigForInput]), sigForInput)
	}

	// get last confirmed commitment from server
	lastCommitmentHash, latestErr := s.server.GetLatestAttestationCommitmentHash()
	if s.setFailure(latestErr) {
		return // will rebound to init
	}

	// sign attestation with combined sigs and last commitment
	signedTx, signErr := s.attester.signAttestation(&s.attestation.Tx, sigs, lastCommitmentHash)
	if s.setFailure(signErr) {
		log.Printf("********** signer failure. resubscribing to signers...")
		s.signer.ReSubscribe()
		return // will rebound to init
	}
	s.attestation.Tx = *signedTx
	s.attestation.Txid = s.attestation.Tx.TxHash()

	s.state = AStatePreSendStore // update attestation state
}

// AStatePreSendStore
// - Store unconfirmed attestation to server prior to sending
func (s *AttestService) doStatePreSendStore() {
	log.Println("*AttestService* PRE SEND STORE")

	// update server with latest unconfirmed attestation, in case the service fails
	errUpdate := s.server.UpdateLatestAttestation(*s.attestation)
	if s.setFailure(errUpdate) {
		return // will rebound to init
	}

	s.state = AStateSendAttestation // update attestation state
}

// AStateSendAttestation
// - Send attestation transaction through the client to the network
// - add ATimeConfirmation waiting time
// - start time for confirmation time
func (s *AttestService) doStateSendAttestation() {
	log.Println("*AttestService* SEND ATTESTATION")

	// sign attestation with combined signatures and send through client to network
	txid, attestationErr := s.attester.sendAttestation(&s.attestation.Tx)
	if s.setFailure(attestationErr) {
		return // will rebound to init
	}
	s.attestation.Txid = txid
	log.Printf("********** attestation transaction committed with txid: (%s)\n", txid)

	s.state = AStateAwaitConfirmation // update attestation state
	attestDelay = ATimeConfirmation   // add confirmation waiting time
	confirmTime = time.Now()          // set time for awaiting confirmation
	isFeeBumped = false               // reset fee bumped flag
}

// AStateAwaitConfirmation
// - Check if the attestation transaction has been confirmed in the main network
// - If confirmed, initiate new attestation, update server and signer clients
// - Check if ATIME_HANDLE_UNCONFIRMED has elapsed since attestation was sent
// - add ATIME_NEW_ATTESTATION if confirmed or ATimeConfirmation if not to waiting time
func (s *AttestService) doStateAwaitConfirmation() {
	log.Printf("*AttestService* AWAITING CONFIRMATION \ntxid: (%s)\ncommitment: (%s)\n", s.attestation.Txid.String(), s.attestation.CommitmentHash().String())

	// if attestation has been unconfirmed for too long
	// set to handle unconfirmed state
	if time.Since(confirmTime) > atimeHandleUnconfirmed {
		s.state = AStateHandleUnconfirmed
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
		s.attestation.UpdateInfo(newTx)
		errUpdate := s.server.UpdateLatestAttestation(*s.attestation)
		if s.setFailure(errUpdate) {
			return // will rebound to init
		}

		s.attester.Fees.ResetFee(s.isRegtest) // reset client fees

		confirmedHash := s.attestation.CommitmentHash()
		s.signer.SendConfirmedHash((&confirmedHash).CloneBytes()) // update clients

		s.state = AStateNextCommitment                              // update attestation state
		attestDelay = atimeNewAttestation - time.Since(confirmTime) // add new attestation waiting time - subtract waiting time
	} else {
		attestDelay = ATimeConfirmation // add confirmation waiting time
	}
}

// AStateHandleUnconfirmed
// - Handle attestations that have been unconfirmed for too long
// - Bump attestation fees and re-initiate sign and send process
func (s *AttestService) doStateHandleUnconfirmed() {
	log.Println("*AttestService* HANDLE UNCONFIRMED")

	log.Printf("********** bumping fees for attestation txid: %s\n", s.attestation.Tx.TxHash().String())
	currentTx := &s.attestation.Tx
	bumpErr := s.attester.bumpAttestationFees(currentTx, isFeeBumped)
	if s.setFailure(bumpErr) {
		return // will rebound to init
	}
	isFeeBumped = true

	s.attestation.Tx = *currentTx
	log.Printf("********** new pre-sign txid: %s\n", s.attestation.Tx.TxHash().String())

	// get last confirmed commitment from server
	lastCommitmentHash, latestErr := s.server.GetLatestAttestationCommitmentHash()
	if s.setFailure(latestErr) {
		return // will rebound to init
	}

	// re-publish pre signed transaction
	txPreImages, getPreImagesErr := s.attester.getTransactionPreImages(lastCommitmentHash, currentTx)
	if s.setFailure(getPreImagesErr) {
		return // will rebound to init
	}
	// get pre image bytes
	var txPreImageBytes [][]byte
	for _, txPreImage := range txPreImages {
		var txBytesBuffer bytes.Buffer
		txPreImage.Serialize(&txBytesBuffer)
		txPreImageBytes = append(txPreImageBytes, txBytesBuffer.Bytes())
	}
	s.signer.SendTxPreImages(txPreImageBytes)

	s.state = AStateSignAttestation // update attestation state
	attestDelay = ATimeSigs         // add sigs waiting time
}

//Main attestation service method - cycles through AttestationStates
func (s *AttestService) doAttestation() {

	// fixed waiting time between states specific states might
	// re-write this to set specific waiting times
	attestDelay = ATimeFixed

	switch s.state {

	case AStateError:
		s.doStateError()

	case AStateInit:
		s.doStateInit()

	case AStateNextCommitment:
		s.doStateNextCommitment()

	case AStateNewAttestation:
		s.doStateNewAttestation()

	case AStateSignAttestation:
		s.doStateSignAttestation()

	case AStatePreSendStore:
		s.doStatePreSendStore()

	case AStateSendAttestation:
		s.doStateSendAttestation()

	case AStateAwaitConfirmation:
		s.doStateAwaitConfirmation()

	case AStateHandleUnconfirmed:
		s.doStateHandleUnconfirmed()
	}
}

// Check if there is an error and set error state
func (s *AttestService) setFailure(err error) bool {
	if err != nil {
		s.errorState = err
		s.state = AStateError
		return true
	}
	return false
}
