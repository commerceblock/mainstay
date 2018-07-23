// Client RPC connectivity and client related functionality

package conf

import (
    "log"
    "os"
    "io/ioutil"
    "github.com/btcsuite/btcd/rpcclient"
    "github.com/btcsuite/btcd/chaincfg"
)

const CONF_PATH = "/src/ocean-attestation/conf/conf.json"

// Get default conf from local file
func GetConfFile(filepath string) []byte {
    conf, err := ioutil.ReadFile(filepath)
    if err != nil {
        log.Fatal(err)
    }
    return conf
}

// Get RPC connection for a client from a conf file
func GetRPC(name string, customConf ...[]byte) *rpcclient.Client{
    var conf []byte
    var host, user, pass string
    if len(customConf) > 0 { //custom config provided
        conf = customConf[0]
        cfg := getCfg(name, conf)
        host = cfg.getValue("rpcurl")
        user = cfg.getValue("rpcuser")
        pass = cfg.getValue("rpcpass")
    } else {
        conf = GetConfFile(os.Getenv("GOPATH") + CONF_PATH)
        cfg := getCfg(name, conf)
        host := os.Getenv(cfg.getValue("rpcurl"))
        if host == "" {
            host = cfg.getValue("rpcurl")
        }
        user := os.Getenv(cfg.getValue("rpcuser"))
        if user == "" {
            user = cfg.getValue("rpcuser")
        }
        pass := os.Getenv(cfg.getValue("rpcpass"))
        if pass == "" {
            pass = cfg.getValue("rpcpass")
        }
    }
    connCfg := &rpcclient.ConnConfig{
        Host:         host,
        User:         user,
        Pass:         pass,
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
        conf = GetConfFile(os.Getenv("GOPATH") + CONF_PATH)
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
