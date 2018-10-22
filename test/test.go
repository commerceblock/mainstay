// Package test implements unit test and regtest utitilies for attestation.
package test

import (
	"log"
	"mainstay/clients"
	"mainstay/conf"
	"os"
	"os/exec"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
)

// For regtest attestation demonstration
const DEMO_INIT_PATH = "/src/mainstay/test/demo-init.sh"

// For unit-testing
const TEST_INIT_PATH = "/src/mainstay/test/test-init.sh"

var testConf = []byte(`
{
    "btc": {
        "rpcurl": "localhost:18443",
        "rpcuser": "user",
        "rpcpass": "pass",
        "chain": "regtest"
    },
    "ocean": {
        "rpcurl": "localhost:18010",
        "rpcuser": "bitcoinrpc",
        "rpcpass": "acc1e7a299bc49449912e235b54dbce5",
        "chain": "main"
    }
}
`)

// Test structure
// Set up testing environment for use by regtest demo or unit tests
type Test struct {
	Btc       *rpcclient.Client
	BtcConfig *chaincfg.Params
	Tx0pk     string
	Tx0hash   string
	isRegtest bool
}

// NewTest returns a pointer to a Test instance
func NewTest(logOutput bool, isRegtest bool) *Test {
	// Run init test script that sets up bitcoin and ocean
	var initPath string
	if isRegtest { // for running the demon in regtest mode along with ocean demo
		initPath = os.Getenv("GOPATH") + DEMO_INIT_PATH
	} else { // for running unit tests
		initPath = os.Getenv("GOPATH") + TEST_INIT_PATH
	}

	cmd := exec.Command("/bin/sh", initPath)
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	if logOutput {
		log.Println(string(output))
	}

	btc := conf.GetRPC("btc", testConf)
	btcchaincfg := conf.GetChainCfgParams("btc", testConf)

	// Get first unspent as initial TX for attestation chain
	unspent, errUnspent := btc.ListUnspent()
	if errUnspent != nil {
		log.Fatal(errUnspent)
	}
	var tx0 btcjson.ListUnspentResult
	for _, vout := range unspent {
		if vout.Amount > 50 { // skip regtest txs
			tx0 = vout
		}
	}
	tx0hash := tx0.TxID
	tx0addr, _ := btcutil.DecodeAddress(tx0.Address, btcchaincfg)
	tx0pk, _ := btc.DumpPrivKey(tx0addr)

	return &Test{btc, btcchaincfg, tx0pk.String(), tx0hash, isRegtest}
}

// Return SidechainClient depending on whether Test is regtest or unit-test
func (t *Test) GetSidechainClient() clients.SidechainClient {
	if t.isRegtest {
		return clients.NewSidechainClientOcean(conf.GetRPC("ocean", testConf))
	}
	return clients.NewSidechainClientFake()
}
