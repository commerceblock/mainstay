// Test struct

package test

import (
    "os/exec"
    "os"
    "log"
    "github.com/btcsuite/btcd/btcjson"
    "github.com/btcsuite/btcutil"
    "github.com/btcsuite/btcd/chaincfg"
    "github.com/btcsuite/btcd/rpcclient"
    "ocean-attestation/conf"
)

const DEMO_INIT_PATH = "/src/ocean-attestation/test/demo-init.sh"
const TEST_INIT_PATH = "/src/ocean-attestation/test/test-init.sh"

var testConf = []byte(`
{
    "btc": {
        "rpcurl": "localhost:18000",
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

type Test struct {
    Btc, Ocean  *rpcclient.Client
    BtcConfig   *chaincfg.Params
    Tx0pk       string
    Tx0hash     string
}

func NewTest(logOutput bool, isDemo bool) *Test {
    // Run init test script that sets up bitcoin and ocean
    var initPath string
    if (isDemo) { // for running the demon in regtest mode along with ocean demo
        initPath = os.Getenv("GOPATH") + DEMO_INIT_PATH
    } else { // for running unit tests
        initPath = os.Getenv("GOPATH") + TEST_INIT_PATH
    }

    cmd := exec.Command("/bin/sh", initPath)
    output, err := cmd.Output()
    if err != nil {
        log.Fatal(err)
    }
    if (logOutput) {
        log.Println(string(output))
    }

    btc  := conf.GetRPC("btc", testConf)
    ocean := conf.GetRPC("ocean", testConf)
    chaincfg := conf.GetChainCfgParams("btc", testConf)

    // Get first unspent as initial TX for attestation chain
    unspent, _ := btc.ListUnspent()
    var tx0 btcjson.ListUnspentResult
    for _, vout := range unspent {
        if (vout.Amount > 50) { // skip regtest txs
            tx0 = vout
        }
    }
    tx0hash := tx0.TxID
    tx0addr, _ := btcutil.DecodeAddress(tx0.Address, chaincfg)
    tx0pk, _ := btc.DumpPrivKey(tx0addr)

    return &Test{btc, ocean, chaincfg, tx0pk.String(), tx0hash}
}
