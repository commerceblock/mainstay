// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"mainstay/attestation"
	confpkg "mainstay/config"
	_ "mainstay/crypto"
	"mainstay/messengers"
	"mainstay/test"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	zmq "github.com/pebbe/zmq4"
)

// The transaction signing tool is used by members of the multisig script
// used to generate new attestations transactions. This process communicates
// with the main attestation service to receive latest commitments and sign transactions

var (
	// use attest client interface for signing
	client    *attestation.AttestClient
	isRegtest bool

	// init transaction parameters
	pk0         string
	script0     string
	chaincodes0 string

	// topup parameters
	addrTopup   string
	pkTopup     string
	scriptTopup string

	// communication with attest service
	sub      *messengers.SubscriberZmq
	pub      *messengers.PublisherZmq
	poller   *zmq.Poller
	host     string
	hostMain string

	attestedHash chainhash.Hash // previous attested hash
	nextHash     chainhash.Hash // next hash to sign with
)

// main conf path for main use in attestation
const ConfPath = "/src/mainstay/cmd/txsigningtool/conf.json"

// demo parameters for use with regtest demo
const DemoConfPath = "/src/mainstay/cmd/txsigningtool/demo-conf.json"
const DemoInitPath = "/src/mainstay/cmd/txsigningtool/demo-init-signingtool.sh"

func parseFlags() {
	flag.BoolVar(&isRegtest, "regtest", false, "Use regtest wallet configuration")
	flag.StringVar(&pk0, "pk", "", "Client pk for genesis attestation transaction")
	flag.StringVar(&script0, "script", "", "Redeem script in case multisig is used")
	flag.StringVar(&addrTopup, "addrTopup", "", "Address for topup transaction")
	flag.StringVar(&pkTopup, "pkTopup", "", "Client pk for topup address")
	flag.StringVar(&scriptTopup, "scriptTopup", "", "Redeem script for topup")

	flag.StringVar(&host, "host", "*:5002", "Client host to publish signatures at")
	hostMainDefault := fmt.Sprintf("127.0.0.1:%d", attestation.DefaultMainPublisherPort)
	flag.StringVar(&hostMain, "hostMain", hostMainDefault, "Mainstay host for signer to subscribe to")
	flag.Parse()

	if pk0 == "" && !isRegtest {
		flag.PrintDefaults()
		log.Fatalf("Need to provide -pk argument. To use test configuration set the -regtest flag.")
	}
}

func init() {
	parseFlags()
	var config *confpkg.Config

	// regtest mode
	// run demo init script to setup bitcoin node and initial transaction
	// generate test config using demo config file
	if isRegtest {
		cmd := exec.Command("/bin/sh", os.Getenv("GOPATH")+DemoInitPath)
		_, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
		}

		confFile, confErr := confpkg.GetConfFile(os.Getenv("GOPATH") + DemoConfPath)
		if confErr != nil {
			log.Fatal(confErr)
		}
		var configErr error
		config, configErr = confpkg.NewConfig(confFile)
		if configErr != nil {
			log.Fatal(configErr)
		}
		pk0 = test.PrivMain
		script0 = test.Script
		pkTopup = test.TopupPrivMain
		scriptTopup = test.TopupScript
		addrTopup = test.TopupAddress
		chaincodes0 = test.InitChaincodes
	} else {
		// regular mode
		// use conf file to setup config
		confFile, confErr := confpkg.GetConfFile(os.Getenv("GOPATH") + ConfPath)
		if confErr != nil {
			log.Fatal(confErr)
		}
		var configErr error
		config, configErr = confpkg.NewConfig(confFile)
		if configErr != nil {
			log.Fatal(configErr)
		}

		// if init script is not set throw error
		if script0 == "" && config.InitScript() == "" {
			flag.PrintDefaults()
			log.Fatalf(`Need to provide -script argument.
                To use test configuration set the -regtest flag.`)
		}
	}

	// overwrite init config if set from command line
	if pk0 != "" {
		config.SetInitPK(pk0)
	}
	if script0 != "" {
		config.SetInitScript(script0)
	}
	if chaincodes0 != "" {
		config.SetInitChaincodes(strings.Split(chaincodes0, ","))
	}

	// overwrite topup config if set from command line
	if pkTopup != "" {
		config.SetTopupPK(pkTopup)
	}
	if addrTopup != "" && scriptTopup != "" {
		config.SetTopupAddress(addrTopup)
		config.SetTopupScript(scriptTopup)
	}

	// init client interface with isSigner flag set
	client = attestation.NewAttestClient(config, true)

	// comms setup
	poller = zmq.NewPoller()
	topics := []string{attestation.TopicNewTx, attestation.TopicConfirmedHash}
	sub = messengers.NewSubscriberZmq(hostMain, topics, poller)
	pub = messengers.NewPublisherZmq(host, poller)
}

func main() {
	// delay to resubscribe
	resubscribeDelay := 5 * time.Minute
	timer := time.NewTimer(resubscribeDelay)
	for {
		select {
		case <-timer.C:
			log.Println("resubscribing to mainstay...")
			// remove socket and close
			sub.Close(poller)
			// re-assign subscriber socket
			topics := []string{attestation.TopicNewTx, attestation.TopicConfirmedHash}
			sub = messengers.NewSubscriberZmq(hostMain, topics, poller)
			timer = time.NewTimer(resubscribeDelay)
		default:
			sockets, _ := poller.Poll(-1)
			for _, socket := range sockets {
				if sub.Socket() == socket.Socket {
					topic, msg := sub.ReadMessage()
					switch topic {
					case attestation.TopicNewTx:
						processTx(msg)
					case attestation.TopicConfirmedHash:
						attestedHash = processHash(msg)
						log.Printf("attestedhash %s\n", attestedHash.String())
					}
				}
			}
			time.Sleep(1 * time.Second)
		}
	}
}

// Get hash from received message
func processHash(msg []byte) chainhash.Hash {
	hash, hashErr := chainhash.NewHash(msg)
	if hashErr != nil {
		log.Fatal(hashErr)
	}
	return *hash
}

// Process received tx, verify and reply with signature
func processTx(msg []byte) {

	var sigs [][]byte

	// get tx pre images from message
	txPreImages := attestation.UnserializeBytes(msg)

	// process each pre image transaction and sign
	for txIt, txPreImage := range txPreImages {
		// add hash type to tx serialization
		txPreImage = append(txPreImage, []byte{1, 0, 0, 0}...)
		txPreImageHash := chainhash.DoubleHashH(txPreImage)

		// sign first tx with tweaked priv key and
		// any remaining txs with topup key
		var sig *btcec.Signature
		var signErr error
		if txIt == 0 {
			priv := client.GetKeyFromHash(attestedHash).PrivKey
			sig, signErr = priv.Sign(txPreImageHash.CloneBytes())
		} else {
			sig, signErr = client.WalletPrivTopup.PrivKey.Sign(txPreImageHash.CloneBytes())
		}
		if signErr != nil {
			log.Fatalf("%v\n", signErr)
		}

		// add hash type to signature as well
		sigBytes := append(sig.Serialize(), []byte{byte(1)}...)

		log.Printf("sending sig(%d) %s\n", txIt, hex.EncodeToString(sigBytes))

		sigs = append(sigs, sigBytes)
	}

	serializedSigs := attestation.SerializeBytes(sigs)
	pub.SendMessage(serializedSigs, attestation.TopicSigs)
}
