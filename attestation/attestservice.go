/*
Package attestation implements the MainStay protocol.

Implemented using an Attestation Service structure that runs multiple concurrent processes:
    - AttestClient that handles generating attestations and maintaining communication a bitcoin wallet
    - AttestServer that handles responding to API requests
    - A Requests channel to receive requests from requestapi
 */
package attestation

import (
    "log"
    "sync"
    "context"
    "time"

    "ocean-attestation/models"
    "ocean-attestation/config"

    "github.com/btcsuite/btcd/rpcclient"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
    "github.com/btcsuite/btcd/wire"
)

// Waiting time between attestations and/or attestation confirmation attempts
const ATTEST_WAIT_TIME = 60

// AttestationService structure
// Encapsulates Attest Server, Attest Client
// and a channel for reading requests and writing responses
type AttestService struct {
    ctx             context.Context
    wg              *sync.WaitGroup
    mainClient      *rpcclient.Client
    attester        *AttestClient
    server          *AttestServer
    channel         *models.Channel
}

var latestAttestation *Attestation
var latestTx *wire.MsgTx
var attestDelay time.Duration // initially 0 delay

// NewAttestService returns a pointer to an AttestService instance
// Initiates Attest Client and Attest Server
func NewAttestService(ctx context.Context, wg *sync.WaitGroup, channel *models.Channel, config *config.Config) *AttestService{
    if (len(config.InitTX()) != 64) {
        log.Fatal("Incorrect txid size")
    }
    attester := NewAttestClient(config.MainClient(), config.OceanClient(), config.MainChainCfg(), config.InitTX())

    genesisHash, err := config.OceanClient().GetBlockHash(0)
    if err!=nil {
        log.Fatal(err)
    }

    latestAttestation = &Attestation{chainhash.Hash{}, chainhash.Hash{}, ASTATE_NEW_ATTESTATION, time.Now()}
    server := NewAttestServer(config.OceanClient(), *latestAttestation, config.InitTX(), *genesisHash)

    return &AttestService{ctx, wg, config.MainClient(), attester, server, channel}
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
// ASTATE_UNCONFIRMED
// - Check for unconfirmed transactions in the mempool of the main client
// - If confirmed, initiate new attestation
// ASTATE_CONFIRMED, ASTATE_NEW_ATTESTATION
// - Either attestation failed or we are initiation the attestation chain
// - Generate new hash for attestation and publish it to other signers
// ASTATE_AWAITING_PUBKEYS
// - Waiting for tweaked keys from signers
// - Collect keys and generate address to pay the previous unspent to
// - Create unsigned attestation transaction and publish it to other signers
// - If successful await for sigs
// ASTATE_AWAITING_SIGS
// - Waiting for signatures from signers
// - Collect signatures, combine them and sign the attestation transaction
// - If successful propagate the transaction and set state to unconfirmed
func (s *AttestService) doAttestation() {
    switch latestAttestation.state {
    case ASTATE_UNCONFIRMED:
        log.Printf("*AttestService* Waiting for confirmation of\ntxid: (%s)\nhash: (%s)\n", latestAttestation.txid.String(), latestAttestation.attestedHash.String())
        newTx, err := s.mainClient.GetTransaction(&latestAttestation.txid)
        if err != nil {
            log.Fatal(err)
        }
        if (newTx.BlockHash != "") {
            latestAttestation.latestTime = time.Now()
            latestAttestation.state = ASTATE_CONFIRMED
            log.Printf("*AttestService* Attestation confirmed for txid: (%s)\n", latestAttestation.txid.String())
            s.server.UpdateLatest(*latestAttestation)
            log.Printf("**AttestServer** Updated latest attested height %d\n", s.server.latestHeight)
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
            hash := s.attester.getNextAttestationHash()
            if (hash == latestAttestation.attestedHash) { // skip attestation if same client hash
                log.Printf("*AttestService* Skipping attestation - Client hash already attested")
                return
            }
            //
            // publish
            //
            latestAttestation = &Attestation{chainhash.Hash{}, hash, ASTATE_AWAITING_PUBKEYS, time.Now()}
        }
    case ASTATE_AWAITING_PUBKEYS:
        log.Printf("*AttestService* AWAITING PUBKEYS\n")
        attestDelay = time.Duration(ATTEST_WAIT_TIME) * time.Second // add wait time

        key := s.attester.getNextAttestationKey(latestAttestation.attestedHash)
        //
        // read keys from subscriber
        // combine keys
        // if fail - restart
        //
        paytoaddr := s.attester.getNextAttestationAddr(key)

        success, txunspent := s.attester.findLastUnspent()
        if (success) {
            latestTx = s.attester.createAttestation(paytoaddr, txunspent, false)
            //
            // publish
            //
            latestAttestation.state = ASTATE_AWAITING_SIGS
        } else {
            log.Fatal("*AttestService* Attestation unsuccessful - No valid unspent found")
        }
    case ASTATE_AWAITING_SIGS:
        log.Printf("*AttestService* AWAITING SIGS\n")
        attestDelay = time.Duration(ATTEST_WAIT_TIME) * time.Second // add wait time

        //
        // read sigs from subscriber
        // combine sigs
        // if fail - restart
        //
        txid := s.attester.signAndSendAttestation(latestTx)
        log.Printf("*AttestService* Attestation committed for txid: (%s)\n", txid)
        latestAttestation.txid = txid
        latestAttestation.state = ASTATE_UNCONFIRMED
    }
}
