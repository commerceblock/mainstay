// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"mainstay/log"
	"io/ioutil"
	confpkg "mainstay/config"
	"net/http"
	"strings"
	"github.com/btcsuite/btcd/wire"
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
	SighashString   []string `json:"sighash_string"`
	MerkleRoot      string `json:"merkle_root"`
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
func (f AttestSignerHttp) GetSigs(sigHashes [][]byte, merkle_root string) []wire.TxWitness {

	witness := make([]wire.TxWitness, len(sigHashes)) // init witness

	sigHashesStr := make([]string, len(sigHashes))
	for i := 0; i < len(sigHashes); i++ {
		sigHashesStr[i] = hex.EncodeToString(sigHashes[i])
	} 

	requestBody := &RequestBody{
		SighashString:  sigHashesStr,
		MerkleRoot:      merkle_root,
	}

	// Encode the request body to JSON
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		log.Info("Error marshalling request body: ", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest(http.MethodPost, f.url, bytes.NewReader(requestBodyJSON))
	if err != nil {
		log.Info("Error creating request: ", err)
	}

	// Set the request headers
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		log.Info("Error sending request: ", err)
	}

	// Close the response body
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Info("Error reading response body: ", err)
	}

	var data map[string]interface{}
	jsonErr := json.Unmarshal([]byte(body), &data)
	if jsonErr != nil {
		log.Info("Error unmarshaling JSON:", err)
	}

	for i := 0; i < len(sigHashesStr); i++ {
		witnessStr := data["witness"].([]interface{})[i].(string)
		witnessData := strings.Split(witnessStr, " ")
		sig, _ := hex.DecodeString(witnessData[0])
		sigBytes := append(sig, []byte{byte(1)}...)
		pubkey, _ := hex.DecodeString(witnessData[1])
		witness[i] = wire.TxWitness{sigBytes, pubkey}
	}

	return witness
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
