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

// verify AStateInit
func verifyStateInit(t *testing.T, attestService *AttestService) {
	assert.Equal(t, &models.Attestation{Txid: chainhash.Hash{}, Tx: wire.MsgTx{}, Confirmed: false},
		attestService.attestation)
	assert.Equal(t, AStateInit, attestService.state)
}

// verify AStateInit wallet failure
// feature is hard to test without crashing the bitcoin
// wallet, therefore just testing it doesn't break anything
func verifyStateInitWalletFailure(t *testing.T, attestService *AttestService) {
	// hard to test - just run here for now
	attestService.stateInitWalletFailure()
	assert.Equal(t, AStateInit, attestService.state)
}

// verify AStateInit to AStateNextCommitment
func verifyStateInitToNextCommitment(t *testing.T, attestService *AttestService) {
	attestService.doAttestation()
	assert.Equal(t, AStateNextCommitment, attestService.state)
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.CommitmentHash())
	assert.Equal(t, chainhash.Hash{}, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)
	assert.Equal(t, models.AttestationInfo{}, attestService.attestation.Info)
	assert.Equal(t, ATimeFixed, attestDelay)
}

// verify AStateInit to AStateAwaitConfirmation
func verifyStateInitToAwaitConfirmation(t *testing.T, attestService *AttestService, latestCommitment *models.Commitment, txid chainhash.Hash) {
	attestService.doAttestation()
	assert.Equal(t, AStateAwaitConfirmation, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	assert.Equal(t, txid, attestService.attestation.Txid)
	assert.Equal(t, false, attestService.attestation.Confirmed)
	assert.Equal(t, models.AttestationInfo{}, attestService.attestation.Info)
}

// verify AStateNextCommitment to AStateNewAttestation
func verifyStateNextCommitmentToNewAttestation(t *testing.T, attestService *AttestService, dbFake *server.DbFake, hash *chainhash.Hash) *models.Commitment {
	latestCommitment, _ := models.NewCommitment([]chainhash.Hash{*hash})
	latestCommitments := []models.ClientCommitment{models.ClientCommitment{*hash, 0}}
	dbFake.SetClientCommitments(latestCommitments)
	attestService.doAttestation()
	assert.Equal(t, AStateNewAttestation, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	assert.Equal(t, ATimeFixed, attestDelay)

	return latestCommitment
}

// verify AStateNewAttestation to AStateSignAttestation
func verifyStateNewAttestationToSignAttestation(t *testing.T, attestService *AttestService) {
	attestService.doAttestation()
	assert.Equal(t, AStateSignAttestation, attestService.state)
	// cant test much more here - we test this in other unit tests
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxIn))
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxOut))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[0].SignatureScript))
	assert.Equal(t, ATimeSigs, attestDelay)
}

// verify AStateSignAttestation to AStatePreSendStore
func verifyStateSignAttestationToPreSendStore(t *testing.T, attestService *AttestService) {
	attestService.doAttestation()
	assert.Equal(t, AStatePreSendStore, attestService.state)
	assert.Equal(t, true, len(attestService.attestation.Tx.TxIn[0].SignatureScript) > 0)
	assert.Equal(t, ATimeFixed, attestDelay)
}

// verify AStatePreSendStore to AStateSendAttestation
func verifyStatePreSendStoreToSendAttestation(t *testing.T, attestService *AttestService) {
	attestService.doAttestation()
	assert.Equal(t, AStateSendAttestation, attestService.state)
	assert.Equal(t, ATimeFixed, attestDelay)
}

// verify AStateSendAttestation to AStateAwaitConfirmation
func verifyStateSendAttestationToAwaitConfirmation(t *testing.T, attestService *AttestService) chainhash.Hash {
	attestService.doAttestation()
	assert.Equal(t, AStateAwaitConfirmation, attestService.state)
	assert.Equal(t, ATimeConfirmation, attestDelay)
	return attestService.attestation.Txid
}

// verify AStateAwaitConfirmation to AStateAwaitConfirmation
func verifyStateAwaitConfirmationToAwaitConfirmation(t *testing.T, attestService *AttestService) {
	attestService.doAttestation()
	assert.Equal(t, AStateAwaitConfirmation, attestService.state)
	assert.Equal(t, ATimeConfirmation, attestDelay)
}

// verify AStateAwaitConfirmation to AStateNextCommitment
func verifyStateAwaitConfirmationToNextCommitment(t *testing.T, attestService *AttestService, config *confpkg.Config, txid chainhash.Hash, timeNew time.Duration) {
	// generate new block to confirm attestation
	rawTx, _ := config.MainClient().GetRawTransaction(&txid)
	walletTx, _ := config.MainClient().GetTransaction(&txid)

	attestService.doAttestation()
	assert.Equal(t, AStateNextCommitment, attestService.state)
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

// verify AStateAwaitConfirmation to AStateHandleUnconfirmed
func verifyStateAwaitConfirmationToHandleUnconfirmed(t *testing.T, attestService *AttestService) {
	attestService.doAttestation()
	assert.Equal(t, AStateHandleUnconfirmed, attestService.state)
}

// verify AStateHandleUnconfirmed to AStateSignAttestation
func verifyStateHandleUnconfirmedToSignAttestation(t *testing.T, attestService *AttestService) {
	attestService.doAttestation()
	assert.Equal(t, AStateSignAttestation, attestService.state)
	// cant test much more here - we test this in other unit tests
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxIn))
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxOut))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[0].SignatureScript))
	assert.Equal(t, ATimeSigs, attestDelay)
	assert.Equal(t, attestService.attester.Fees.minFee+attestService.attester.Fees.feeIncrement,
		attestService.attester.Fees.GetFee())
}

// Test Attest Service states
// Regular test cycle through states
// Complete test for multiple signatures
// Any crucial functionality added should go through this test as it uses a 2 of 3
// multisig which is the same configuration as in the mainnet of the Mainstay service
func TestAttestService_Multi(t *testing.T) {

	// Test INIT
	test := test.NewTestMulti()
	configs := test.Configs
	config := configs[0]

	// randomly test with invalid config here
	// timing config no effect on server
	for _, config := range configs {
		timingConfig := confpkg.TimingConfig{-1, -1}
		config.SetTimingConfig(timingConfig)
	}

	dbFake := server.NewDbFake()
	server := server.NewServer(dbFake)
	signerSingle := NewAttestSignerFake([]*confpkg.Config{config})
	attestService := NewAttestService(nil, nil, server, signerSingle, config)

	// Test initial state of attest service
	verifyStateInit(t, attestService)
	// Test AStateInit -> AStateNextCommitment
	verifyStateInitToNextCommitment(t, attestService)

	// Test AStateInit -> AStateError
	// error case when server latest commitment not set
	// need to re-initiate attestation and set latest commitment in server
	attestService.doAttestation()
	assert.Equal(t, AStateError, attestService.state)
	assert.Equal(t, errors.New(models.ErrorCommitmentListEmpty), attestService.errorState)
	assert.Equal(t, ATimeFixed, attestDelay)

	// Test AStateError -> AStateInit -> AStateNextCommitment again
	attestService.doAttestation()
	verifyStateInit(t, attestService)
	verifyStateInitToNextCommitment(t, attestService)

	// Test AStateNextCommitment -> AStateNewAttestation
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment := verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

	// Test AStateNewAttestation -> AStateSignAttestation
	verifyStateNewAttestationToSignAttestation(t, attestService)

	// test failure at GetSigs()
	// use singerSingle first and notice that transaction signing fails
	attestService.doAttestation()
	assert.Equal(t, AStateError, attestService.state)
	assert.Equal(t, errors.New(ErrorSigsMissingForVin), attestService.errorState)
	assert.Equal(t, ATimeFixed, attestDelay)

	// set signer to the correct signerMulti that does multiple signings
	// and observe that attestation creation and signing now succeeds
	signerMulti := NewAttestSignerFake(configs)
	attestService.signer = signerMulti
	attestService.doAttestation()

	// Test initial state of attest service
	// skip testing this as service has not actually reset yet
	//verifyStateInit(t, attestService)

	// Test AStateInit -> AStateNextCommitment
	verifyStateInitToNextCommitment(t, attestService)
	// use same commitment as nothing changed
	latestCommitment = verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

	// Test AStateNewAttestation -> AStateSignAttestation
	verifyStateNewAttestationToSignAttestation(t, attestService)
	// Test AStateSignAttestation -> AStatePreSendStore
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test AStatePreSendStore -> AStateSendAttestation
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test AStateSendAttestation -> AStateAwaitConfirmation
	txid := verifyStateSendAttestationToAwaitConfirmation(t, attestService)
	// Test AStateAwaitConfirmation -> AStateAwaitConfirmation
	verifyStateAwaitConfirmationToAwaitConfirmation(t, attestService)
	// Test AStateAwaitConfirmation -> AStateAwaitConfirmation
	verifyStateAwaitConfirmationToAwaitConfirmation(t, attestService)
	// Test AStateAwaitConfirmation -> AStateNextCommitment
	config.MainClient().Generate(1)
	verifyStateAwaitConfirmationToNextCommitment(t, attestService, config, txid, DefaultATimeNewAttestation)

	// Test AStateNextCommitment -> AStateNextCommitment
	attestService.doAttestation()
	assert.Equal(t, AStateNextCommitment, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	assert.Equal(t, DefaultATimeNewAttestation, attestDelay)

	// Test AStateNextCommitment -> AStateNewAttestation
	// stuck in next commitment
	// need to update server latest commitment
	hashY, _ := chainhash.NewHashFromStr("baaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment = verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashY)

	// Test AStateNewAttestation -> AStateSignAttestation
	verifyStateNewAttestationToSignAttestation(t, attestService)
	// Test AStateSignAttestation -> AStatePreSendStore
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test AStatePreSendStore -> AStateSendAttestation
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test AStateSendAttestation -> AStateAwaitConfirmation
	txid = verifyStateSendAttestationToAwaitConfirmation(t, attestService)
	// Test AStateAwaitConfirmation -> AStateNextCommitment
	config.MainClient().Generate(1)
	verifyStateAwaitConfirmationToNextCommitment(t, attestService, config, txid, DefaultATimeNewAttestation)
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
	attestService := NewAttestService(nil, nil, server, NewAttestSignerFake([]*confpkg.Config{config}), config)

	// Test initial state of attest service
	verifyStateInit(t, attestService)
	// Test AStateInit -> AStateNextCommitment
	verifyStateInitToNextCommitment(t, attestService)

	// Test AStateInit -> AStateError
	// error case when server latest commitment not set
	// need to re-initiate attestation and set latest commitment in server
	attestService.doAttestation()
	assert.Equal(t, AStateError, attestService.state)
	assert.Equal(t, errors.New(models.ErrorCommitmentListEmpty), attestService.errorState)
	assert.Equal(t, ATimeFixed, attestDelay)

	// Test AStateError -> AStateInit -> AStateNextCommitment again
	attestService.doAttestation()
	verifyStateInit(t, attestService)
	verifyStateInitToNextCommitment(t, attestService)

	// Test AStateNextCommitment -> AStateNewAttestation
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment := verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

	// Test AStateNewAttestation -> AStateSignAttestation
	verifyStateNewAttestationToSignAttestation(t, attestService)
	// Test AStateSignAttestation -> AStatePreSendStore
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test AStatePreSendStore -> AStateSendAttestation
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test AStateSendAttestation -> AStateAwaitConfirmation
	txid := verifyStateSendAttestationToAwaitConfirmation(t, attestService)
	// Test AStateAwaitConfirmation -> AStateAwaitConfirmation
	verifyStateAwaitConfirmationToAwaitConfirmation(t, attestService)
	// Test AStateAwaitConfirmation -> AStateAwaitConfirmation
	verifyStateAwaitConfirmationToAwaitConfirmation(t, attestService)
	// Test AStateAwaitConfirmation -> AStateNextCommitment
	config.MainClient().Generate(1)
	verifyStateAwaitConfirmationToNextCommitment(t, attestService, config, txid, DefaultATimeNewAttestation)

	// Test AStateNextCommitment -> AStateNextCommitment
	attestService.doAttestation()
	assert.Equal(t, AStateNextCommitment, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	assert.Equal(t, DefaultATimeNewAttestation, attestDelay)

	// Test AStateNextCommitment -> AStateNewAttestation
	// stuck in next commitment
	// need to update server latest commitment
	hashY, _ := chainhash.NewHashFromStr("baaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment = verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashY)

	// Test AStateNewAttestation -> AStateSignAttestation
	verifyStateNewAttestationToSignAttestation(t, attestService)
	// Test AStateSignAttestation -> AStatePreSendStore
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test AStatePreSendStore -> AStateSendAttestation
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test AStateSendAttestation -> AStateAwaitConfirmation
	txid = verifyStateSendAttestationToAwaitConfirmation(t, attestService)
	// Test AStateAwaitConfirmation -> AStateNextCommitment
	config.MainClient().Generate(1)
	verifyStateAwaitConfirmationToNextCommitment(t, attestService, config, txid, DefaultATimeNewAttestation)
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
	attestService := NewAttestService(nil, nil, server, NewAttestSignerFake([]*confpkg.Config{config}), config)

	attestService.attester.Fees.ResetFee(true)

	// Test initial state of attest service
	verifyStateInit(t, attestService)
	// Test AStateInit -> AStateNextCommitment
	verifyStateInitToNextCommitment(t, attestService)

	// Test AStateNextCommitment -> AStateNewAttestation
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	_ = verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

	// Test AStateNewAttestation -> AStateSignAttestation
	verifyStateNewAttestationToSignAttestation(t, attestService)
	assert.Equal(t, attestService.attester.Fees.minFee, attestService.attester.Fees.GetFee())
	// Test AStateSignAttestation -> AStatePreSendStore
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test AStatePreSendStore -> AStateSendAttestation
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test AStateSendAttestation -> AStateAwaitConfirmation
	txid := verifyStateSendAttestationToAwaitConfirmation(t, attestService)

	// set confirm time back to test what happens in handle unconfirmed case
	confirmTime = confirmTime.Add(-time.Duration(customAtimeHandleUnconfirmed) * time.Minute)

	// Test AStateAwaitConfirmation -> AStateHandleUnconfirmed
	verifyStateAwaitConfirmationToHandleUnconfirmed(t, attestService)
	// Test AStateHandleUnconfirmed -> AStateSignAttestation
	verifyStateHandleUnconfirmedToSignAttestation(t, attestService)
	assert.Equal(t, attestService.attester.Fees.minFee+attestService.attester.Fees.feeIncrement,
		attestService.attester.Fees.GetFee())

	// Test AStateSignAttestation -> AStatePreSendStore
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test AStatePreSendStore -> AStateSendAttestation
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test AStateSendAttestation -> AStateAwaitConfirmation
	txid = verifyStateSendAttestationToAwaitConfirmation(t, attestService)
	// Test AStateAwaitConfirmation -> AStateNextCommitment
	config.MainClient().Generate(1)
	verifyStateAwaitConfirmationToNextCommitment(t, attestService, config, txid,
		time.Duration(customAtimeNewAttestation)*time.Minute)
	assert.Equal(t, attestService.attester.Fees.minFee, attestService.attester.Fees.GetFee())

	// Test AStateNextCommitment -> AStateNewAttestation
	// set server commitment before creationg new attestation
	hashY, _ := chainhash.NewHashFromStr("baaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	_ = verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashY)

	// add also unspent this time
	_ = createTopupUnspent(t, test.Config)
	attestService.attester.MainClient.Generate(1)

	// Test AStateNewAttestation -> AStateSignAttestation
	attestService.doAttestation()
	assert.Equal(t, AStateSignAttestation, attestService.state)
	// cant test much more here - we test this in other unit tests
	assert.Equal(t, 2, len(attestService.attestation.Tx.TxIn))
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxOut))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[0].SignatureScript))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[1].SignatureScript))
	assert.Equal(t, ATimeSigs, attestDelay)
	assert.Equal(t, attestService.attester.Fees.minFee, attestService.attester.Fees.GetFee())
	// Test AStateSignAttestation -> AStatePreSendStore
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test AStatePreSendStore -> AStateSendAttestation
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test AStateSendAttestation -> AStateAwaitConfirmation
	txid = verifyStateSendAttestationToAwaitConfirmation(t, attestService)

	// set confirm time back to test what happens in handle unconfirmed case
	confirmTime = confirmTime.Add(-time.Duration(customAtimeHandleUnconfirmed) * time.Minute)

	// Test AStateAwaitConfirmation -> AStateHandleUnconfirmed
	verifyStateAwaitConfirmationToHandleUnconfirmed(t, attestService)
	// Test AStateHandleUnconfirmed -> AStateSignAttestation
	attestService.doAttestation()
	assert.Equal(t, AStateSignAttestation, attestService.state)
	// cant test much more here - we test this in other unit tests
	assert.Equal(t, 2, len(attestService.attestation.Tx.TxIn))
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxOut))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[0].SignatureScript))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[1].SignatureScript))
	assert.Equal(t, ATimeSigs, attestDelay)
	assert.Equal(t, attestService.attester.Fees.minFee+attestService.attester.Fees.feeIncrement,
		attestService.attester.Fees.GetFee())

	// Test AStateSignAttestation -> AStatePreSendStore
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test AStatePreSendStore -> AStateSendAttestation
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test AStateSendAttestation -> AStateAwaitConfirmation
	txid = verifyStateSendAttestationToAwaitConfirmation(t, attestService)
	// Test AStateAwaitConfirmation -> AStateNextCommitment
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
	attestService := NewAttestService(nil, nil, server, NewAttestSignerFake([]*confpkg.Config{config}), config)

	// Test initial state of attest service
	verifyStateInit(t, attestService)
	// Test AStateInit -> AStateNextCommitment
	verifyStateInitToNextCommitment(t, attestService)

	// Test AStateNextCommitment -> AStateNewAttestation
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	_ = verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

	// Test AStateNewAttestation -> AStateSignAttestation
	verifyStateNewAttestationToSignAttestation(t, attestService)
	// Test AStateSignAttestation -> AStatePreSendStore
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test AStatePreSendStore -> AStateSendAttestation
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test AStateSendAttestation -> AStateAwaitConfirmation
	txid := verifyStateSendAttestationToAwaitConfirmation(t, attestService)
	// Test AStateAwaitConfirmation -> AStateAwaitConfirmation
	verifyStateAwaitConfirmationToAwaitConfirmation(t, attestService)
	// Test AStateAwaitConfirmation -> AStateAwaitConfirmation
	verifyStateAwaitConfirmationToAwaitConfirmation(t, attestService)
	// Test AStateAwaitConfirmation -> AStateNextCommitment
	config.MainClient().Generate(1)
	verifyStateAwaitConfirmationToNextCommitment(t, attestService, config, txid, DefaultATimeNewAttestation)

	// Test AStateNextCommitment -> AStateNewAttestation
	// stuck in next commitment
	// need to update server latest commitment
	hashY, _ := chainhash.NewHashFromStr("baaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	_ = verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashY)

	// create top up unspent
	_ = createTopupUnspent(t, test.Config)
	attestService.attester.MainClient.Generate(1)

	// Test AStateNewAttestation -> AStateSignAttestation
	attestService.doAttestation()
	assert.Equal(t, AStateSignAttestation, attestService.state)
	// cant test much more here - we test this in other unit tests
	assert.Equal(t, 2, len(attestService.attestation.Tx.TxIn))
	assert.Equal(t, 1, len(attestService.attestation.Tx.TxOut))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[0].SignatureScript))
	assert.Equal(t, 0, len(attestService.attestation.Tx.TxIn[1].SignatureScript))
	assert.Equal(t, ATimeSigs, attestDelay)

	// Test AStateSignAttestation -> AStatePreSendStore
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test AStatePreSendStore -> AStateSendAttestation
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test AStateSendAttestation -> AStateAwaitConfirmation
	txid = verifyStateSendAttestationToAwaitConfirmation(t, attestService)
	// Test AStateAwaitConfirmation -> AStateNextCommitment
	config.MainClient().Generate(1)
	verifyStateAwaitConfirmationToNextCommitment(t, attestService, config, txid, DefaultATimeNewAttestation)

	// Test AStateNextCommitment -> AStateNewAttestation
	// set server commitment before creationg new attestation
	hashZ, _ := chainhash.NewHashFromStr("caaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	_ = verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashZ)

	// Test AStateNewAttestation -> AStateSignAttestation
	verifyStateNewAttestationToSignAttestation(t, attestService)
	// Test AStateSignAttestation -> AStatePreSendStore
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test AStatePreSendStore -> AStateSendAttestation
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test AStateSendAttestation -> AStateAwaitConfirmation
	txid = verifyStateSendAttestationToAwaitConfirmation(t, attestService)
	// Test AStateAwaitConfirmation -> AStateAwaitConfirmation
	verifyStateAwaitConfirmationToAwaitConfirmation(t, attestService)
	// Test AStateAwaitConfirmation -> AStateAwaitConfirmation
	verifyStateAwaitConfirmationToAwaitConfirmation(t, attestService)
	// Test AStateAwaitConfirmation -> AStateNextCommitment
	config.MainClient().Generate(1)
	verifyStateAwaitConfirmationToNextCommitment(t, attestService, config, txid, DefaultATimeNewAttestation)
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
	attestService := NewAttestService(nil, nil, server, NewAttestSignerFake([]*confpkg.Config{config}), config)

	// Test initial state of attest service
	verifyStateInit(t, attestService)
	verifyStateInitWalletFailure(t, attestService)
	// Test AStateInit -> AStateNextCommitment
	verifyStateInitToNextCommitment(t, attestService)

	// failure - re init attestation service with restart
	attestService = NewAttestService(nil, nil, server, NewAttestSignerFake([]*confpkg.Config{config}), config)
	// Test AStateInit -> AStateNextCommitment again
	verifyStateInitToNextCommitment(t, attestService)

	// failure - re init attestation service from state failure
	attestService.state = AStateInit
	// Test AStateInit -> AStateNextCommitment again
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
	attestService := NewAttestService(nil, nil, server, NewAttestSignerFake([]*confpkg.Config{config}), config)

	// Test initial state of attest service
	verifyStateInit(t, attestService)
	verifyStateInitWalletFailure(t, attestService)
	// Test AStateInit -> AStateNextCommitment
	verifyStateInitToNextCommitment(t, attestService)

	// Test AStateNextCommitment -> AStateNewAttestation
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment := verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

	// failure - re init attestation service
	attestService = NewAttestService(nil, nil, server, NewAttestSignerFake([]*confpkg.Config{config}), config)
	// Test AStateInit -> AStateNextCommitment
	verifyStateInitToNextCommitment(t, attestService)
	// Test AStateNextCommitment -> AStateNewAttestation
	attestService.doAttestation()
	assert.Equal(t, AStateNewAttestation, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())

	// failure - re init attestation service from inner state failure
	attestService.state = AStateInit
	// Test AStateInit -> AStateNextCommitment
	verifyStateInitToNextCommitment(t, attestService)
	// Test AStateNextCommitment -> AStateNewAttestation
	attestService.doAttestation()
	assert.Equal(t, AStateNewAttestation, attestService.state)
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
	attestService := NewAttestService(nil, nil, server, NewAttestSignerFake([]*confpkg.Config{config}), config)

	// Test initial state of attest service
	verifyStateInit(t, attestService)
	verifyStateInitWalletFailure(t, attestService)
	// Test AStateInit -> AStateNextCommitment
	verifyStateInitToNextCommitment(t, attestService)

	// Test AStateNextCommitment -> AStateNewAttestation
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment := verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

	// Test AStateNewAttestation -> AStateSignAttestation
	verifyStateNewAttestationToSignAttestation(t, attestService)

	// failure - re init attestation service
	attestService = NewAttestService(nil, nil, server, NewAttestSignerFake([]*confpkg.Config{config}), config)
	// Test AStateInit -> AStateNextCommitment
	verifyStateInitToNextCommitment(t, attestService)
	// Test AStateNextCommitment -> AStateNewAttestation
	// set server commitment before creationg new attestation
	attestService.doAttestation()
	assert.Equal(t, AStateNewAttestation, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	// Test AStateNewAttestation -> AStateSignAttestation
	verifyStateNewAttestationToSignAttestation(t, attestService)

	// failure - re init attestation service from inner state failure
	attestService.state = AStateInit
	// Test AStateInit -> AStateNextCommitment
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
	attestService := NewAttestService(nil, nil, server, NewAttestSignerFake([]*confpkg.Config{config}), config)

	// Test initial state of attest service
	verifyStateInit(t, attestService)
	verifyStateInitWalletFailure(t, attestService)

	// Test AStateInit -> AStateNextCommitment
	verifyStateInitToNextCommitment(t, attestService)

	// Test AStateNextCommitment -> AStateNewAttestation
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment := verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

	// Test AStateNewAttestation -> AStateSignAttestation
	verifyStateNewAttestationToSignAttestation(t, attestService)
	// Test AStateSignAttestation -> AStatePreSendStore
	verifyStateSignAttestationToPreSendStore(t, attestService)

	// failure - re init attestation service
	attestService = NewAttestService(nil, nil, server, NewAttestSignerFake([]*confpkg.Config{config}), config)

	// Test AStateInit -> AStateNextCommitment
	verifyStateInitToNextCommitment(t, attestService)
	// Test AStateNextCommitment -> AStateNewAttestation
	// set server commitment before creationg new attestation
	attestService.doAttestation()
	assert.Equal(t, AStateNewAttestation, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	// Test AStateNewAttestation -> AStateSignAttestation
	verifyStateNewAttestationToSignAttestation(t, attestService)
	// Test AStateSignAttestation -> AStatePreSendStore
	verifyStateSignAttestationToPreSendStore(t, attestService)

	// failure - re init attestation service from inner state failure
	attestService.state = AStateInit
	// Test AStateInit -> AStateNextCommitment
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
	attestService := NewAttestService(nil, nil, server, NewAttestSignerFake([]*confpkg.Config{config}), config)

	// Test initial state of attest service
	verifyStateInit(t, attestService)
	verifyStateInitWalletFailure(t, attestService)

	// Test AStateInit -> AStateNextCommitment
	verifyStateInitToNextCommitment(t, attestService)

	// Test AStateNextCommitment -> AStateNewAttestation
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment := verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

	// Test AStateNewAttestation -> AStateSignAttestation
	verifyStateNewAttestationToSignAttestation(t, attestService)
	// Test AStateSignAttestation -> AStatePreSendStore
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test AStatePreSendStore -> AStateSendAttestation
	verifyStatePreSendStoreToSendAttestation(t, attestService)

	// failure - re init attestation service
	attestService = NewAttestService(nil, nil, server, NewAttestSignerFake([]*confpkg.Config{config}), config)

	// Test AStateInit -> AStateNextCommitment
	verifyStateInitToNextCommitment(t, attestService)
	// Test AStateNextCommitment -> AStateNewAttestation
	// set server commitment before creationg new attestation
	attestService.doAttestation()
	assert.Equal(t, AStateNewAttestation, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	// Test AStateNewAttestation -> AStateSignAttestation
	verifyStateNewAttestationToSignAttestation(t, attestService)
	// Test AStateSignAttestation -> AStatePreSendStore
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test AStatePreSendStore -> AStateSendAttestation
	verifyStatePreSendStoreToSendAttestation(t, attestService)

	// failure - re init attestation service from inner state failure
	attestService.state = AStateInit
	// Test AStateInit -> AStateNextCommitment
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
		attestService := NewAttestService(nil, nil, server, NewAttestSignerFake([]*confpkg.Config{config}), config)

		// Test initial state of attest service
		verifyStateInit(t, attestService)
		// Test AStateInit -> AStateNextCommitment
		attestService.doAttestation()
		assert.Equal(t, AStateNextCommitment, attestService.state)
		assert.Equal(t, prevAttestation.CommitmentHash(), attestService.attestation.CommitmentHash())
		assert.Equal(t, prevAttestation.Txid, attestService.attestation.Txid)
		assert.Equal(t, prevAttestation.Confirmed, attestService.attestation.Confirmed)
		assert.Equal(t, prevAttestation.Info, attestService.attestation.Info)
		assert.Equal(t, ATimeFixed, attestDelay)

		// Test AStateNextCommitment -> AStateNewAttestation
		// set server commitment before creationg new attestation
		hashX, _ := chainhash.NewHashFromStr(fmt.Sprintf("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b%d", i))
		latestCommitment := verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

		// Test AStateNewAttestation -> AStateSignAttestation
		verifyStateNewAttestationToSignAttestation(t, attestService)
		// Test AStateSignAttestation -> AStatePreSendStore
		verifyStateSignAttestationToPreSendStore(t, attestService)
		// Test AStatePreSendStore -> AStateSendAttestation
		verifyStatePreSendStoreToSendAttestation(t, attestService)
		// Test AStateSendAttestation -> AStateAwaitConfirmation
		txid := verifyStateSendAttestationToAwaitConfirmation(t, attestService)

		// failure - re init attestation service
		attestService = NewAttestService(nil, nil, server, NewAttestSignerFake([]*confpkg.Config{config}), config)

		// Test AStateInit -> AStateAwaitConfirmation
		verifyStateInitToAwaitConfirmation(t, attestService, latestCommitment, txid)

		// generate new block to confirm attestation
		config.MainClient().Generate(1)
		rawTx, _ := config.MainClient().GetRawTransaction(&txid)
		walletTx, _ := config.MainClient().GetTransaction(&txid)
		// Test AStateAwaitConfirmation -> AStateNextCommitment
		attestService.doAttestation()
		assert.Equal(t, AStateNextCommitment, attestService.state)
		assert.Equal(t, true, attestService.attestation.Confirmed)
		assert.Equal(t, txid, attestService.attestation.Txid)
		assert.Equal(t, models.AttestationInfo{
			Txid:      txid.String(),
			Blockhash: walletTx.BlockHash,
			Amount:    rawTx.MsgTx().TxOut[0].Value,
			Time:      walletTx.Time}, attestService.attestation.Info)

		// failure - re init attestation service from inner state failure
		attestService.state = AStateInit

		// Test AStateInit -> AStateAwaitConfirmation
		attestService.doAttestation()
		assert.Equal(t, AStateNextCommitment, attestService.state)
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
	attestService := NewAttestService(nil, nil, server, NewAttestSignerFake([]*confpkg.Config{config}), config)

	// Test initial state of attest service
	verifyStateInit(t, attestService)
	verifyStateInitWalletFailure(t, attestService)
	// Test AStateInit -> AStateNextCommitment
	verifyStateInitToNextCommitment(t, attestService)

	// Test AStateNextCommitment -> AStateNewAttestation
	// set server commitment before creationg new attestation
	hashX, _ := chainhash.NewHashFromStr("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latestCommitment := verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

	// Test AStateNewAttestation -> AStateSignAttestation
	verifyStateNewAttestationToSignAttestation(t, attestService)
	// Test AStateSignAttestation -> AStatePreSendStore
	verifyStateSignAttestationToPreSendStore(t, attestService)
	// Test AStatePreSendStore -> AStateSendAttestation
	verifyStatePreSendStoreToSendAttestation(t, attestService)
	// Test AStateSendAttestation -> AStateAwaitConfirmation
	txid := verifyStateSendAttestationToAwaitConfirmation(t, attestService)

	// generate new block to confirm attestation
	config.MainClient().Generate(1)
	rawTx, _ := config.MainClient().GetRawTransaction(&txid)
	walletTx, _ := config.MainClient().GetTransaction(&txid)
	// Test AStateAwaitConfirmation -> AStateNextCommitment
	attestService.doAttestation()
	assert.Equal(t, AStateNextCommitment, attestService.state)
	assert.Equal(t, true, attestService.attestation.Confirmed)
	assert.Equal(t, txid, attestService.attestation.Txid)
	assert.Equal(t, models.AttestationInfo{
		Txid:      txid.String(),
		Blockhash: walletTx.BlockHash,
		Amount:    rawTx.MsgTx().TxOut[0].Value,
		Time:      walletTx.Time}, attestService.attestation.Info)

	// failure - re init attestation service
	attestService = NewAttestService(nil, nil, server, NewAttestSignerFake([]*confpkg.Config{config}), config)
	// Test AStateInit -> AStateNextCommitment
	attestService.doAttestation()
	assert.Equal(t, AStateNextCommitment, attestService.state)
	assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
	assert.Equal(t, txid, attestService.attestation.Txid)
	assert.Equal(t, true, attestService.attestation.Confirmed)
	assert.Equal(t, models.AttestationInfo{
		Txid:      txid.String(),
		Blockhash: walletTx.BlockHash,
		Amount:    rawTx.MsgTx().TxOut[0].Value,
		Time:      walletTx.Time}, attestService.attestation.Info)

	// failure - re init attestation service from inner state
	attestService.state = AStateInit
	// Test AStateInit -> AStateNextCommitment
	attestService.doAttestation()
	assert.Equal(t, AStateNextCommitment, attestService.state)
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
		attestService := NewAttestService(nil, nil, server, NewAttestSignerFake([]*confpkg.Config{config}), config)

		attestService.attester.Fees.ResetFee(true)

		// Test initial state of attest service
		verifyStateInit(t, attestService)
		// Test AStateInit -> AStateNextCommitment
		attestService.doAttestation()
		assert.Equal(t, AStateNextCommitment, attestService.state)
		assert.Equal(t, prevAttestation.CommitmentHash(), attestService.attestation.CommitmentHash())
		assert.Equal(t, prevAttestation.Txid, attestService.attestation.Txid)
		assert.Equal(t, prevAttestation.Confirmed, attestService.attestation.Confirmed)
		assert.Equal(t, prevAttestation.Info, attestService.attestation.Info)
		assert.Equal(t, ATimeFixed, attestDelay)

		// Test AStateNextCommitment -> AStateNewAttestation
		// set server commitment before creationg new attestation
		hashX, _ := chainhash.NewHashFromStr(fmt.Sprintf("aaaaaaa1111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b%d", i))
		latestCommitment := verifyStateNextCommitmentToNewAttestation(t, attestService, dbFake, hashX)

		// Test AStateNewAttestation -> AStateSignAttestation
		verifyStateNewAttestationToSignAttestation(t, attestService)
		assert.Equal(t, attestService.attester.Fees.minFee, attestService.attester.Fees.GetFee())
		// Test AStateSignAttestation -> AStatePreSendStore
		verifyStateSignAttestationToPreSendStore(t, attestService)
		// Test AStatePreSendStore -> AStateSendAttestation
		verifyStatePreSendStoreToSendAttestation(t, attestService)
		// Test AStateSendAttestation -> AStateAwaitConfirmation
		txid := verifyStateSendAttestationToAwaitConfirmation(t, attestService)

		// set confirm time back to test what happens in handle unconfirmed case
		confirmTime = confirmTime.Add(-DefaultATimeHandleUnconfirmed)

		// Test AStateAwaitConfirmation -> AStateHandleUnconfirmed
		verifyStateAwaitConfirmationToHandleUnconfirmed(t, attestService)
		// Test AStateHandleUnconfirmed -> AStateSignAttestation
		verifyStateHandleUnconfirmedToSignAttestation(t, attestService)
		assert.Equal(t, attestService.attester.Fees.minFee+attestService.attester.Fees.feeIncrement,
			attestService.attester.Fees.GetFee())

		// failure - re init attestation service with restart
		attestService = NewAttestService(nil, nil, server, NewAttestSignerFake([]*confpkg.Config{config}), config)
		attestService.attester.Fees.ResetFee(true)
		// Test AStateInit -> AStateAwaitConfirmation
		verifyStateInitToAwaitConfirmation(t, attestService, latestCommitment, txid)

		// failure - re init attestation service from inner state failure
		attestService.state = AStateInit
		// Test AStateInit -> AStateAwaitConfirmation
		verifyStateInitToAwaitConfirmation(t, attestService, latestCommitment, txid)
		// set confirm time back to test what happens in handle unconfirmed case
		confirmTime = confirmTime.Add(-DefaultATimeHandleUnconfirmed)

		// Test AStateAwaitConfirmation -> AStateHandleUnconfirmed
		verifyStateAwaitConfirmationToHandleUnconfirmed(t, attestService)
		// Test AStateHandleUnconfirmed -> AStateSignAttestation
		verifyStateHandleUnconfirmedToSignAttestation(t, attestService)
		assert.Equal(t, attestService.attester.Fees.minFee+attestService.attester.Fees.feeIncrement,
			attestService.attester.Fees.GetFee())

		// Test AStateSignAttestation -> AStatePreSendStore
		verifyStateSignAttestationToPreSendStore(t, attestService)
		// Test AStatePreSendStore -> AStateSendAttestation
		verifyStatePreSendStoreToSendAttestation(t, attestService)

		// failure - re init attestation service with restart
		attestService = NewAttestService(nil, nil, server, NewAttestSignerFake([]*confpkg.Config{config}), config)
		attestService.attester.Fees.ResetFee(true)
		// Test AStateInit -> AStateAwaitConfirmation
		verifyStateInitToAwaitConfirmation(t, attestService, latestCommitment, txid)

		// failure - re init attestation service from inner state failure
		attestService.state = AStateInit
		// Test AStateInit -> AStateAwaitConfirmation
		verifyStateInitToAwaitConfirmation(t, attestService, latestCommitment, txid)
		// set confirm time back to test what happens in handle unconfirmed case
		confirmTime = confirmTime.Add(-DefaultATimeHandleUnconfirmed)

		// Test AStateAwaitConfirmation -> AStateHandleUnconfirmed
		verifyStateAwaitConfirmationToHandleUnconfirmed(t, attestService)
		// Test AStateHandleUnconfirmed -> AStateSignAttestation
		verifyStateHandleUnconfirmedToSignAttestation(t, attestService)
		assert.Equal(t, attestService.attester.Fees.minFee+attestService.attester.Fees.feeIncrement,
			attestService.attester.Fees.GetFee())

		// Test AStateSignAttestation -> AStatePreSendStore
		verifyStateSignAttestationToPreSendStore(t, attestService)
		// Test AStatePreSendStore -> AStateSendAttestation
		verifyStatePreSendStoreToSendAttestation(t, attestService)
		// Test AStateSendAttestation -> AStateAwaitConfirmation
		txid = verifyStateSendAttestationToAwaitConfirmation(t, attestService)
		// Test AStateInit -> AStateAwaitConfirmation
		attestService.doAttestation()
		assert.Equal(t, AStateAwaitConfirmation, attestService.state)
		assert.Equal(t, latestCommitment.GetCommitmentHash(), attestService.attestation.CommitmentHash())
		assert.Equal(t, txid, attestService.attestation.Txid)
		assert.Equal(t, false, attestService.attestation.Confirmed)
		assert.Equal(t, models.AttestationInfo{}, attestService.attestation.Info)

		// failure - re init attestation service from inner state failure
		attestService.state = AStateInit
		// Test AStateInit -> AStateAwaitConfirmation
		verifyStateInitToAwaitConfirmation(t, attestService, latestCommitment, txid)

		// generate new block to confirm attestation
		config.MainClient().Generate(1)
		rawTx, _ := config.MainClient().GetRawTransaction(&txid)
		walletTx, _ := config.MainClient().GetTransaction(&txid)
		// Test AStateAwaitConfirmation -> AStateNextCommitment
		attestService.doAttestation()
		assert.Equal(t, AStateNextCommitment, attestService.state)
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
