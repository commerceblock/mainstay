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

const INIT_PATH = "/src/ocean-attestation/test/test-init.sh"

var testConf = []byte(`
{
    "btc": {
        "rpcurl": "localhost:18000",
        "rpcuser": "user",
        "rpcpass": "pass",
        "chain": "regtest"
    },
    "ocean": {
        "rpcurl": "localhost:18001",
        "rpcuser": "user",
        "rpcpass": "pass",
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

func NewTest(logOutput bool) *Test {
    // Run init test script that sets up bitcoin and ocean
    initPath := os.Getenv("GOPATH") + INIT_PATH
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
