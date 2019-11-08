// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

// Staychain confirmation tool

import (
	"flag"
	"os"
	"strings"

	"mainstay/clients"
	"mainstay/config"
	"mainstay/staychain"
	"mainstay/log"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Use staychain package to read attestations, verify and print information

const ClientChainName = "clientchain"
const ConfPath = "/src/mainstay/cmd/confirmationtool/conf.json"
const DefaultApiHost = "http://localhost:80" // to replace with actual mainstay url

var (
	tx          string
	script      string
	chaincodes  string
	apiHost     string
	position    int
	showDetails bool
	mainConfig  *config.Config
	client      clients.SidechainClient
)

// init
func init() {
	flag.BoolVar(&showDetails, "detailed", false, "Detailed information on attestation transaction")
	flag.StringVar(&tx, "tx", "", "Tx id from which to start searching the staychain")
	flag.StringVar(&script, "script", "", "Redeem script of multisig used by attestaton service")
	flag.StringVar(&chaincodes, "chaincodes", "", "Chaincodes for multisig pubkeys")
	flag.StringVar(&apiHost, "apiHost", DefaultApiHost, "Host address for mainstay API")
	flag.IntVar(&position, "position", -1, "Client merkle commitment position")
	flag.Parse()

	if tx == "" || script == "" || position == -1 || chaincodes == "" {
		flag.PrintDefaults()
		log.Errorf("Need to provide all -tx, -script, -chaincodes and -position argument.")
	}

	confFile, confErr := config.GetConfFile(os.Getenv("GOPATH") + ConfPath)
	if confErr != nil {
		log.Error(confErr)
	}
	var mainConfigErr error
	mainConfig, mainConfigErr = config.NewConfig(confFile)
	if mainConfigErr != nil {
		log.Error(mainConfigErr)
	}
	client = config.NewClientFromConfig(ClientChainName, false, confFile)
}

// main method
func main() {
	defer mainConfig.MainClient().Shutdown()
	defer client.Close()

	txraw := getRawTxFromHash(tx)
	fetcher := staychain.NewChainFetcher(mainConfig.MainClient(), txraw)
	chain := staychain.NewChain(fetcher)
	verifier := staychain.NewChainVerifier(mainConfig.MainChainCfg(),
		client, position, script, strings.Split(chaincodes, ","), apiHost)

	// await new attestations and verify
	for transaction := range chain.Updates() {
		log.Infoln("Verifying attestation")
		log.Infof("txid: %s\n", transaction.Txid)
		info, err := verifier.Verify(transaction)
		if err != nil {
			log.Error(err)
		} else {
			printAttestation(transaction, info)
		}
	}
}

// Get raw transaction from a tx string hash using rpc client
func getRawTxFromHash(hashstr string) staychain.Tx {
	txhash, errHash := chainhash.NewHashFromStr(hashstr)
	if errHash != nil {
		log.Warnln("Invalid tx id provided")
		log.Error(errHash)
	}
	txraw, errGet := mainConfig.MainClient().GetRawTransactionVerbose(txhash)
	if errGet != nil {
		log.Warnln("Inititial transaction does not exist")
		log.Error(errGet)
	}
	return staychain.Tx(*txraw)
}

// print attestation information
func printAttestation(tx staychain.Tx, info staychain.ChainVerifierInfo) {
	log.Infoln("Attestation Verified")
	if showDetails {
		log.Infof("%+v\n", tx)
	} else {
		log.Infof("Bitcoin blockhash: %s\n", tx.BlockHash)
	}
	if info != (staychain.ChainVerifierInfo{}) {
		log.Infof("CLIENT blockhash: %s\n", info.Hash().String())
		log.Infof("CLIENT blockheight: %d\n", info.Height())
	}
	log.Infof("\n")
	log.Infof("\n")
	log.Infof("\n")
}
