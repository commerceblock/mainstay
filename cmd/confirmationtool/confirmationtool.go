// Staychain confirmation tool
package main

import (
	"flag"
	"log"
	"os"

	"mainstay/clients"
	"mainstay/config"
	"mainstay/staychain"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Use staychain package to read attestations, verify and print information

const CLIENT_CHAIN_NAME = "clientchain"
const CONF_PATH = "/src/mainstay/cmd/confirmationtool/conf.json"
const API_HOST = "http://localhost:8080" // to replace with actual mainstay url

var (
	tx          string
	script      string
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
	flag.IntVar(&position, "position", -1, "Client merkle commitment position")
	flag.Parse()

	if tx == "" || script == "" || position == -1 {
		flag.PrintDefaults()
		log.Fatalf("Need to provide all -tx, -script and -position argument.")
	}

	confFile, confErr := config.GetConfFile(os.Getenv("GOPATH") + CONF_PATH)
	if confErr != nil {
		log.Fatal(confErr)
	}
	var mainConfigErr error
	mainConfig, mainConfigErr = config.NewConfig(confFile)
	if mainConfigErr != nil {
		log.Fatal(mainConfigErr)
	}
	client = config.NewClientFromConfig(CLIENT_CHAIN_NAME, false, confFile)
}

// main method
func main() {
	defer mainConfig.MainClient().Shutdown()
	defer client.Close()

	txraw := getRawTxFromHash(tx)
	fetcher := staychain.NewChainFetcher(mainConfig.MainClient(), txraw)
	chain := staychain.NewChain(fetcher)
	verifier := staychain.NewChainVerifier(mainConfig.MainChainCfg(), client, position, script, API_HOST)

	// await new attestations and verify
	for transaction := range chain.Updates() {
		log.Println("Verifying attestation")
		log.Printf("txid: %s\n", transaction.Txid)
		info, err := verifier.Verify(transaction)
		if err != nil {
			log.Fatal(err)
		} else {
			printAttestation(transaction, info)
		}
	}
}

// Get raw transaction from a tx string hash using rpc client
func getRawTxFromHash(hashstr string) staychain.Tx {
	txhash, errHash := chainhash.NewHashFromStr(hashstr)
	if errHash != nil {
		log.Println("Invalid tx id provided")
		log.Fatal(errHash)
	}
	txraw, errGet := mainConfig.MainClient().GetRawTransactionVerbose(txhash)
	if errGet != nil {
		log.Println("Inititial transcaction does not exist")
		log.Fatal(errGet)
	}
	return staychain.Tx(*txraw)
}

// print attestation information
func printAttestation(tx staychain.Tx, info staychain.ChainVerifierInfo) {
	log.Println("Attestation Verified")
	if showDetails {
		log.Printf("%+v\n", tx)
	} else {
		log.Printf("BITCOIN blockhash: %s\n", tx.BlockHash)
	}
	if info != (staychain.ChainVerifierInfo{}) {
		log.Printf("CLIENT blockhash: %s\n", info.Hash().String())
		log.Printf("CLIENT blockheight: %d\n", info.Height())
	}
	log.Printf("\n")
	log.Printf("\n")
	log.Printf("\n")
}
