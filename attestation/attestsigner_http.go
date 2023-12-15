// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	confpkg "mainstay/config"
	"mainstay/crypto"
	"net/http"
)

// AttestSignerFake struct
//
// Implements AttestSigner interface and provides
// mock functionality for receiving sigs from signers
type AttestSignerHttp struct {
	client http.Client
	url    string
}

type RequestBody struct {
	TxHex           string `json:"tx_hex"`
	Value           int    `json:"value"`
	MerkleRoot      string `json:"merkle_root"`
	RedeemScriptHex string `json:"redeem_script_hex"`
}

// store latest hash and transaction
var signerTxPreImageBytes []byte
var signerConfirmedHashBytes []byte

// Return new AttestSignerFake instance
func NewAttestSignerHttp(config confpkg.SignerConfig) AttestSignerHttp {
	return AttestSignerHttp{
		client: http.Client{},
		url:    config.Url,
	}
}

// Resubscribe - do nothing
func (f AttestSignerHttp) ReSubscribe() {
	return
}

// Store received confirmed hash
func (f AttestSignerHttp) SendConfirmedHash(hash []byte) {
	signerConfirmedHashBytes = hash
}

// Store received new tx
func (f AttestSignerHttp) SendTxPreImages(txs [][]byte) {
	signerTxPreImageBytes = SerializeBytes(txs)
}

// Return signatures for received tx and hashes
func (f AttestSignerHttp) GetSigs(txHash string, redeem_script string, merkle_root string) [][]crypto.Sig {
	// get unserialized tx pre images
	txPreImages := UnserializeBytes(signerTxPreImageBytes)

	sigs := make([][]crypto.Sig, len(txPreImages)) // init sigs

	// value hardcoded for now, needs a fix
	requestBody := &RequestBody{
		TxHex:           txHash,
		Value:           10000,
		MerkleRoot:      merkle_root,
		RedeemScriptHex: redeem_script,
	}

	// Encode the request body to JSON
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Error marshalling request body:", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest(http.MethodPost, f.url, bytes.NewReader(requestBodyJSON))
	if err != nil {
		fmt.Println("Error creating request:", err)
	}

	// Set the request headers
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
	}

	// Close the response body
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
	}

	// Print the response
	fmt.Println(string(body))
	sig, _ := hex.DecodeString(string(body))
	sigs[0][0] = sig
	return sigs
}

// Transform received list of bytes into a single byte
// slice with format: [len bytes0] [bytes0] [len bytes1] [bytes1]
func SerializeBytes(data [][]byte) []byte {

	// empty case return nothing
	if len(data) == 0 {
		return []byte{}
	}

	var serializedBytes []byte

	// iterate through each byte slice adding
	// length and data bytes to bytes slice
	for _, dataX := range data {
		serializedBytes = append(serializedBytes, byte(len(dataX)))
		serializedBytes = append(serializedBytes, dataX...)
	}

	return serializedBytes
}

// Transform single byte slice (result of SerializeBytes)
// into a list of byte slices excluding lengths
func UnserializeBytes(data []byte) [][]byte {

	// empty case return nothing
	if len(data) == 0 {
		return [][]byte{}
	}

	var dataList [][]byte

	// process data slice
	it := 0
	for it < len(data) {
		// get next data by reading byte size
		txSize := data[it]

		// check if next size excees the bounds and break
		// maybe TODO: error handling
		if (int(txSize) + 1 + it) > len(data) {
			break
		}

		dataX := append([]byte{}, data[it+1:it+1+int(txSize)]...)
		dataList = append(dataList, dataX)

		it += 1 + int(txSize)
	}

	return dataList
}
