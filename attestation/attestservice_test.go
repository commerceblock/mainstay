package attestation

import (
	"errors"
	"testing"
	"time"

	"mainstay/models"
	"mainstay/server"
	"mainstay/test"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"
)

// Test Attest Service states
// Regular test cycle through states
// No failures except un updated server commitments
func TestAttestService_Regular(t *testing.T) {

	// Test INIT
	test := test.NewTest(false, false)
	config := test.Config

	dbFake := server.NewDbFake()
	server := server.NewServer(dbFake)
	attestService := NewAttestService(nil, nil, server, config)

	// Test initial state of attest service
	assert.Equal(t, &models.Attestation{Txid: chainhash.Hash{}, Tx: wire.MsgTx{}, Confirmed: false},
		attestService.attestation)
	assert.Equal(t, ASTATE_INIT, attestService.state)

	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.CommitmentHash())
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)
	assert.Equal(t, ATIME_FIXED, attestDelay)

	// Test ASTATE_INIT -> ASTATE_ERROR
	// error case when server latest commitment not set
	// need to re-initiate attestation and set latest commitment in server
	attestService.doAttestation()
	assert.Equal(t, ASTATE_ERROR, attestService.state)
	assert.Equal(t, errors.New(models.ERROR_COMMITMENT_LIST_EMPTY), attestService.errorState)
	assert.Equal(t, ATIME_FIXED, attestDelay)

	// Test ASTATE_ERROR -> ASTATE_INIT -> ASTATE_NEXT_COMMITMENT again
	attestService.doAttestation()
	assert.Equal(t, ASTATE_INIT, attestService.state)
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.CommitmentHash())
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)
	assert.Equal(t, ATIME_FIXED, attestDelay)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment, _ := models.NewCommitment([]chainhash.Hash{*hashX})
	latestCommitments := []models.ClientCommitment{models.ClientCommitment{*hashX, 0}}
	dbFake.SetClientCommitments(latestCommitments)
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEW_ATTESTATION, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	assert.Equal(t, ATIME_FIXED, attestDelay)

	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SIGN_ATTESTATION, attestService.state)
	// cant test much more here - we test this in other unit tests
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxIn))
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxOut))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[0].SignatureScript))
	assert.Equal(t, ATIME_SIGS, attestDelay)

	// Test ASTATE_SIGN_ATTESTATION -> ASTATE_SEND_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SEND_ATTESTATION, attestService.state)
	assert.Equal(t, true, len(attestService.attestation.Tx.TxIn[0].SignatureScript) > 0)
	assert.Equal(t, ATIME_FIXED, attestDelay)

	// Test ASTATE_SEND_ATTESTATION -> ASTATE_AWAIT_CONFIRMATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_AWAIT_CONFIRMATION, attestService.state)
	txid := attestService.attestation.Txid
	assert.Equal(t, ATIME_CONFIRMATION, attestDelay)

	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_AWAIT_CONFIRMATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_AWAIT_CONFIRMATION, attestService.state)
	assert.Equal(t, ATIME_CONFIRMATION, attestDelay)

	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_AWAIT_CONFIRMATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_AWAIT_CONFIRMATION, attestService.state)
	assert.Equal(t, ATIME_CONFIRMATION, attestDelay)

	// generate new block to confirm attestation
	config.MainClient().Generate(1)
	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, true, attestService.attestation.Confirmed)
	assert.Equal(t, txid, attestService.attestation.Txid)
	assert.Equal(t, true, attestDelay < ATIME_NEW_ATTESTATION)
	assert.Equal(t, true, attestDelay > (ATIME_NEW_ATTESTATION-time.Since(confirmTime)))

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	assert.Equal(t, ATIME_FIXED, attestDelay)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// stuck in next commitment
	// need to update server latest commitment
	hashY, _ := chainhash.NewHashFromStr("baaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment, _ = models.NewCommitment([]chainhash.Hash{*hashY})
	latestCommitments = []models.ClientCommitment{models.ClientCommitment{*hashY, 0}}
	dbFake.SetClientCommitments(latestCommitments)
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEW_ATTESTATION, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	assert.Equal(t, ATIME_FIXED, attestDelay)

	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SIGN_ATTESTATION, attestService.state)
	// cant test much more here - we test this in other unit tests
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxIn))
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxOut))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[0].SignatureScript))
	assert.Equal(t, ATIME_SIGS, attestDelay)

	// Test ASTATE_SIGN_ATTESTATION -> ASTATE_SEND_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SEND_ATTESTATION, attestService.state)
	assert.Equal(t, true, len(attestService.attestation.Tx.TxIn[0].SignatureScript) > 0)
	assert.Equal(t, ATIME_FIXED, attestDelay)

	// Test ASTATE_SEND_ATTESTATION -> ASTATE_AWAIT_CONFIRMATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_AWAIT_CONFIRMATION, attestService.state)
	txid = attestService.attestation.Txid
	assert.Equal(t, ATIME_CONFIRMATION, attestDelay)

	// generate new block to confirm attestation
	config.MainClient().Generate(1)
	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, true, attestService.attestation.Confirmed)
	assert.Equal(t, txid, attestService.attestation.Txid)
	assert.Equal(t, true, attestDelay < ATIME_NEW_ATTESTATION)
	assert.Equal(t, true, attestDelay > (ATIME_NEW_ATTESTATION-time.Since(confirmTime)))
}

// Test Attest Service when Attestation remains unconfirmed
func TestAttestService_HandleUnconfirmed(t *testing.T) {

	// Test INIT
	test := test.NewTest(false, false)
	config := test.Config

	dbFake := server.NewDbFake()
	server := server.NewServer(dbFake)
	attestService := NewAttestService(nil, nil, server, config)

	// Test initial state of attest service
	assert.Equal(t, &models.Attestation{Txid: chainhash.Hash{}, Tx: wire.MsgTx{}, Confirmed: false},
		attestService.attestation)
	assert.Equal(t, ASTATE_INIT, attestService.state)

	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.CommitmentHash())
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)
	assert.Equal(t, ATIME_FIXED, attestDelay)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment, _ := models.NewCommitment([]chainhash.Hash{*hashX})
	latestCommitments := []models.ClientCommitment{models.ClientCommitment{*hashX, 0}}
	dbFake.SetClientCommitments(latestCommitments)
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEW_ATTESTATION, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	assert.Equal(t, ATIME_FIXED, attestDelay)

	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SIGN_ATTESTATION, attestService.state)
	// cant test much more here - we test this in other unit tests
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxIn))
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxOut))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[0].SignatureScript))
	assert.Equal(t, ATIME_SIGS, attestDelay)

	// Test ASTATE_SIGN_ATTESTATION -> ASTATE_SEND_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SEND_ATTESTATION, attestService.state)
	assert.Equal(t, true, len(attestService.attestation.Tx.TxIn[0].SignatureScript) > 0)
	assert.Equal(t, ATIME_FIXED, attestDelay)

	// Test ASTATE_SEND_ATTESTATION -> ASTATE_AWAIT_CONFIRMATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_AWAIT_CONFIRMATION, attestService.state)
	_ = attestService.attestation.Txid
	assert.Equal(t, ATIME_CONFIRMATION, attestDelay)

	// set confirm time back to test what happens in handle unconfirmed case
	confirmTime = confirmTime.Add(-ATIME_HANDLE_UNCONFIRMED)

	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_AWAIT_CONFIRMATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_HANDLE_UNCONFIRMED, attestService.state)

	// TODO....
	// ...
	// Also TODO: failure case after ASTATE_HANDLE_UNCONFIRMED
}

// Test Attest Service states
// State cycle test with failures
// Test behaviour with fail after init state
func TestAttestService_FailureInit(t *testing.T) {

	// Test INIT
	test := test.NewTest(false, false)
	config := test.Config

	dbFake := server.NewDbFake()
	server := server.NewServer(dbFake)
	attestService := NewAttestService(nil, nil, server, config)

	// Test initial state of attest service
	assert.Equal(t, &models.Attestation{Txid: chainhash.Hash{}, Tx: wire.MsgTx{}, Confirmed: false},
		attestService.attestation)
	assert.Equal(t, ASTATE_INIT, attestService.state)

	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.CommitmentHash())
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)

	// failure - re init attestation service with restart
	attestService = NewAttestService(nil, nil, server, config)
	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT again
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.CommitmentHash())
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)

	// failure - re init attestation service from state failure
	attestService.state = ASTATE_INIT
	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT again
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.CommitmentHash())
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)
}

// Test Attest Service states
// State cycle test with failures
// Test behaviour with fail after next commitment state
func TestAttestService_FailureNextCommitment(t *testing.T) {

	// Test INIT
	test := test.NewTest(false, false)
	config := test.Config

	dbFake := server.NewDbFake()
	server := server.NewServer(dbFake)
	attestService := NewAttestService(nil, nil, server, config)

	// Test initial state of attest service
	assert.Equal(t, &models.Attestation{Txid: chainhash.Hash{}, Tx: wire.MsgTx{}, Confirmed: false},
		attestService.attestation)
	assert.Equal(t, ASTATE_INIT, attestService.state)

	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.CommitmentHash())
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment, _ := models.NewCommitment([]chainhash.Hash{*hashX})
	latestCommitments := []models.ClientCommitment{models.ClientCommitment{*hashX, 0}}
	dbFake.SetClientCommitments(latestCommitments)
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEW_ATTESTATION, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())

	// failure - re init attestation service
	attestService = NewAttestService(nil, nil, server, config)
	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.CommitmentHash())
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEW_ATTESTATION, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())

	// failure - re init attestation service from inner state failure
	attestService.state = ASTATE_INIT
	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.CommitmentHash())
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)
}

// Test Attest Service states
// State cycle test with failures
// Test behaviour with fail after new attestation state
func TestAttestService_FailureNewAttestation(t *testing.T) {

	// Test INIT
	test := test.NewTest(false, false)
	config := test.Config

	dbFake := server.NewDbFake()
	server := server.NewServer(dbFake)
	attestService := NewAttestService(nil, nil, server, config)

	// Test initial state of attest service
	assert.Equal(t, &models.Attestation{Txid: chainhash.Hash{}, Tx: wire.MsgTx{}, Confirmed: false},
		attestService.attestation)
	assert.Equal(t, ASTATE_INIT, attestService.state)

	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.CommitmentHash())
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment, _ := models.NewCommitment([]chainhash.Hash{*hashX})
	latestCommitments := []models.ClientCommitment{models.ClientCommitment{*hashX, 0}}
	dbFake.SetClientCommitments(latestCommitments)
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEW_ATTESTATION, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())

	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SIGN_ATTESTATION, attestService.state)
	// cant test much more here - we test this in other unit tests
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxIn))
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxOut))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[0].SignatureScript))

	// failure - re init attestation service
	attestService = NewAttestService(nil, nil, server, config)

	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.CommitmentHash())
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEW_ATTESTATION, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())

	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SIGN_ATTESTATION, attestService.state)
	// cant test much more here - we test this in other unit tests
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxIn))
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxOut))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[0].SignatureScript))

	// failure - re init attestation service from inner state failure
	attestService.state = ASTATE_INIT

	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.CommitmentHash())
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)
}

// Test Attest Service states
// State cycle test with failures
// Test behaviour with fail after sign attestation state
func TestAttestService_FailureSignAttestation(t *testing.T) {

	// Test INIT
	test := test.NewTest(false, false)
	config := test.Config

	dbFake := server.NewDbFake()
	server := server.NewServer(dbFake)
	attestService := NewAttestService(nil, nil, server, config)

	// Test initial state of attest service
	assert.Equal(t, &models.Attestation{Txid: chainhash.Hash{}, Tx: wire.MsgTx{}, Confirmed: false},
		attestService.attestation)
	assert.Equal(t, ASTATE_INIT, attestService.state)

	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.CommitmentHash())
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment, _ := models.NewCommitment([]chainhash.Hash{*hashX})
	latestCommitments := []models.ClientCommitment{models.ClientCommitment{*hashX, 0}}
	dbFake.SetClientCommitments(latestCommitments)
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEW_ATTESTATION, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())

	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SIGN_ATTESTATION, attestService.state)
	// cant test much more here - we test this in other unit tests
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxIn))
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxOut))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[0].SignatureScript))

	// Test ASTATE_SIGN_ATTESTATION -> ASTATE_SEND_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SEND_ATTESTATION, attestService.state)
	assert.Equal(t, true, len(attestService.attestation.Tx.TxIn[0].SignatureScript) > 0)

	// failure - re init attestation service
	attestService = NewAttestService(nil, nil, server, config)

	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.CommitmentHash())
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEW_ATTESTATION, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())

	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SIGN_ATTESTATION, attestService.state)
	// cant test much more here - we test this in other unit tests
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxIn))
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxOut))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[0].SignatureScript))

	// Test ASTATE_SIGN_ATTESTATION -> ASTATE_SEND_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SEND_ATTESTATION, attestService.state)
	assert.Equal(t, true, len(attestService.attestation.Tx.TxIn[0].SignatureScript) > 0)

	// failure - re init attestation service from inner state failure
	attestService.state = ASTATE_INIT

	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.CommitmentHash())
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)
}

// Test Attest Service states
// State cycle test with failures
// Test behaviour with fail after send attestation state
func TestAttestService_FailureSendAttestation(t *testing.T) {

	// Test INIT
	test := test.NewTest(false, false)
	config := test.Config

	dbFake := server.NewDbFake()
	server := server.NewServer(dbFake)
	attestService := NewAttestService(nil, nil, server, config)

	// Test initial state of attest service
	assert.Equal(t, &models.Attestation{Txid: chainhash.Hash{}, Tx: wire.MsgTx{}, Confirmed: false},
		attestService.attestation)
	assert.Equal(t, ASTATE_INIT, attestService.state)

	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.CommitmentHash())
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment, _ := models.NewCommitment([]chainhash.Hash{*hashX})
	latestCommitments := []models.ClientCommitment{models.ClientCommitment{*hashX, 0}}
	dbFake.SetClientCommitments(latestCommitments)
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEW_ATTESTATION, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())

	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SIGN_ATTESTATION, attestService.state)
	// cant test much more here - we test this in other unit tests
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxIn))
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxOut))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[0].SignatureScript))

	// Test ASTATE_SIGN_ATTESTATION -> ASTATE_SEND_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SEND_ATTESTATION, attestService.state)
	assert.Equal(t, true, len(attestService.attestation.Tx.TxIn[0].SignatureScript) > 0)

	// Test ASTATE_SEND_ATTESTATION -> ASTATE_AWAIT_CONFIRMATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_AWAIT_CONFIRMATION, attestService.state)
	txid := attestService.attestation.Txid

	// failure - re init attestation service
	attestService = NewAttestService(nil, nil, server, config)

	// Test ASTATE_INIT -> ASTATE_AWAIT_CONFIRMATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_AWAIT_CONFIRMATION, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	assert.Equal(t, txid, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)

	// generate new block to confirm attestation
	config.MainClient().Generate(1)
	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, true, attestService.attestation.Confirmed)
	assert.Equal(t, txid, attestService.attestation.Txid)

	// failure - re init attestation service from inner state failure
	attestService.state = ASTATE_INIT

	// Test ASTATE_INIT -> ASTATE_AWAIT_CONFIRMATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	assert.Equal(t, txid, attestService.attestation.Txid)
	assert.Equal(t, true, attestService.attestation.Confirmed)
}

// Test Attest Service states
// State cycle test with failures
// Test behaviour with fail after await confirmation state
func TestAttestService_FailureAwaitConfirmation(t *testing.T) {

	// Test INIT
	test := test.NewTest(false, false)
	config := test.Config

	dbFake := server.NewDbFake()
	server := server.NewServer(dbFake)
	attestService := NewAttestService(nil, nil, server, config)

	// Test initial state of attest service
	assert.Equal(t, &models.Attestation{Txid: chainhash.Hash{}, Tx: wire.MsgTx{}, Confirmed: false},
		attestService.attestation)
	assert.Equal(t, ASTATE_INIT, attestService.state)

	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.CommitmentHash())
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment, _ := models.NewCommitment([]chainhash.Hash{*hashX})
	latestCommitments := []models.ClientCommitment{models.ClientCommitment{*hashX, 0}}
	dbFake.SetClientCommitments(latestCommitments)
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEW_ATTESTATION, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())

	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SIGN_ATTESTATION, attestService.state)
	// cant test much more here - we test this in other unit tests
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxIn))
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxOut))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[0].SignatureScript))

	// Test ASTATE_SIGN_ATTESTATION -> ASTATE_SEND_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SEND_ATTESTATION, attestService.state)
	assert.Equal(t, true, len(attestService.attestation.Tx.TxIn[0].SignatureScript) > 0)

	// Test ASTATE_SEND_ATTESTATION -> ASTATE_AWAIT_CONFIRMATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_AWAIT_CONFIRMATION, attestService.state)
	txid := attestService.attestation.Txid

	// generate new block to confirm attestation
	config.MainClient().Generate(1)
	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, true, attestService.attestation.Confirmed)
	assert.Equal(t, txid, attestService.attestation.Txid)

	// failure - re init attestation service
	attestService = NewAttestService(nil, nil, server, config)

	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	assert.Equal(t, txid, attestService.attestation.Txid)
	assert.Equal(t, true, attestService.attestation.Confirmed)

	// failure again and check nothing has changed
	attestService = NewAttestService(nil, nil, server, config)

	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	assert.Equal(t, txid, attestService.attestation.Txid)
	assert.Equal(t, true, attestService.attestation.Confirmed)

	// failure - re init attestation service from inner state
	attestService.state = ASTATE_INIT

	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	assert.Equal(t, txid, attestService.attestation.Txid)
	assert.Equal(t, true, attestService.attestation.Confirmed)
}
