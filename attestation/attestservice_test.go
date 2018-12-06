// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	"errors"
	"fmt"
	"testing"
	"time"

	confpkg "mainstay/config"
	"mainstay/models"
	"mainstay/server"
	"mainstay/test"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"
)

// verify ASTATE_INIT
func verifyStateInit(t *testing.T, attestService *AttestService) {
	assert.Equal(t, &models.Attestation{Txid: chainhash.Hash{}, Tx: wire.MsgTx{}, Confirmed: false},
		attestService.attestation)
	assert.Equal(t, ASTATE_INIT, attestService.state)
}

// verify ASTATE_INIT to ASTATE_NEXT_COMMITMENT
func verifyStateInitToNextCommitment(t *testing.T, attestService *AttestService) {
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.CommitmentHash())
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)
	assert.Equal(t, models.AttestationInfo{}, attestService.attestation.Info)
	assert.Equal(t, ATIME_FIXED, attestDelay)
}

// verify ASTATE_INIT to ASTATE_AWAIT_CONFIRMATION
func verifyStateInitToAwaitConfirmation(t *testing.T, attestService *AttestService, latestCommitment *models.Commitment, txid chainhash.Hash) {
	attestService.doAttestation()
	assert.Equal(t, ASTATE_AWAIT_CONFIRMATION, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	assert.Equal(t, txid, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)
	assert.Equal(t, models.AttestationInfo{}, attestService.attestation.Info)
}

// verify ASTATE_NEXT_COMMITMENT to ASTATE_NEW_ATTESTATION
func verifyStateNextCommitmentToNewAttestation(t *testing.T, attestService *AttestService, dbFake *server.DbFake, hash *chainhash.Hash) *models.Commitment {
	latestCommitment, _ := models.NewCommitment([]chainhash.Hash{*hash})
	latestCommitments := []models.ClientCommitment{models.ClientCommitment{*hash, 0}}
	dbFake.SetClientCommitments(latestCommitments)
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEW_ATTESTATION, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	assert.Equal(t, ATIME_FIXED, attestDelay)

	return latestCommitment
}

// verify ASTATE_NEW_ATTESTATION to ASTATE_SIGN_ATTESTATION
func verifyStateNewAttestationToSignAttestation(t *testing.T, attestService *AttestService) {
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SIGN_ATTESTATION, attestService.state)
	// cant test much more here - we test this in other unit tests
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxIn))
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxOut))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[0].SignatureScript))
	assert.Equal(t, ATIME_SIGS, attestDelay)
}

// verify ASTATE_SIGN_ATTESTATION to ASTATE_PRE_SEND_STORE
func verifyStateSignAttestationToPreSendStore(t *testing.T, attestService *AttestService) {
	attestService.doAttestation()
	assert.Equal(t, ASTATE_PRE_SEND_STORE, attestService.state)
	assert.Equal(t, true, len(attestService.attestation.Tx.TxIn[0].SignatureScript) > 0)
	assert.Equal(t, ATIME_FIXED, attestDelay)
}

// verify ASTATE_PRE_SEND_STORE to ASTATE_SEND_ATTESTATION
func verifyStatePreSendStoreToSendAttestation(t *testing.T, attestService *AttestService) {
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SEND_ATTESTATION, attestService.state)
	assert.Equal(t, ATIME_FIXED, attestDelay)
}

// verify ASTATE_SEND_ATTESTATION to ASTATE_AWAIT_CONFIRMATION
func verifyStateSendAttestationToAwaitConfirmation(t *testing.T, attestService *AttestService) chainhash.Hash {
	attestService.doAttestation()
	assert.Equal(t, ASTATE_AWAIT_CONFIRMATION, attestService.state)
	assert.Equal(t, ATIME_CONFIRMATION, attestDelay)
	return attestService.attestation.Txid
}

// verify ASTATE_AWAIT_CONFIRMATION to ASTATE_AWAIT_CONFIRMATION
func verifyStateAwaitConfirmationToAwaitConfirmation(t *testing.T, attestService *AttestService) {
	attestService.doAttestation()
	assert.Equal(t, ASTATE_AWAIT_CONFIRMATION, attestService.state)
	assert.Equal(t, ATIME_CONFIRMATION, attestDelay)
}

// verify ASTATE_AWAIT_CONFIRMATION to ASTATE_NEXT_COMMITMENT
func verifyStateAwaitConfirmationToNextCommitment(t *testing.T, attestService *AttestService, config *confpkg.Config, txid chainhash.Hash, timeNew time.Duration) {
	// generate new block to confirm attestation
	rawTx, _ := config.MainClient().GetRawTransaction(&txid)
	walletTx, _ := config.MainClient().GetTransaction(&txid)

	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, true, attestService.attestation.Confirmed)
	assert.Equal(t, txid, attestService.attestation.Txid)
	assert.Equal(t, true, attestDelay < timeNew)
	assert.Equal(t, true, attestDelay > (timeNew-time.Since(confirmTime)))
	assert.Equal(t, models.AttestationInfo{
		Txid:      txid.String(),
		Blockhash: walletTx.BlockHash,
		Amount:    rawTx.MsgTx().TxOut[0].Value,
		Time:      walletTx.Time}, attestService.attestation.Info)
}

// verify ASTATE_AWAIT_CONFIRMATION to ASTATE_HANDLE_UNCONFIRMED
func verifyStateAwaitConfirmationToHandleUnconfirmed(t *testing.T, attestService *AttestService) {
	attestService.doAttestation()
	assert.Equal(t, ASTATE_HANDLE_UNCONFIRMED, attestService.state)
}

// verify ASTATE_HANDLE_UNCONFIRMED to ASTATE_SIGN_ATTESTATION
func verifyStateHandleUnconfirmedToSignAttestation(t *testing.T, attestService *AttestService) {
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SIGN_ATTESTATION, attestService.state)
	// cant test much more here - we test this in other unit tests
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxIn))
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxOut))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[0].SignatureScript))
	assert.Equal(t, ATIME_SIGS, attestDelay)
	assert.Equal(t, attestService.attester.Fees.minFee+attestService.attester.Fees.feeIncrement,
		attestService.attester.Fees.GetFee())
}

// Test Attest Service states
// Regular test cycle through states
// No failures except un updated server commitments
func TestAttestService_Regular(t *testing.T) {

	// Test INIT
	test := test.NewTest(false, false)
	config := test.Config

	// randomly test with invalid config here
	// timing config no effect on server
	timingConfig := confpkg.TimingConfig{-1, -1}
	config.SetTimingConfig(timingConfig)

	dbFake := server.NewDbFake()
	server := server.NewServer(dbFake)
	attestService := NewAttestService(nil, nil, server, NewAttestSignerFake(config), config)

	// Test initial state of attest service
	verifyStateInit(t, attestService)
	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	verifyStateInitToNextCommitment(t, attestService)

	// Test ASTATE_INIT -> ASTATE_ERROR
	// error case when server latest commitment not set
	// need to re-initiate attestation and set latest commitment in server
	attestService.doAttestation()
	assert.Equal(t, ASTATE_ERROR, attestService.state)
	assert.Equal(t, errors.New(models.ERROR_COMMITMENT_LIST_EMPTY), attestService.errorState)
	assert.Equal(t, ATIME_FIXED, attestDelay)

	// Test ASTATE_ERROR -> ASTATE_INIT -> ASTATE_NEXT_COMMITMENT again
	attestService.doAttestation()
	verifyStateInit(t, attestService)
	verifyStateInitToNextCommitment(t, attestService)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment := verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	verifyStateNewAttestationToSignAttestation(t, attestService)
	// Test ASTATE_SIGN_ATTESTATION -> ASTATE_PRE_SEND_STORE
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test ASTATE_PRE_SEND_STORE -> ASTATE_SEND_ATTESTATION
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test ASTATE_SEND_ATTESTATION -> ASTATE_AWAIT_CONFIRMATION
	txid := verifyStateSendAttestationToAwaitConfirmation(t, attestService)
	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_AWAIT_CONFIRMATION
	verifyStateAwaitConfirmationToAwaitConfirmation(t, attestService)
	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_AWAIT_CONFIRMATION
	verifyStateAwaitConfirmationToAwaitConfirmation(t, attestService)
	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_NEXT_COMMITMENT
	config.MainClient().Generate(1)
	verifyStateAwaitConfirmationToNextCommitment(t, attestService, config, txid, DEFAULT_ATIME_NEW_ATTESTATION)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	assert.Equal(t, DEFAULT_ATIME_NEW_ATTESTATION, attestDelay)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// stuck in next commitment
	// need to update server latest commitment
	hashY, _ := chainhash.NewHashFromStr("baaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment = verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashY)

	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	verifyStateNewAttestationToSignAttestation(t, attestService)
	// Test ASTATE_SIGN_ATTESTATION -> ASTATE_PRE_SEND_STORE
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test ASTATE_PRE_SEND_STORE -> ASTATE_SEND_ATTESTATION
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test ASTATE_SEND_ATTESTATION -> ASTATE_AWAIT_CONFIRMATION
	txid = verifyStateSendAttestationToAwaitConfirmation(t, attestService)
	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_NEXT_COMMITMENT
	config.MainClient().Generate(1)
	verifyStateAwaitConfirmationToNextCommitment(t, attestService, config, txid, DEFAULT_ATIME_NEW_ATTESTATION)
}

// Test Attest Service when Attestation remains unconfirmed
func TestAttestService_Unconfirmed(t *testing.T) {

	// Test INIT
	test := test.NewTest(false, false)
	config := test.Config

	// randomly test custom config here
	customAtimeNewAttestation := 5
	customAtimeHandleUnconfirmed := 10
	timingConfig := confpkg.TimingConfig{customAtimeNewAttestation, customAtimeHandleUnconfirmed}
	config.SetTimingConfig(timingConfig)

	dbFake := server.NewDbFake()
	server := server.NewServer(dbFake)
	attestService := NewAttestService(nil, nil, server, NewAttestSignerFake(config), config)

	attestService.attester.Fees.ResetFee(true)

	// Test initial state of attest service
	verifyStateInit(t, attestService)
	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	verifyStateInitToNextCommitment(t, attestService)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	_ = verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	verifyStateNewAttestationToSignAttestation(t, attestService)
	assert.Equal(t, attestService.attester.Fees.minFee, attestService.attester.Fees.GetFee())
	// Test ASTATE_SIGN_ATTESTATION -> ASTATE_PRE_SEND_STORE
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test ASTATE_PRE_SEND_STORE -> ASTATE_SEND_ATTESTATION
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test ASTATE_SEND_ATTESTATION -> ASTATE_AWAIT_CONFIRMATION
	txid := verifyStateSendAttestationToAwaitConfirmation(t, attestService)

	// set confirm time back to test what happens in handle unconfirmed case
	confirmTime = confirmTime.Add(-time.Duration(customAtimeHandleUnconfirmed) * time.Minute)

	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_HANDLE_UNCONFIRMED
	verifyStateAwaitConfirmationToHandleUnconfirmed(t, attestService)
	// Test ASTATE_HANDLE_UNCONFIRMED -> ASTATE_SIGN_ATTESTATION
	verifyStateHandleUnconfirmedToSignAttestation(t, attestService)
	assert.Equal(t, attestService.attester.Fees.minFee+attestService.attester.Fees.feeIncrement,
		attestService.attester.Fees.GetFee())

	// Test ASTATE_SIGN_ATTESTATION -> ASTATE_PRE_SEND_STORE
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test ASTATE_PRE_SEND_STORE -> ASTATE_SEND_ATTESTATION
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test ASTATE_SEND_ATTESTATION -> ASTATE_AWAIT_CONFIRMATION
	txid = verifyStateSendAttestationToAwaitConfirmation(t, attestService)
	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_NEXT_COMMITMENT
	config.MainClient().Generate(1)
	verifyStateAwaitConfirmationToNextCommitment(t, attestService, config, txid,
		time.Duration(customAtimeNewAttestation)*time.Minute)
	assert.Equal(t, attestService.attester.Fees.minFee, attestService.attester.Fees.GetFee())

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	hashY, _ := chainhash.NewHashFromStr("baaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	_ = verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashY)

	// add also unspent this time
	_ = createTopupUnspent(t, test.Config)
	attestService.attester.MainClient.Generate(1)

	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SIGN_ATTESTATION, attestService.state)
	// cant test much more here - we test this in other unit tests
	assert.Equal(t, 2, len(attestService.attestation.Tx.TxIn))
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxOut))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[0].SignatureScript))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[1].SignatureScript))
	assert.Equal(t, ATIME_SIGS, attestDelay)
	assert.Equal(t, attestService.attester.Fees.minFee, attestService.attester.Fees.GetFee())
	// Test ASTATE_SIGN_ATTESTATION -> ASTATE_PRE_SEND_STORE
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test ASTATE_PRE_SEND_STORE -> ASTATE_SEND_ATTESTATION
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test ASTATE_SEND_ATTESTATION -> ASTATE_AWAIT_CONFIRMATION
	txid = verifyStateSendAttestationToAwaitConfirmation(t, attestService)

	// set confirm time back to test what happens in handle unconfirmed case
	confirmTime = confirmTime.Add(-time.Duration(customAtimeHandleUnconfirmed) * time.Minute)

	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_HANDLE_UNCONFIRMED
	verifyStateAwaitConfirmationToHandleUnconfirmed(t, attestService)
	// Test ASTATE_HANDLE_UNCONFIRMED -> ASTATE_SIGN_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SIGN_ATTESTATION, attestService.state)
	// cant test much more here - we test this in other unit tests
	assert.Equal(t, 2, len(attestService.attestation.Tx.TxIn))
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxOut))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[0].SignatureScript))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[1].SignatureScript))
	assert.Equal(t, ATIME_SIGS, attestDelay)
	assert.Equal(t, attestService.attester.Fees.minFee+attestService.attester.Fees.feeIncrement,
		attestService.attester.Fees.GetFee())

	// Test ASTATE_SIGN_ATTESTATION -> ASTATE_PRE_SEND_STORE
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test ASTATE_PRE_SEND_STORE -> ASTATE_SEND_ATTESTATION
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test ASTATE_SEND_ATTESTATION -> ASTATE_AWAIT_CONFIRMATION
	txid = verifyStateSendAttestationToAwaitConfirmation(t, attestService)
	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_NEXT_COMMITMENT
	config.MainClient().Generate(1)
	verifyStateAwaitConfirmationToNextCommitment(t, attestService, config, txid,
		time.Duration(customAtimeNewAttestation)*time.Minute)
	assert.Equal(t, attestService.attester.Fees.minFee, attestService.attester.Fees.GetFee())
}

// Test Attest Service when dealing with topup Attestation
func TestAttestService_WithTopup(t *testing.T) {

	// Test INIT
	test := test.NewTest(false, false)
	config := test.Config

	// randomly test with invalid config here
	// timing config no effect on server
	timingConfig := confpkg.TimingConfig{-1, -1}
	config.SetTimingConfig(timingConfig)

	dbFake := server.NewDbFake()
	server := server.NewServer(dbFake)
	attestService := NewAttestService(nil, nil, server, NewAttestSignerFake(config), config)

	// Test initial state of attest service
	verifyStateInit(t, attestService)
	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	verifyStateInitToNextCommitment(t, attestService)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	_ = verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	verifyStateNewAttestationToSignAttestation(t, attestService)
	// Test ASTATE_SIGN_ATTESTATION -> ASTATE_PRE_SEND_STORE
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test ASTATE_PRE_SEND_STORE -> ASTATE_SEND_ATTESTATION
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test ASTATE_SEND_ATTESTATION -> ASTATE_AWAIT_CONFIRMATION
	txid := verifyStateSendAttestationToAwaitConfirmation(t, attestService)
	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_AWAIT_CONFIRMATION
	verifyStateAwaitConfirmationToAwaitConfirmation(t, attestService)
	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_AWAIT_CONFIRMATION
	verifyStateAwaitConfirmationToAwaitConfirmation(t, attestService)
	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_NEXT_COMMITMENT
	config.MainClient().Generate(1)
	verifyStateAwaitConfirmationToNextCommitment(t, attestService, config, txid, DEFAULT_ATIME_NEW_ATTESTATION)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// stuck in next commitment
	// need to update server latest commitment
	hashY, _ := chainhash.NewHashFromStr("baaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	_ = verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashY)

	// create top up unspent
	_ = createTopupUnspent(t, test.Config)
	attestService.attester.MainClient.Generate(1)

	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_SIGN_ATTESTATION, attestService.state)
	// cant test much more here - we test this in other unit tests
	assert.Equal(t, 2, len(attestService.attestation.Tx.TxIn))
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxOut))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[0].SignatureScript))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[1].SignatureScript))
	assert.Equal(t, ATIME_SIGS, attestDelay)

	// Test ASTATE_SIGN_ATTESTATION -> ASTATE_PRE_SEND_STORE
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test ASTATE_PRE_SEND_STORE -> ASTATE_SEND_ATTESTATION
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test ASTATE_SEND_ATTESTATION -> ASTATE_AWAIT_CONFIRMATION
	txid = verifyStateSendAttestationToAwaitConfirmation(t, attestService)
	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_NEXT_COMMITMENT
	config.MainClient().Generate(1)
	verifyStateAwaitConfirmationToNextCommitment(t, attestService, config, txid, DEFAULT_ATIME_NEW_ATTESTATION)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	hashZ, _ := chainhash.NewHashFromStr("caaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	_ = verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashZ)

	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	verifyStateNewAttestationToSignAttestation(t, attestService)
	// Test ASTATE_SIGN_ATTESTATION -> ASTATE_PRE_SEND_STORE
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test ASTATE_PRE_SEND_STORE -> ASTATE_SEND_ATTESTATION
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test ASTATE_SEND_ATTESTATION -> ASTATE_AWAIT_CONFIRMATION
	txid = verifyStateSendAttestationToAwaitConfirmation(t, attestService)
	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_AWAIT_CONFIRMATION
	verifyStateAwaitConfirmationToAwaitConfirmation(t, attestService)
	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_AWAIT_CONFIRMATION
	verifyStateAwaitConfirmationToAwaitConfirmation(t, attestService)
	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_NEXT_COMMITMENT
	config.MainClient().Generate(1)
	verifyStateAwaitConfirmationToNextCommitment(t, attestService, config, txid, DEFAULT_ATIME_NEW_ATTESTATION)
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
	attestService := NewAttestService(nil, nil, server, NewAttestSignerFake(config), config)

	// Test initial state of attest service
	verifyStateInit(t, attestService)
	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	verifyStateInitToNextCommitment(t, attestService)

	// failure - re init attestation service with restart
	attestService = NewAttestService(nil, nil, server, NewAttestSignerFake(config), config)
	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT again
	verifyStateInitToNextCommitment(t, attestService)

	// failure - re init attestation service from state failure
	attestService.state = ASTATE_INIT
	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT again
	verifyStateInitToNextCommitment(t, attestService)
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
	attestService := NewAttestService(nil, nil, server, NewAttestSignerFake(config), config)

	// Test initial state of attest service
	verifyStateInit(t, attestService)
	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	verifyStateInitToNextCommitment(t, attestService)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment := verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

	// failure - re init attestation service
	attestService = NewAttestService(nil, nil, server, NewAttestSignerFake(config), config)
	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	verifyStateInitToNextCommitment(t, attestService)
	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEW_ATTESTATION, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())

	// failure - re init attestation service from inner state failure
	attestService.state = ASTATE_INIT
	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	verifyStateInitToNextCommitment(t, attestService)
	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEW_ATTESTATION, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
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
	attestService := NewAttestService(nil, nil, server, NewAttestSignerFake(config), config)

	// Test initial state of attest service
	verifyStateInit(t, attestService)
	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	verifyStateInitToNextCommitment(t, attestService)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment := verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	verifyStateNewAttestationToSignAttestation(t, attestService)

	// failure - re init attestation service
	attestService = NewAttestService(nil, nil, server, NewAttestSignerFake(config), config)
	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	verifyStateInitToNextCommitment(t, attestService)
	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEW_ATTESTATION, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	verifyStateNewAttestationToSignAttestation(t, attestService)

	// failure - re init attestation service from inner state failure
	attestService.state = ASTATE_INIT
	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	verifyStateInitToNextCommitment(t, attestService)
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
	attestService := NewAttestService(nil, nil, server, NewAttestSignerFake(config), config)

	// Test initial state of attest service
	verifyStateInit(t, attestService)

	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	verifyStateInitToNextCommitment(t, attestService)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment := verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	verifyStateNewAttestationToSignAttestation(t, attestService)
	// Test ASTATE_SIGN_ATTESTATION -> ASTATE_PRE_SEND_STORE
	verifyStateSignAttestationToPreSendStore(t, attestService)

	// failure - re init attestation service
	attestService = NewAttestService(nil, nil, server, NewAttestSignerFake(config), config)

	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	verifyStateInitToNextCommitment(t, attestService)
	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEW_ATTESTATION, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	verifyStateNewAttestationToSignAttestation(t, attestService)
	// Test ASTATE_SIGN_ATTESTATION -> ASTATE_PRE_SEND_STORE
	verifyStateSignAttestationToPreSendStore(t, attestService)

	// failure - re init attestation service from inner state failure
	attestService.state = ASTATE_INIT
	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	verifyStateInitToNextCommitment(t, attestService)
}

// Test Attest Service states
// State cycle test with failures
// Test behaviour with fail after pre send store state
func TestAttestService_FailurePreSendStore(t *testing.T) {

	// Test INIT
	test := test.NewTest(false, false)
	config := test.Config

	dbFake := server.NewDbFake()
	server := server.NewServer(dbFake)
	attestService := NewAttestService(nil, nil, server, NewAttestSignerFake(config), config)

	// Test initial state of attest service
	verifyStateInit(t, attestService)

	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	verifyStateInitToNextCommitment(t, attestService)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment := verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	verifyStateNewAttestationToSignAttestation(t, attestService)
	// Test ASTATE_SIGN_ATTESTATION -> ASTATE_PRE_SEND_STORE
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test ASTATE_PRE_SEND_STORE -> ASTATE_SEND_ATTESTATION
	verifyStatePreSendStoreToSendAttestation(t, attestService)

	// failure - re init attestation service
	attestService = NewAttestService(nil, nil, server, NewAttestSignerFake(config), config)

	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	verifyStateInitToNextCommitment(t, attestService)
	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEW_ATTESTATION, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	verifyStateNewAttestationToSignAttestation(t, attestService)
	// Test ASTATE_SIGN_ATTESTATION -> ASTATE_PRE_SEND_STORE
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test ASTATE_PRE_SEND_STORE -> ASTATE_SEND_ATTESTATION
	verifyStatePreSendStoreToSendAttestation(t, attestService)

	// failure - re init attestation service from inner state failure
	attestService.state = ASTATE_INIT
	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	verifyStateInitToNextCommitment(t, attestService)
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

	prevAttestation := models.NewAttestationDefault()
	for i := range []int{1, 2, 3} {
		attestService := NewAttestService(nil, nil, server, NewAttestSignerFake(config), config)

		// Test initial state of attest service
		verifyStateInit(t, attestService)
		// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
		attestService.doAttestation()
		assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
		assert.Equal(t, prevAttestation.CommitmentHash(), attestService.attestation.CommitmentHash())
		assert.Equal(t, prevAttestation.Txid, attestService.attestation.Txid)
		assert.Equal(t, prevAttestation.Confirmed, attestService.attestation.Confirmed)
		assert.Equal(t, prevAttestation.Info, attestService.attestation.Info)
		assert.Equal(t, ATIME_FIXED, attestDelay)

		// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
		// set server commitment before creationg new attestation
		hashX, _ := chainhash.NewHashFromStr(fmt.Sprintf("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b%d", i))
		latestCommitment := verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

		// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
		verifyStateNewAttestationToSignAttestation(t, attestService)
		// Test ASTATE_SIGN_ATTESTATION -> ASTATE_PRE_SEND_STORE
		verifyStateSignAttestationToPreSendStore(t, attestService)
		// Test ASTATE_PRE_SEND_STORE -> ASTATE_SEND_ATTESTATION
		verifyStatePreSendStoreToSendAttestation(t, attestService)
		// Test ASTATE_SEND_ATTESTATION -> ASTATE_AWAIT_CONFIRMATION
		txid := verifyStateSendAttestationToAwaitConfirmation(t, attestService)

		// failure - re init attestation service
		attestService = NewAttestService(nil, nil, server, NewAttestSignerFake(config), config)

		// Test ASTATE_INIT -> ASTATE_AWAIT_CONFIRMATION
		verifyStateInitToAwaitConfirmation(t, attestService, latestCommitment, txid)

		// generate new block to confirm attestation
		config.MainClient().Generate(1)
		rawTx, _ := config.MainClient().GetRawTransaction(&txid)
		walletTx, _ := config.MainClient().GetTransaction(&txid)
		// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_NEXT_COMMITMENT
		attestService.doAttestation()
		assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
		assert.Equal(t, true, attestService.attestation.Confirmed)
		assert.Equal(t, txid, attestService.attestation.Txid)
		assert.Equal(t, models.AttestationInfo{
			Txid:      txid.String(),
			Blockhash: walletTx.BlockHash,
			Amount:    rawTx.MsgTx().TxOut[0].Value,
			Time:      walletTx.Time}, attestService.attestation.Info)

		// failure - re init attestation service from inner state failure
		attestService.state = ASTATE_INIT

		// Test ASTATE_INIT -> ASTATE_AWAIT_CONFIRMATION
		attestService.doAttestation()
		assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
		assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
		assert.Equal(t, txid, attestService.attestation.Txid)
		assert.Equal(t, true, attestService.attestation.Confirmed)
		assert.Equal(t, models.AttestationInfo{
			Txid:      txid.String(),
			Blockhash: walletTx.BlockHash,
			Amount:    rawTx.MsgTx().TxOut[0].Value,
			Time:      walletTx.Time}, attestService.attestation.Info)

		prevAttestation = attestService.attestation
	}
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
	attestService := NewAttestService(nil, nil, server, NewAttestSignerFake(config), config)

	// Test initial state of attest service
	verifyStateInit(t, attestService)
	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	verifyStateInitToNextCommitment(t, attestService)

	// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment := verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

	// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
	verifyStateNewAttestationToSignAttestation(t, attestService)
	// Test ASTATE_SIGN_ATTESTATION -> ASTATE_PRE_SEND_STORE
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test ASTATE_PRE_SEND_STORE -> ASTATE_SEND_ATTESTATION
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test ASTATE_SEND_ATTESTATION -> ASTATE_AWAIT_CONFIRMATION
	txid := verifyStateSendAttestationToAwaitConfirmation(t, attestService)

	// generate new block to confirm attestation
	config.MainClient().Generate(1)
	rawTx, _ := config.MainClient().GetRawTransaction(&txid)
	walletTx, _ := config.MainClient().GetTransaction(&txid)
	// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, true, attestService.attestation.Confirmed)
	assert.Equal(t, txid, attestService.attestation.Txid)
	assert.Equal(t, models.AttestationInfo{
		Txid:      txid.String(),
		Blockhash: walletTx.BlockHash,
		Amount:    rawTx.MsgTx().TxOut[0].Value,
		Time:      walletTx.Time}, attestService.attestation.Info)

	// failure - re init attestation service
	attestService = NewAttestService(nil, nil, server, NewAttestSignerFake(config), config)
	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	assert.Equal(t, txid, attestService.attestation.Txid)
	assert.Equal(t, true, attestService.attestation.Confirmed)
	assert.Equal(t, models.AttestationInfo{
		Txid:      txid.String(),
		Blockhash: walletTx.BlockHash,
		Amount:    rawTx.MsgTx().TxOut[0].Value,
		Time:      walletTx.Time}, attestService.attestation.Info)

	// failure - re init attestation service from inner state
	attestService.state = ASTATE_INIT
	// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
	attestService.doAttestation()
	assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	assert.Equal(t, txid, attestService.attestation.Txid)
	assert.Equal(t, true, attestService.attestation.Confirmed)
	assert.Equal(t, models.AttestationInfo{
		Txid:      txid.String(),
		Blockhash: walletTx.BlockHash,
		Amount:    rawTx.MsgTx().TxOut[0].Value,
		Time:      walletTx.Time}, attestService.attestation.Info)
}

// Test Attest Service states
// State cycle test with failures
// Test behaviour with fail after handle unconfirmed state
func TestAttestService_FailureHandleUnconfirmed(t *testing.T) {

	// Test INIT
	test := test.NewTest(false, false)
	config := test.Config

	dbFake := server.NewDbFake()
	server := server.NewServer(dbFake)

	prevAttestation := models.NewAttestationDefault()
	for i := range []int{1, 2, 3} {
		attestService := NewAttestService(nil, nil, server, NewAttestSignerFake(config), config)

		attestService.attester.Fees.ResetFee(true)

		// Test initial state of attest service
		verifyStateInit(t, attestService)
		// Test ASTATE_INIT -> ASTATE_NEXT_COMMITMENT
		attestService.doAttestation()
		assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
		assert.Equal(t, prevAttestation.CommitmentHash(), attestService.attestation.CommitmentHash())
		assert.Equal(t, prevAttestation.Txid, attestService.attestation.Txid)
		assert.Equal(t, prevAttestation.Confirmed, attestService.attestation.Confirmed)
		assert.Equal(t, prevAttestation.Info, attestService.attestation.Info)
		assert.Equal(t, ATIME_FIXED, attestDelay)

		// Test ASTATE_NEXT_COMMITMENT -> ASTATE_NEW_ATTESTATION
		// set server commitment before creationg new attestation
		hashX, _ := chainhash.NewHashFromStr(fmt.Sprintf("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b%d", i))
		latestCommitment := verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

		// Test ASTATE_NEW_ATTESTATION -> ASTATE_SIGN_ATTESTATION
		verifyStateNewAttestationToSignAttestation(t, attestService)
		assert.Equal(t, attestService.attester.Fees.minFee, attestService.attester.Fees.GetFee())
		// Test ASTATE_SIGN_ATTESTATION -> ASTATE_PRE_SEND_STORE
		verifyStateSignAttestationToPreSendStore(t, attestService)
		// Test ASTATE_PRE_SEND_STORE -> ASTATE_SEND_ATTESTATION
		verifyStatePreSendStoreToSendAttestation(t, attestService)
		// Test ASTATE_SEND_ATTESTATION -> ASTATE_AWAIT_CONFIRMATION
		txid := verifyStateSendAttestationToAwaitConfirmation(t, attestService)

		// set confirm time back to test what happens in handle unconfirmed case
		confirmTime = confirmTime.Add(-DEFAULT_ATIME_HANDLE_UNCONFIRMED)

		// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_HANDLE_UNCONFIRMED
		verifyStateAwaitConfirmationToHandleUnconfirmed(t, attestService)
		// Test ASTATE_HANDLE_UNCONFIRMED -> ASTATE_SIGN_ATTESTATION
		verifyStateHandleUnconfirmedToSignAttestation(t, attestService)
		assert.Equal(t, attestService.attester.Fees.minFee+attestService.attester.Fees.feeIncrement,
			attestService.attester.Fees.GetFee())

		// failure - re init attestation service with restart
		attestService = NewAttestService(nil, nil, server, NewAttestSignerFake(config), config)
		attestService.attester.Fees.ResetFee(true)
		// Test ASTATE_INIT -> ASTATE_AWAIT_CONFIRMATION
		verifyStateInitToAwaitConfirmation(t, attestService, latestCommitment, txid)

		// failure - re init attestation service from inner state failure
		attestService.state = ASTATE_INIT
		// Test ASTATE_INIT -> ASTATE_AWAIT_CONFIRMATION
		verifyStateInitToAwaitConfirmation(t, attestService, latestCommitment, txid)
		// set confirm time back to test what happens in handle unconfirmed case
		confirmTime = confirmTime.Add(-DEFAULT_ATIME_HANDLE_UNCONFIRMED)

		// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_HANDLE_UNCONFIRMED
		verifyStateAwaitConfirmationToHandleUnconfirmed(t, attestService)
		// Test ASTATE_HANDLE_UNCONFIRMED -> ASTATE_SIGN_ATTESTATION
		verifyStateHandleUnconfirmedToSignAttestation(t, attestService)
		assert.Equal(t, attestService.attester.Fees.minFee+attestService.attester.Fees.feeIncrement,
			attestService.attester.Fees.GetFee())

		// Test ASTATE_SIGN_ATTESTATION -> ASTATE_PRE_SEND_STORE
		verifyStateSignAttestationToPreSendStore(t, attestService)
		// Test ASTATE_PRE_SEND_STORE -> ASTATE_SEND_ATTESTATION
		verifyStatePreSendStoreToSendAttestation(t, attestService)

		// failure - re init attestation service with restart
		attestService = NewAttestService(nil, nil, server, NewAttestSignerFake(config), config)
		attestService.attester.Fees.ResetFee(true)
		// Test ASTATE_INIT -> ASTATE_AWAIT_CONFIRMATION
		verifyStateInitToAwaitConfirmation(t, attestService, latestCommitment, txid)

		// failure - re init attestation service from inner state failure
		attestService.state = ASTATE_INIT
		// Test ASTATE_INIT -> ASTATE_AWAIT_CONFIRMATION
		verifyStateInitToAwaitConfirmation(t, attestService, latestCommitment, txid)
		// set confirm time back to test what happens in handle unconfirmed case
		confirmTime = confirmTime.Add(-DEFAULT_ATIME_HANDLE_UNCONFIRMED)

		// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_HANDLE_UNCONFIRMED
		verifyStateAwaitConfirmationToHandleUnconfirmed(t, attestService)
		// Test ASTATE_HANDLE_UNCONFIRMED -> ASTATE_SIGN_ATTESTATION
		verifyStateHandleUnconfirmedToSignAttestation(t, attestService)
		assert.Equal(t, attestService.attester.Fees.minFee+attestService.attester.Fees.feeIncrement,
			attestService.attester.Fees.GetFee())

		// Test ASTATE_SIGN_ATTESTATION -> ASTATE_PRE_SEND_STORE
		verifyStateSignAttestationToPreSendStore(t, attestService)
		// Test ASTATE_PRE_SEND_STORE -> ASTATE_SEND_ATTESTATION
		verifyStatePreSendStoreToSendAttestation(t, attestService)
		// Test ASTATE_SEND_ATTESTATION -> ASTATE_AWAIT_CONFIRMATION
		txid = verifyStateSendAttestationToAwaitConfirmation(t, attestService)
		// Test ASTATE_INIT -> ASTATE_AWAIT_CONFIRMATION
		attestService.doAttestation()
		assert.Equal(t, ASTATE_AWAIT_CONFIRMATION, attestService.state)
		assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
		assert.Equal(t, txid, attestService.attestation.Txid)
		assert.Equal(t, false, attestService.attestation.Confirmed)
		assert.Equal(t, models.AttestationInfo{}, attestService.attestation.Info)

		// failure - re init attestation service from inner state failure
		attestService.state = ASTATE_INIT
		// Test ASTATE_INIT -> ASTATE_AWAIT_CONFIRMATION
		verifyStateInitToAwaitConfirmation(t, attestService, latestCommitment, txid)

		// generate new block to confirm attestation
		config.MainClient().Generate(1)
		rawTx, _ := config.MainClient().GetRawTransaction(&txid)
		walletTx, _ := config.MainClient().GetTransaction(&txid)
		// Test ASTATE_AWAIT_CONFIRMATION -> ASTATE_NEXT_COMMITMENT
		attestService.doAttestation()
		assert.Equal(t, ASTATE_NEXT_COMMITMENT, attestService.state)
		assert.Equal(t, true, attestService.attestation.Confirmed)
		assert.Equal(t, txid, attestService.attestation.Txid)
		assert.Equal(t, models.AttestationInfo{
			Txid:      txid.String(),
			Blockhash: walletTx.BlockHash,
			Amount:    rawTx.MsgTx().TxOut[0].Value,
			Time:      walletTx.Time}, attestService.attestation.Info)

		prevAttestation = attestService.attestation
	}
}
