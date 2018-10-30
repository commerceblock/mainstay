package server

import (
	"context"
	"mainstay/clients"
	"mainstay/models"
	"mainstay/requestapi"
	"mainstay/test"
	"sync"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/assert"
)

// Test Server responses to attestation service Requests
func TestServerRespondAttestation(t *testing.T) {
	// TEST INIT
	test := test.NewTest(false, false)
	sideClientFake := test.Config.OceanClient().(*clients.SidechainClientFake)
	server := NewServer(nil, nil, test.Config)

	// Generate blocks in side chain and update server latest
	sideClientFake.Generate(10)
	bestblockhash, _ := sideClientFake.GetBestBlockHash()
	server.latestCommitment = *bestblockhash

	// Test ATTESTATION_COMMITMENT request
	reqCommitment := requestapi.BaseRequest{}
	reqCommitment.SetRequestType(requestapi.ATTESTATION_COMMITMENT)
	respCommitment, _ := server.respondAttestation(reqCommitment).(requestapi.AttestastionCommitmentResponse)
	assert.Equal(t, "", respCommitment.ResponseError())
	assert.Equal(t, *bestblockhash, respCommitment.Commitment)

	// Test ATTESTATION_LATEST request
	reqLatest := requestapi.BaseRequest{}
	reqLatest.SetRequestType(requestapi.ATTESTATION_LATEST)
	respLatest, _ := server.respondAttestation(reqLatest).(requestapi.AttestationLatestResponse)
	assert.Equal(t, "", respLatest.ResponseError())
	assert.Equal(t, chainhash.Hash{}, respLatest.Attestation.Txid)
	assert.Equal(t, chainhash.Hash{}, respLatest.Attestation.AttestedHash)

	// Generate new attestation and update server
	txid, _ := chainhash.NewHashFromStr("11111111111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latest := models.NewAttestation(*txid, *bestblockhash)

	// Test ATTESTATION_UPDATE
	reqUpdate := requestapi.AttestationUpdateRequest{Attestation: *latest}
	reqUpdate.SetRequestType(requestapi.ATTESTATION_UPDATE)
	respUpdate, _ := server.respondAttestation(reqUpdate).(requestapi.AttestationUpdateResponse)
	assert.Equal(t, "", respUpdate.ResponseError())
	assert.Equal(t, true, respUpdate.Updated)
	assert.Equal(t, *txid, server.latestAttestation.Txid)
	assert.Equal(t, *bestblockhash, server.latestAttestation.AttestedHash)

	// Test ATTESTATION_LATEST request again after update
	respLatest2, _ := server.respondAttestation(reqLatest).(requestapi.AttestationLatestResponse)
	assert.Equal(t, "", respLatest2.ResponseError())
	assert.Equal(t, *txid, respLatest2.Attestation.Txid)
	assert.Equal(t, *bestblockhash, respLatest2.Attestation.AttestedHash)
}

// Test Server
func TestServer(t *testing.T) {
	// TEST INIT
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	test := test.NewTest(false, false)
	sideClientFake := test.Config.OceanClient().(*clients.SidechainClientFake)
	server := NewServer(ctx, wg, test.Config)

	attChannel := server.AttestationServiceChannel()

	wg.Add(1)
	go server.Run()

	// Generate blocks in side chain and update server latest
	sideClientFake.Generate(10)
	bestblockhash, _ := sideClientFake.GetBestBlockHash()
	server.latestCommitment = *bestblockhash

	// Generate new attestation and update server
	txid, _ := chainhash.NewHashFromStr("11111111111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latest := models.NewAttestation(*txid, *bestblockhash)

	// Test requestion ATTESTATION_UPDATE through AttestationServiceChannel
	responseChan := make(chan requestapi.Response)
	reqUpdate := requestapi.AttestationUpdateRequest{Attestation: *latest}
	reqUpdate.SetRequestType(requestapi.ATTESTATION_UPDATE)
	attChannel <- requestapi.RequestWithResponseChannel{reqUpdate, responseChan}
	responseUpdate := (<-responseChan).(requestapi.AttestationUpdateResponse)
	assert.Equal(t, "", responseUpdate.ResponseError())
	assert.Equal(t, true, responseUpdate.Updated)
	assert.Equal(t, *txid, server.latestAttestation.Txid)
	assert.Equal(t, *bestblockhash, server.latestAttestation.AttestedHash)

	// Test requestion ATTESTATION_LATEST through AttestationServiceChannel
	reqLatest := requestapi.BaseRequest{}
	reqLatest.SetRequestType(requestapi.ATTESTATION_LATEST)
	attChannel <- requestapi.RequestWithResponseChannel{reqLatest, responseChan}
	responseLatest := (<-responseChan).(requestapi.AttestationLatestResponse)
	assert.Equal(t, "", responseLatest.ResponseError())
	assert.Equal(t, *txid, responseLatest.Attestation.Txid)
	assert.Equal(t, *bestblockhash, responseLatest.Attestation.AttestedHash)

	cancel() // shut server down
}
