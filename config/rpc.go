// Package conf handles reading conf files and establishing client RPC connections.
package config

import (
    "log"
    "os"
    "io/ioutil"
    "github.com/btcsuite/btcd/rpcclient"
    "github.com/btcsuite/btcd/chaincfg"
)

// Client RPC connectivity and client related functionality

// Get default conf from local file
func GetConfFile(filepath string) []byte {
    conf, err := ioutil.ReadFile(filepath)
    if err != nil {
        log.Fatal(err)
    }
    return conf
}

// Get RPC connection for a client from a conf file
func GetRPC(name string, conf []byte) *rpcclient.Client{
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

// Chain configuration parameters from btcsuite for btc client only
func GetChainCfgParams(name string, conf []byte) *chaincfg.Params {
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
