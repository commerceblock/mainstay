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
	tx0          string
	pk0          string
	script       string
	isRegtest    bool
	sub          *messengers.SubscriberZmq
	pub          *messengers.PublisherZmq
	attestedHash chainhash.Hash
	nextHash     chainhash.Hash
	poller       *zmq.Poller
	client       *attestation.AttestClient
)

// main conf path for main use in attestation
const CONF_PATH = "/src/mainstay/cmd/txsigningtool/conf.json"

// demo parameters for use with regtest demo
const DEMO_CONF_PATH = "/src/mainstay/cmd/txsigningtool/demo-conf.json"
const DEMO_INIT_PATH = "/src/mainstay/cmd/txsigningtool/demo-init-signingtool.sh"

func parseFlags() {
	flag.BoolVar(&isRegtest, "regtest", false, "Use regtest wallet configuration")
	flag.StringVar(&tx0, "tx", "", "Tx id for genesis attestation transaction")
	flag.StringVar(&pk0, "pk", "", "Client pk for genesis attestation transaction")
	flag.StringVar(&script, "script", "", "Redeem script in case multisig is used")
	flag.Parse()

	if (tx0 == "" || pk0 == "") && !isRegtest {
		flag.PrintDefaults()
		log.Fatalf("Need to provide both -tx and -pk argument. To use test configuration set the -regtest flag.")
	}
}

func init() {
	parseFlags()
	var config *confpkg.Config

	if isRegtest {
		// btc regtest node setup
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
		script = test.SCRIPT
	} else {
		confFile, confErr := confpkg.GetConfFile(os.Getenv("GOPATH") + CONF_PATH)
		if confErr != nil {
			log.Fatal(confErr)
		}
		var configErr error
		config, configErr = confpkg.NewConfig(confFile)
		if configErr != nil {
			log.Fatal(configErr)
		}
	}

	config.SetInitTX(tx0)
	config.SetInitPK(pk0)
	config.SetMultisigScript(script)
	client = attestation.NewAttestClient(config, true) // isSigner flag set

	// comms setup
	poller = zmq.NewPoller()
	topics := []string{confpkg.TOPIC_NEW_HASH, confpkg.TOPIC_NEW_TX, confpkg.TOPIC_CONFIRMED_HASH}
	sub = messengers.NewSubscriberZmq(fmt.Sprintf("127.0.0.1:%d", confpkg.MAIN_PUBLISHER_PORT), topics, poller)
	pub = messengers.NewPublisherZmq(5001, poller)
}

func main() {
	for {
		sockets, _ := poller.Poll(-1)
		for _, socket := range sockets {
			if sub.Socket() == socket.Socket {
				topic, msg := sub.ReadMessage()
				switch topic {
				case confpkg.TOPIC_NEW_TX:
					processTx(msg)
				case confpkg.TOPIC_NEW_HASH:
					nextHash = processHash(msg)
					fmt.Printf("nexthash %s\n", nextHash.String())
				case confpkg.TOPIC_CONFIRMED_HASH:
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

	scriptSig := signedMsgTx.TxIn[0].SignatureScript
	if len(scriptSig) > 0 {
		sigs, _ := crypto.ParseScriptSig(scriptSig)
		fmt.Printf("sending sig %s\n", hex.EncodeToString(sigs[0]))
		pub.SendMessage(sigs[0], confpkg.TOPIC_SIGS)
	}
}
