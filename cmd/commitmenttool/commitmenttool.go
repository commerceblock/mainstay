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
	"net/http"
	"os"
	"strings"
	"time"

	"mainstay/config"
	"mainstay/log"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// consts
const (
	DefaultApiHost       = "https://mainstay.xyz"    // testnet mainstay url
	ApiCommitmentSendUrl = "/api/v1/commitment/send" // url to send commitments to

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

	position  int    // client position
	authtoken string // client authorisation token
	privkey   string // client private key
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
	flag.StringVar(&privkey, "privkey", "", "Client private key for signing (optional)")
	flag.Parse()
}

// Init mode
// Generate new ECDSA priv-pub key pair for the client to use
// when signing new commitments and sending to Mainstay API
func doInitMode() {
	log.Infoln("****************************")
	log.Infoln("****** Init mode ***********")
	log.Infoln("****************************")

	log.Infof("Generating new key...\n")
	newPriv, newPrivErr := btcec.NewPrivateKey(btcec.S256())
	if newPrivErr != nil {
		log.Error(newPrivErr)
	}

	newPrivBytesStr := hex.EncodeToString(newPriv.Serialize())
	log.Infof("generated priv: %s\n", newPrivBytesStr)
	newPubBytesStr := hex.EncodeToString(newPriv.PubKey().SerializeCompressed())
	log.Infof("generated pub: %s\n", newPubBytesStr)

	log.Infof("The private key should be used for signing future client commitments\n")
	log.Infof("The public key should be provided when posting these to Mainstay API\n")
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

	log.Infof("Response status: %s", resp.Status)

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
		log.Errorf("Key ('%s') decode error: %v\n", privkey, decodeErr)
	}
	privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privkeyBytes)

	// sign message
	sig, signErr := privKey.Sign(msg)
	if signErr != nil {
		log.Errorf("Signing error: %v\n", signErr)
	}
	return sig.Serialize()
}

// Ocean mode
// Recurrent commitments of Ocean blockhash to Mainstay API
// At regular intervals, fetch commitment, sign and send
func doOceanMode() {
	log.Infoln("****************************")
	log.Infoln("****** Ocean mode **********")
	log.Infoln("****************************")

	// check priv key is set
	if privkey == "" {
		log.Infoln("no private key provided")
	}

	// get conf file
	confFile, confErr := config.GetConfFile(os.Getenv("GOPATH") + ConfPath)
	if confErr != nil {
		log.Error(confErr)
	}

	// get ocean sidechain client from config
	client := config.NewClientFromConfig(ClientChainName, false, confFile)

	sleepTime := 0 * time.Second // start immediately
	for {
		timer := time.NewTimer(sleepTime)
		select {
		case <-timer.C:
			log.Infoln("Fetching next blockhash commitment...")

			// get next blockhash
			blockhash, blockhashErr := client.GetBestBlockHash()
			if blockhashErr != nil {
				log.Errorf("Client fetching error: %v\n", blockhashErr)
			}
			log.Infoln("Commitment: ", blockhash.String())

			// get reverse blockhash bytes as this is how blockhashes are displayed
			revBlockHashBytes, _ := hex.DecodeString(blockhash.String())

			// sign commitment
			var sigBytes []byte
			if privkey != "" {
				sigBytes = sign(revBlockHashBytes)
			}

			// send signed commitment
			sendErr := send(sigBytes, hex.EncodeToString(revBlockHashBytes))
			if sendErr != nil {
				log.Errorf("Commitment send error: %v\n", sendErr)
			} else {
				log.Infoln("Success!")
			}

			sleepTime = time.Duration(delay) * time.Minute
			log.Infof("********** sleeping for: %s ...\n", sleepTime.String())
		}
	}
}

// Standard mode
// One time commitment to the Mainstay API
// Sign the commitment provided and POST to API
func doStandardMode() {
	log.Infoln("****************************")
	log.Infoln("****** Commitment mode *****")
	log.Infoln("****************************")

	log.Infoln()
	log.Info("Insert commitment: ")
	var commitment string
	fmt.Scanln(&commitment)

	// try commitment decoding
	commitmentBytes, decodeErr := hex.DecodeString(commitment)
	if decodeErr != nil {
		log.Errorf("Commitment ('%s') decode error: %v\n", commitment, decodeErr)
	}
	_, hashErr := chainhash.NewHash(commitmentBytes)
	if hashErr != nil {
		log.Errorf("Commitment ('%s') to hash error: %v\n", commitment, hashErr)
	}

	log.Infoln()
	log.Info("Sign commitment, send commitment or both? ")
	var whatToDo string
	fmt.Scanln(&whatToDo)

	var sigBytes []byte
	if strings.ToLower(whatToDo) == "send" {
		log.Infoln()
		log.Info("Insert signature: ")
		var signature string
		fmt.Scanln(&signature)
		if signature == "" {
			log.Infoln("no signature provided")
		} else {
			var sigBytesErr error
			sigBytes, sigBytesErr = b64.StdEncoding.DecodeString(signature)
			if sigBytesErr != nil {
				log.Errorf("Signature (%s) decoding error: %v\n", signature, sigBytesErr)
			}
		}
	} else if strings.ToLower(whatToDo) == "sign" || strings.ToLower(whatToDo) == "both" {
		log.Infoln()
		log.Info("Insert private key: ")
		fmt.Scanln(&privkey)
		if privkey == "" {
			log.Infoln("no private key provided")
		} else {
			sigBytes = sign(commitmentBytes)
			log.Infoln()
			log.Infoln("Signature: " + b64.StdEncoding.EncodeToString(sigBytes))
		}
	} else {
		log.Error("Invalid option")
	}

	if strings.ToLower(whatToDo) == "send" || strings.ToLower(whatToDo) == "both" {
		// ask for position and auth token
		log.Infoln()
		log.Info("Insert position: ")
		fmt.Scan(&position)

		log.Infoln()
		log.Info("Insert auth token: ")
		fmt.Scan(&authtoken)

		// send signed commitment
		sendErr := send(sigBytes, commitment)
		if sendErr != nil {
			log.Errorf("Commitment send error: %v\n", sendErr)
		}
		log.Infoln("Success!")
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
}
