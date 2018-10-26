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

// Test Server responses to request service Requests
func TestServerRespond(t *testing.T) {
	// TEST INIT
	test := test.NewTest(false, false)
	sideClientFake := test.Config.OceanClient().(*clients.SidechainClientFake)
	server := NewServer(nil, nil, test.Config)

	// Generate blocks in side chain and update server latest
	sideClientFake.Generate(10)
	bestblockhash, _ := sideClientFake.GetBestBlockHash()
	server.latestCommitment = *bestblockhash

	// Generate new attestation and update server
	sidehash := server.latestCommitment
	txid, _ := chainhash.NewHashFromStr("11111111111d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7")
	latest := models.NewAttestation(*txid, sidehash)

	// Test updateLatest
	server.updateLatest(*latest)
	assert.Equal(t, *txid, server.latestAttestation.Txid)
	assert.Equal(t, sidehash, server.latestAttestation.AttestedHash)
	assert.Equal(t, int32(10), server.latestHeight)

	// Test respond for various Request/Response
	bestblockhash, _ = server.sideClient.GetBestBlockHash()

	req := requestapi.BaseRequest{}
	req.SetRequestType(requestapi.SERVER_BEST_BLOCK)
	resp1, _ := server.respond(req).(requestapi.BestBlockResponse)
	assert.Equal(t, "", resp1.ResponseError())
	assert.Equal(t, bestblockhash.String(), resp1.BlockHash)

	req.SetRequestType(requestapi.SERVER_BEST_BLOCK_HEIGHT)
	resp1b, _ := server.respond(req).(requestapi.BestBlockHeightResponse)
	assert.Equal(t, "", resp1b.ResponseError())
	assert.Equal(t, int32(10), resp1b.BlockHeight)

	req.SetRequestType(requestapi.SERVER_LATEST_ATTESTATION)
	resp2, _ := server.respond(req).(requestapi.LatestAttestationResponse)
	assert.Equal(t, "", resp2.ResponseError())
	assert.Equal(t, txid.String(), resp2.TxHash)

	reqVerify := requestapi.ServerVerifyBlockRequest{}
	reqVerify.SetRequestType(requestapi.SERVER_VERIFY_BLOCK)
	reqVerify.Id = "1"
	resp3, _ := server.respond(reqVerify).(requestapi.VerifyBlockResponse)
	assert.Equal(t, "", resp3.ResponseError())
	assert.Equal(t, true, resp3.Attested)

	reqVerify.Id = "11"
	resp4, _ := server.respond(reqVerify).(requestapi.VerifyBlockResponse)
	assert.Equal(t, "", resp4.ResponseError())
	assert.Equal(t, false, resp4.Attested)

	req.SetRequestType("WhenMoon")
	resp5, _ := server.respond(req).(requestapi.Response)
	assert.Equal(t, "**Server** Non supported request type WhenMoon", resp5.ResponseError())

	// Test SERVER_VERIFY_BLOCK request for the attested best block and a new generated block
	sideClientFake.Generate(1)
	bestblockhashnew, _ := server.sideClient.GetBestBlockHash()
	server.latestCommitment = *bestblockhashnew
	assert.Equal(t, *bestblockhashnew, server.latestCommitment)

	reqVerify.Id = bestblockhash.String()
	resp6, _ := server.respond(reqVerify).(requestapi.VerifyBlockResponse)
	assert.Equal(t, "", resp6.ResponseError())
	assert.Equal(t, true, resp6.Attested)

	reqVerify.Id = bestblockhashnew.String()
	resp7, _ := server.respond(reqVerify).(requestapi.VerifyBlockResponse)
	assert.Equal(t, "", resp7.ResponseError())
	assert.Equal(t, false, resp7.Attested)
}

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

	reqChannel := server.RequestServiceChannel()
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

	// Test requesting SERVER_BEST_BLOCK through RequestServiceChannel
	reqBest := requestapi.BaseRequest{}
	reqBest.SetRequestType(requestapi.SERVER_BEST_BLOCK)
	reqChannel <- requestapi.RequestWithResponseChannel{reqBest, responseChan}
	responseBest := (<-responseChan).(requestapi.BestBlockResponse)
	assert.Equal(t, "", responseBest.ResponseError())
	assert.Equal(t, bestblockhash.String(), responseBest.BlockHash)

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
