/*
Package attestation implements the MainStay protocol.

Implemented using an Attestation Service structure that runs a main processes:
    - AttestClient that handles generating attestations and maintaining communication a bitcoin wallet
*/
package attestation

import (
	"bytes"
	"context"
	"errors"
	"log"
	confpkg "mainstay/config"
	"mainstay/crypto"
	"mainstay/messengers"
	"mainstay/models"
	"mainstay/server"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
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
// Encapsulates Attest Server, Attest Client
// and a channel for reading requests and writing responses
type AttestService struct {
	ctx           context.Context
	wg            *sync.WaitGroup
	config        *confpkg.Config
	attester      *AttestClient
	server        *server.Server
	publisher     *messengers.PublisherZmq
	subscribers   []*messengers.SubscriberZmq
	state         AttestationState
}

var latestAttestation *models.Attestation // hold latest state
var latestCommitment chainhash.Hash       // hold latest commitment
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

// Main attestation method. States:
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
			latestAttestation = models.NewAttestation(unconfirmedTxid, s.getTxAttestedHash(unconfirmedTxid))
			s.state = ASTATE_UNCONFIRMED
		} else {
			success, txunspent, unspentErr := s.attester.findLastUnspent()
			if s.failureState(unspentErr) {
				return
			} else if success {
				txunspentHash, _ := chainhash.NewHashFromStr(txunspent.TxID)
				latestCommitment = s.getTxAttestedHash(*txunspentHash)
				latestAttestation = models.NewAttestation(*txunspentHash, latestCommitment)
				s.state = ASTATE_CONFIRMED

				s.updateServer(*latestAttestation) // update server with latest attestation

				s.publisher.SendMessage(latestCommitment.CloneBytes(), confpkg.TOPIC_CONFIRMED_HASH) // update clients
			} else {
				log.Println("*AttestService* No unconfirmed (mempool) or unspent tx found. Sleeping...")
				attestDelay = time.Duration(ATTEST_WAIT_TIME*10) * time.Second // add wait time
			}
		}

	// ASTATE_UNCONFIRMED
	// - Check for unconfirmed transactions in the mempool of the main client
	// - If confirmed, initiate new attestation
	case ASTATE_UNCONFIRMED:
		log.Printf("*AttestService* AWAITING CONFIRMATION \ntxid: (%s)\nhash: (%s)\n", latestAttestation.Txid.String(), latestAttestation.AttestedHash.String())
		newTx, err := s.config.MainClient().GetTransaction(&latestAttestation.Txid)
		if s.failureState(err) {
			return
		}
		if newTx.BlockHash != "" {
			log.Printf("********** Attestation confirmed for txid: (%s)\n", latestAttestation.Txid.String())
			attestDelay = time.Duration(0) * time.Second

			s.state = ASTATE_CONFIRMED

			s.updateServer(*latestAttestation) // update server with latest attestation

			latestCommitment = latestAttestation.AttestedHash
			s.publisher.SendMessage(latestCommitment.CloneBytes(), confpkg.TOPIC_CONFIRMED_HASH) //update clients
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
			latestAttestation = models.NewAttestation(unconfirmedTxid, s.getTxAttestedHash(unconfirmedTxid))
			s.state = ASTATE_UNCONFIRMED

			log.Printf("*AttestService* Waiting for confirmation of\ntxid: (%s)\nhash: (%s)\n", latestAttestation.Txid.String(), latestAttestation.AttestedHash.String())
		} else {
			log.Println("*AttestService* NEW ATTESTATION")

			// get latest commitment hash from server
            latestHash, latestErr := s.server.GetLatestCommitment()
            if s.failureState(latestErr) {
                return
            }

			log.Printf("********** hash: %s\n", latestHash.String())
			if latestHash == latestAttestation.AttestedHash { // skip attestation if same client hash
				log.Printf("********** Skipping attestation - Client hash already attested")
				return
			}
			// publish new attestation hash
			s.publisher.SendMessage(latestHash.CloneBytes(), confpkg.TOPIC_NEW_HASH)
			latestAttestation = models.NewAttestation(chainhash.Hash{}, latestHash)
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

		key, keyErr := s.attester.GetNextAttestationKey(latestAttestation.AttestedHash)
		if s.failureState(keyErr) {
			return
		}

		paytoaddr, _ := s.attester.GetNextAttestationAddr(key, latestAttestation.AttestedHash)
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

		txid, attestationErr := s.attester.signAndSendAttestation(&latestAttestation.Tx, sigs, latest.AttestedHash)
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

// THIS WILL BE REPLACED BY QUERYING SERVER FOR COMMITMENT OF CORRESPONDING TX

// Find the attested sidechain hash from a transaction, by testing for all sidechain hashes
func (s *AttestService) getTxAttestedHash(txid chainhash.Hash) chainhash.Hash {
	oceanClient := s.config.OceanClient()

	// Get latest block and block height from sidechain
	latesthash, err := oceanClient.GetBestBlockHash()
	if err != nil {
		log.Fatal(err)
	}
	latestheight, err := oceanClient.GetBlockHeight(latesthash)
	if err != nil {
		log.Fatal(err)
	}

	// Get address from transaction
	tx, err := s.attester.MainClient.GetRawTransaction(&txid)
	if err != nil {
		log.Fatal(err)
	}
	_, addrs, _, errExtract := txscript.ExtractPkScriptAddrs(tx.MsgTx().TxOut[0].PkScript, s.attester.MainChainCfg)
	if errExtract != nil {
		log.Fatal(errExtract)
	}
	addr := addrs[0]

	tweakedPriv := crypto.TweakPrivKey(s.attester.WalletPriv, latesthash.CloneBytes(), s.attester.MainChainCfg)
	addrTweaked, _ := s.attester.GetNextAttestationAddr(tweakedPriv, *latesthash)
	// Check first if the attestation came from the latest block
	if addr.String() == addrTweaked.String() {
		return *latesthash
	}

	// Iterate backwards through all sidechain hashes to find the block hash that was attested
	for h := latestheight - 1; h >= 0; h-- {
		hash, err := oceanClient.GetBlockHash(int64(h))
		if err != nil {
			log.Fatal(err)
		}
		tweakedPriv := crypto.TweakPrivKey(s.attester.WalletPriv, hash.CloneBytes(), s.attester.MainChainCfg)
		addrTweaked, _ := s.attester.GetNextAttestationAddr(tweakedPriv, *hash)
		if addr.String() == addrTweaked.String() {
			return *hash
		}
	}

	return chainhash.Hash{}
}
