package requestapi

import (
	"bytes"
	"encoding/json"
	"mainstay/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Request server handling Test
func TestRequestService(t *testing.T) {
	channel := models.NewChannel()

	go func() { // Wait for a best block request and reply with genesis
		req := <-channel.Requests
		res := models.BestBlockResponse{models.Response{req, ""}, "357abd41543a09f9290ff4b4ae008e317f252b80c96492bd9f346cced0943a7f"}
		channel.Responses <- res
	}()

	router := NewRouter(channel)

	r, err := http.NewRequest("GET", "/api/bestblock/", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	resp := bytes.NewReader(w.Body.Bytes())
	dec := json.NewDecoder(resp)
	var decResp map[string]interface{}
	errr := dec.Decode(&decResp)
	if errr != nil {
		t.Fatal(err)
	}

	// Verify the correct response was sent to the channel upon handling the request
	if decResp["blockhash"] != "357abd41543a09f9290ff4b4ae008e317f252b80c96492bd9f346cced0943a7f" {
		t.Fatal("Incorrect best block hash")
	}
}
