// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

// Commitment tool

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"mainstay/config"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// consts
const (
	DefaultApiHost       = "https://testnet.mainstay.xyz" // testnet mainstay url
	ApiCommitmentSendUrl = "/api/v1/commitment/send"      // url to send commitments to

	// config for sidechain connectivity (optional)
	ClientChainName = "ocean"
	ConfPath        = "/src/mainstay/cmd/commitmenttool/conf.json"
)

// vars
var (
	apiHost string // mainstay host
	isInit  bool   // init flag
	isOcean bool   // ocean flag
	delay   int    // commitment delay

	position   int    // client position
	authtoken  string // client authorisation token
	privkey    string // client private key
	signature  string // client signature for commitment
	commitment string // client commitment
)

// init
func init() {
	// basic configurations
	flag.StringVar(&apiHost, "apiHost", DefaultApiHost, "Host address for mainstay API")

	// mode options
	flag.BoolVar(&isInit, "init", false, "Init mode")
	flag.BoolVar(&isOcean, "ocean", false, "Ocean mode")
	flag.IntVar(&delay, "delay", 60, "Delay in minutes between commitments")

	// commitment variables
	flag.IntVar(&position, "position", -1, "Client merkle commitment position")
	flag.StringVar(&authtoken, "authtoken", "", "Client authorization token")
	flag.StringVar(&privkey, "privkey", "", "Client private key for signing")
	flag.StringVar(&signature, "signature", "", "Client signature for commitment")
	flag.StringVar(&commitment, "commitment", "", "Client commitment to sign and send")
	flag.Parse()
}

// Init mode
// Generate new ECDSA priv-pub key pair for the client to use
// when signing new commitments and sending to Mainstay API
func doInitMode() {
	log.Printf("Init mode\n")

	log.Printf("Generating new key...\n")
	newPriv, newPrivErr := btcec.NewPrivateKey(btcec.S256())
	if newPrivErr != nil {
		log.Fatal(newPrivErr)
	}

	newPrivBytesStr := hex.EncodeToString(newPriv.Serialize())
	log.Printf("generated priv: %s\n", newPrivBytesStr)
	newPubBytesStr := hex.EncodeToString(newPriv.PubKey().SerializeCompressed())
	log.Printf("generated pub: %s\n", newPubBytesStr)

	log.Printf("The private key should be used for signing future client commitments\n")
	log.Printf("The public key should be provided when posting these to Mainstay API\n")
}

// Send commitment and signature to Mainstay API
// Request requires providing pubkey and authtoken
//
// data sent:
// - pubkey (serialized hex format compressed or uncompressed)
// - authtoken (authorization token generated on signup)
// - msg (32 byte hash commitment in hex encoded string)
// - signature (ECDSA signature encoded to base64)
func send(sig []byte, msg string) error {

	// construct payload and signature and bring to base64 format
	payload := fmt.Sprintf("{\"commitment\": \"%s\", \"position\": %d, \"token\": \"%s\"}",
		msg, position, authtoken)
	payload64 := b64.StdEncoding.EncodeToString([]byte(payload))
	sig64 := b64.StdEncoding.EncodeToString(sig)
	var chunk = fmt.Sprintf("{\"X-MAINSTAY-PAYLOAD\": \"%s\", \"X-MAINSTAY-SIGNATURE\": \"%s\"}",
		payload64, sig64)

	// send post request along with chunk as body
	url := fmt.Sprintf("%s%s", apiHost, ApiCommitmentSendUrl)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(chunk)))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	log.Println("response Status:", resp.Status)

	// check status response
	if resp.StatusCode == 200 {
		dec := json.NewDecoder(resp.Body)
		var respJson map[string]interface{}
		decErr := dec.Decode(&respJson)
		if decErr != nil {
			return decErr
		}
		if val, ok := respJson["error"]; ok {
			return errors.New(val.(string))
		}

		return nil
	}

	return errors.New(fmt.Sprintf("Response status %s", resp.Status))
}

// Decode private key and get btcec ECDSA key
// Sign received byte message with private key
func sign(msg []byte) []byte {
	// try key decoding
	privkeyBytes, decodeErr := hex.DecodeString(privkey)
	if decodeErr != nil {
		log.Fatal(fmt.Sprintf("Key ('%s') decode error: %v\n", privkey, decodeErr))
	}
	privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privkeyBytes)

	// sign message
	sig, signErr := privKey.Sign(msg)
	if signErr != nil {
		log.Fatal(fmt.Sprintf("Signing error: %v\n", signErr))
	}
	return sig.Serialize()
}

// Ocean mode
// Recurrent commitments of Ocean blockhash to Mainstay API
// At regular intervals, fetch commitment, sign and send
func doOceanMode() {
	log.Printf("Ocean mode\n")

	// check priv key is set
	if privkey == "" {
		log.Fatal("Need to provide -privkey.")
	}

	// get conf file
	confFile, confErr := config.GetConfFile(os.Getenv("GOPATH") + ConfPath)
	if confErr != nil {
		log.Fatal(confErr)
	}

	// get ocean sidechain client from config
	client := config.NewClientFromConfig(ClientChainName, false, confFile)

	sleepTime := 0 * time.Second // start immediately
	for {
		timer := time.NewTimer(sleepTime)
		select {
		case <-timer.C:
			log.Println("Fetching next blockhash commitment...")

			// get next blockhash
			blockhash, blockhashErr := client.GetBestBlockHash()
			if blockhashErr != nil {
				log.Fatal(fmt.Sprintf("Client fetching error: %v\n", blockhashErr))
			}
			log.Println("Commitment: ", blockhash.String())

			// get reverse blockhash bytes as this is how blockhashes are displayed
			revBlockHashBytes, _ := hex.DecodeString(blockhash.String())

			// sign commitment
			sigBytes := sign(revBlockHashBytes)

			// send signed commitment
			sendErr := send(sigBytes, hex.EncodeToString(revBlockHashBytes))
			if sendErr != nil {
				log.Fatal(fmt.Sprintf("Commitment send error: %v\n", sendErr))
			} else {
				log.Println("Success!")
			}

			sleepTime = time.Duration(delay) * time.Minute
			log.Printf("********** sleeping for: %s ...\n", sleepTime.String())
		}
	}
}

// Standard mode
// One time commitment to the Mainstay API
// Sign the commitment provided and POST to API
func doStandardMode() {
	log.Printf("Standard mode\n")

	// try commitment decoding
	commitmentBytes, decodeErr := hex.DecodeString(commitment)
	if decodeErr != nil {
		log.Fatal(fmt.Sprintf("Commitment ('%s') decode error: %v\n", commitment, decodeErr))
	}
	hash, hashErr := chainhash.NewHash(commitmentBytes)
	if hashErr != nil {
		log.Fatal(fmt.Sprintf("Commitment ('%s') to hash error: %v\n", commitment, hashErr))
	}

	if signature == "" && privkey == "" {
		log.Fatal("Need to provide either -signature or -privkey.")
	}

	// get commitment signature
	var sigBytes []byte
	if signature != "" {
		var sigBytesErr error
		sigBytes, sigBytesErr = b64.StdEncoding.DecodeString(signature)
		if sigBytesErr != nil {
			log.Fatal(fmt.Sprintf("Signature (%s) decoding error: %v\n", signature, sigBytesErr))
		}
	} else {
		sigBytes = sign(hash.CloneBytes())
	}

	// send signed commitment
	sendErr := send(sigBytes, commitment)
	if sendErr != nil {
		log.Fatal(fmt.Sprintf("Commitment send error: %v\n", sendErr))
	} else {
		log.Println("Success!")
	}
}

// main
func main() {
	// choose mode to run on based on input parameters
	if isInit {
		doInitMode()
	} else if isOcean {
		doOceanMode()
	} else {
		doStandardMode()
	}
	log.Printf("Finishing...\n")
}
