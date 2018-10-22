// Package conf handles reading conf files and establishing client RPC connections.
package conf

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
)

// Client RPC connectivity and client related functionality

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
func GetRPC(name string, customConf ...[]byte) *rpcclient.Client {
	var conf []byte
	if len(customConf) > 0 { //custom config provided
		conf = customConf[0]
	} else {
		conf = GetConfFile(os.Getenv("GOPATH") + CONF_PATH)
	}
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
