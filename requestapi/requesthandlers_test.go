package requestapi

import (
	"bytes"
	"encoding/json"
	"mainstay/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleBlock(t *testing.T) {
	println("\x1B[33mNo Write Test, Just Declare\x1B[0m")
}

func TestHandleBestBlock(t *testing.T) {
	println("\x1B[33mNo Write Test, Just Declare\x1B[0m")
}

func TestHandleBestBlockHeight(t *testing.T) {
	println("\x1B[33mNo Write Test, Just Declare\x1B[0m")
}

func TestHandleCommitmentSend(t *testing.T) {
	channel := models.NewChannel()
	go func() {
		<-channel.RequestPost
		response := models.CommintmentSendResponce{models.Response{""}, true}
		channel.Responses <- response
	}()
	router := NewRouter(channel)
	request, err := http.NewRequest(POST, ROUTE_COMMITMENT_SEND, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Add("CLIENT-ID", "Fuck 1")
	request.Header.Add("HASH", "Fuck 2")
	request.Header.Add("HEIGHT", "Fuck 3")
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)
	resp := bytes.NewReader(writer.Body.Bytes())
	dec := json.NewDecoder(resp)
	var decResp map[string]interface{}
	if dec.Decode(&decResp) != nil {
		t.Fatal(err)
	}
}

func TestHandleIndex(t *testing.T) {
	println("\x1B[33mNo Write Test, Just Declare\x1B[0m")
}

func TestHandleLatestAttestation(t *testing.T) {
	println("\x1B[33mNo Write Test, Just Declare\x1B[0m")
}

func TestHandleServerVerify(t *testing.T) {
	println("\x1B[33mNo Write Test, Just Declare\x1B[0m")
}

func TestHandleTransaction(t *testing.T) {
	println("\x1B[33mNo Write Test, Just Declare\x1B[0m")
}
