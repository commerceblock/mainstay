// Package conf handles reading conf files and establishing client RPC connections.
package config

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
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

// Get RPC connection for a client name from a conf file
func GetRPC(name string, conf []byte) *rpcclient.Client {
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

// Get env from conf file argument using base name and argument name
func GetEnvFromConf(baseName string, argName string, conf []byte) string {
	cfg := getCfg(baseName, conf)
	argValue := cfg.tryGetValue(argName)
	if argValue != "" {
		argValueEnv := os.Getenv(argValue)
		if argValueEnv == "" {
			return argValue
		}
		return argValueEnv
	}
	return ""
}
