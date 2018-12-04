// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"mainstay/attestation"
	confpkg "mainstay/config"
	"mainstay/crypto"
	"mainstay/messengers"
	"mainstay/test"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
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
	pk0     string
	script0 string

	// topup parameters
	addrTopup   string
	pkTopup     string
	scriptTopup string

	// communication with attest service
	sub    *messengers.SubscriberZmq
	pub    *messengers.PublisherZmq
	poller *zmq.Poller
	host   string

	attestedHash chainhash.Hash // previous attested hash
	nextHash     chainhash.Hash // next hash to sign with
)

// main conf path for main use in attestation
const CONF_PATH = "/src/mainstay/cmd/txsigningtool/conf.json"

// demo parameters for use with regtest demo
const DEMO_CONF_PATH = "/src/mainstay/cmd/txsigningtool/demo-conf.json"
const DEMO_INIT_PATH = "/src/mainstay/cmd/txsigningtool/demo-init-signingtool.sh"

func parseFlags() {
	flag.BoolVar(&isRegtest, "regtest", false, "Use regtest wallet configuration")
	flag.StringVar(&pk0, "pk", "", "Client pk for genesis attestation transaction")
	flag.StringVar(&script0, "script", "", "Redeem script in case multisig is used")
	flag.StringVar(&addrTopup, "addrTopup", "", "Address for topup transaction")
	flag.StringVar(&pk0, "pkTopup", "", "Client pk for topup address")
	flag.StringVar(&scriptTopup, "scriptTopup", "", "Redeem script for topup")

	flag.StringVar(&host, "host", "*:5001", "Client host to publish signatures at")
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
		cmd := exec.Command("/bin/sh", os.Getenv("GOPATH")+DEMO_INIT_PATH)
		_, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
		}

		confFile, confErr := confpkg.GetConfFile(os.Getenv("GOPATH") + DEMO_CONF_PATH)
		if confErr != nil {
			log.Fatal(confErr)
		}
		var configErr error
		config, configErr = confpkg.NewConfig(confFile)
		if configErr != nil {
			log.Fatal(configErr)
		}
		pk0 = test.PRIV_CLIENT
		script0 = test.SCRIPT
		pkTopup = test.TOPUP_PRIV_CLIENT
		scriptTopup = test.TOPUP_SCRIPT
		addrTopup = test.TOPUP_ADDRESS
	} else {
		// regular mode
		// use conf file to setup config
		confFile, confErr := confpkg.GetConfFile(os.Getenv("GOPATH") + CONF_PATH)
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

	// get publisher addr from config, if set
	publisherAddr := fmt.Sprintf("127.0.0.1:%d", attestation.DEFAULT_MAIN_PUBLISHER_PORT)
	if config.SignerConfig().Publisher != "" {
		publisherAddr = config.SignerConfig().Publisher
	}

	// comms setup
	poller = zmq.NewPoller()
	topics := []string{attestation.TOPIC_NEW_HASH, attestation.TOPIC_NEW_TX, attestation.TOPIC_CONFIRMED_HASH}
	sub = messengers.NewSubscriberZmq(publisherAddr, topics, poller)
	pub = messengers.NewPublisherZmq(host, poller)
}

func main() {
	for {
		sockets, _ := poller.Poll(-1)
		for _, socket := range sockets {
			if sub.Socket() == socket.Socket {
				topic, msg := sub.ReadMessage()
				switch topic {
				case attestation.TOPIC_NEW_TX:
					processTx(msg)
				case attestation.TOPIC_NEW_HASH:
					nextHash = processHash(msg)
					fmt.Printf("nexthash %s\n", nextHash.String())
				case attestation.TOPIC_CONFIRMED_HASH:
					attestedHash = processHash(msg)
					fmt.Printf("attestedhash %s\n", attestedHash.String())
				}
			}
		}
		time.Sleep(1 * time.Second)
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

// Check received tx and verify tx address and client generated address match
func verifyTx(tx wire.MsgTx) bool {
	nextKey, keyErr := client.GetNextAttestationKey(nextHash)
	if keyErr != nil {
		log.Fatal(keyErr)
	}

	nextAddr, _ := client.GetNextAttestationAddr(nextKey, nextHash)

	// exactr addr from unsigned tx and verify addresses match
	_, txScriptAddrs, _, err := txscript.ExtractPkScriptAddrs(tx.TxOut[0].PkScript, client.MainChainCfg)
	if err != nil {
		log.Fatal(err)
	}
	txAddr := txScriptAddrs[0]
	if txAddr.String() == nextAddr.String() {
		fmt.Printf("tx address %s verified\n", txAddr.String())
		return true
	}
	fmt.Printf("tx address %s not verified\n", txAddr.String())
	return false
}

// Process received tx, verify and reply with signature
func processTx(msg []byte) {
	// parse received tx into useful format
	var msgTx wire.MsgTx
	if err := msgTx.Deserialize(bytes.NewReader(msg)); err != nil {
		log.Fatal(err)
	}

	// verify transaction first
	if !verifyTx(msgTx) {
		return
	}

	signedMsgTx, _, signErr := client.SignTransaction(attestedHash, msgTx)
	if signErr != nil {
		log.Fatal(signErr)
	}

	var mySigs []byte
	for i, txin := range signedMsgTx.TxIn {
		scriptSig := txin.SignatureScript
		if len(scriptSig) > 0 {
			sigs, _ := crypto.ParseScriptSig(scriptSig)
			fmt.Printf("sending sig(%d) %s\n", i, hex.EncodeToString(sigs[0]))
			mySigs = append(mySigs, byte(len(sigs[0])))
			mySigs = append(mySigs, sigs[0]...)
		}
	}
	pub.SendMessage(mySigs, attestation.TOPIC_SIGS)
}
