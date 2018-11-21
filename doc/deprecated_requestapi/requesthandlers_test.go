package requestapi

// import (
// 	"bytes"
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// )

// func TestHandleIndex(t *testing.T) {
// 	println("\x1B[33mNo Write Test, Just Declare\x1B[0m")
// }

// func TestHandleVerifyBlock(t *testing.T) {
// 	println("\x1B[33mNo Write Test, Just Declare\x1B[0m")
// }

// func TestHandleBestBlock(t *testing.T) {
// 	println("\x1B[33mNo Write Test, Just Declare\x1B[0m")
// }

// func TestHandleBestBlockHeight(t *testing.T) {
// 	println("\x1B[33mNo Write Test, Just Declare\x1B[0m")
// }

// func TestHandleCommitmentSend(t *testing.T) {
// 	channel := NewChannel()
// 	go func() {
// 		tmp := <-channel.Requests
// 		// TODO: Add real test
// 		commitmentReq := tmp.(ServerCommitmentSendRequest)
// 		_ = commitmentReq.ClientId
// 		_ = commitmentReq.Hash
// 		_ = commitmentReq.Height
// 		response := CommitmentSendResponse{Verified: true}
// 		channel.Responses <- response
// 	}()
// 	router := NewRouter(channel)
// 	request, err := http.NewRequest(POST, ROUTE_COMMITMENT_SEND, nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	request.Header.Add("CLIENT-ID", "123456789")
// 	request.Header.Add("HASH", "123456789")
// 	request.Header.Add("HEIGHT", "123456789")
// 	writer := httptest.NewRecorder()
// 	router.ServeHTTP(writer, request)
// 	resp := bytes.NewReader(writer.Body.Bytes())
// 	dec := json.NewDecoder(resp)
// 	var decResp map[string]interface{}
// 	if dec.Decode(&decResp) != nil {
// 		t.Fatal(err)
// 	}
// 	if decResp["verified"] != true {
// 		t.Fatal("Incorrect Commitment Send")
// 	}
// }

// func TestHandleLatestAttestation(t *testing.T) {
// 	println("\x1B[33mNo Write Test, Just Declare\x1B[0m")
// }
