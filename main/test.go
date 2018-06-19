// Test struct

package main

import (
    "os/exec"
    "os"
    "log"
    "ocean-attestation/conf"
    "github.com/btcsuite/btcutil"
    "github.com/btcsuite/btcd/chaincfg"
    "github.com/btcsuite/btcd/rpcclient"
)

var testConf = []byte(`
{
    "btc": {
        "rpcurl": "localhost:18000",
        "rpcuser": "user",
        "rpcpass": "pass"
    },
    "ocean": {
        "rpcurl": "localhost:18001",
        "rpcuser": "user",
        "rpcpass": "pass"
    }
}
`)

type Test struct {
    btc, ocean  *rpcclient.Client
    tx0pk       string
    tx0hash     string
}

func NewTest() *Test {
    // Run init test script that sets up bitcoin and ocean
    initPath := os.Getenv("GOPATH") + "/src/ocean-attestation/start_test.sh"
    cmd := exec.Command("/bin/sh", initPath)
    _, err := cmd.Output()
    if err != nil {
        log.Fatal(err)
    }
    //log.Println(string(output))

    btc  := conf.GetRPC("btc", testConf)
    ocean := conf.GetRPC("ocean", testConf)

    // Get first unspent as initial TX for attestation chain
    unspent, _ := btc.ListUnspent()
    tx0 := &unspent[0]
    tx0hash := tx0.TxID
    tx0addr, _ := btcutil.DecodeAddress(tx0.Address, &chaincfg.RegressionNetParams)
    tx0pk, _ := btc.DumpPrivKey(tx0addr)

    return &Test{btc, ocean, tx0pk.String(), tx0hash}
}
