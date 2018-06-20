// Client RPC connectivity and client related functionality

package conf

import (
    "log"
    "os"
    "io/ioutil"
    "github.com/btcsuite/btcd/rpcclient"
    "github.com/btcsuite/btcd/chaincfg"
)

// Get default conf from local file
func getDefaultConf() []byte {
    conf, err := ioutil.ReadFile(os.Getenv("GOPATH") + "/src/ocean-attestation/conf/conf.json")
    if err != nil {
        log.Fatal(err)
    }
    return conf
}

// Get RPC connection for a client from a conf file
func GetRPC(name string, customConf ...[]byte) *rpcclient.Client{
    var conf []byte
    if len(customConf) > 0 { //custom config provided
        conf = customConf[0]
    } else {
        conf = getDefaultConf()
    }
    cfg := getCfg(name, conf)
    connCfg := &rpcclient.ConnConfig{
        Host:         cfg.getValue("rpcurl"),
        User:         cfg.getValue("rpcuser"),
        Pass:         cfg.getValue("rpcpass"),
        HTTPPostMode: true,
        DisableTLS:   true,
    }
    client, err := rpcclient.New(connCfg, nil)
    if err != nil {
        log.Fatal(err)
    }
    return client
}

func GetChainCfgParams(name string, customConf ...[]byte) *chaincfg.Params {
    var conf []byte
    if len(customConf) > 0 { //custom config provided
        conf = customConf[0]
    } else {
        conf = getDefaultConf()
    }

    cfg := getCfg(name, conf)

    chain := cfg.getValue("chain")

    if chain == "regtest" {
        return &chaincfg.RegressionNetParams
    } else if chain == "testnet" {
        return &chaincfg.TestNet3Params
    } else if chain == "main" {
        return &chaincfg.MainNetParams
    }
    return &chaincfg.RegressionNetParams
}
