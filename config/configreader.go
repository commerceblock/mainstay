package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
)

// Client RPC connectivity and client related functionality

const (
	RPC_CLIENT_URL_NAME   = "rpcurl"
	RPC_CLIENT_USER_NAME  = "rpcuser"
	RPC_CLIENT_PASS_NAME  = "rpcpass"
	RPC_CLIENT_CHAIN_NAME = "chain"

	ERROR_RPC_CONNECTION_FAILURE = "failed connecting to rpc client"

	BAD_DATA_CLIENT_CHAIN = "invalid value for client chain. 'main', 'testnet' and 'regtest' allowed only"
)

// Get default conf from local file
func GetConfFile(filepath string) ([]byte, error) {
	conf, err := ioutil.ReadFile(filepath)
	if err != nil {
		return []byte{}, err
	}
	return conf, nil
}

// Get RPC connection for a client name from a conf file
func GetRPC(name string, conf []byte) (*rpcclient.Client, error) {
	// get client from config
	cfg, cfgErr := getCfg(name, conf)
	if cfgErr != nil {
		return nil, errors.New(fmt.Sprintf("%s %s", cfgErr, name))
	}

	// get client url value
	urlValue, urlValueErr := cfg.getValue(RPC_CLIENT_URL_NAME)
	if urlValueErr != nil {
		return nil, errors.New(fmt.Sprintf("%s %s", urlValueErr, RPC_CLIENT_URL_NAME))
	}
	host := os.Getenv(urlValue)
	if host == "" {
		host = urlValue
	}

	// get client user value
	userValue, userValueErr := cfg.getValue(RPC_CLIENT_USER_NAME)
	if userValueErr != nil {
		return nil, errors.New(fmt.Sprintf("%s %s", userValueErr, RPC_CLIENT_USER_NAME))
	}
	user := os.Getenv(userValue)
	if user == "" {
		user = userValue
	}

	// get client password value
	passValue, passValueErr := cfg.getValue(RPC_CLIENT_PASS_NAME)
	if passValueErr != nil {
		return nil, errors.New(fmt.Sprintf("%s %s", passValueErr, RPC_CLIENT_PASS_NAME))
	}
	pass := os.Getenv(passValue)
	if pass == "" {
		pass = passValue
	}

	connCfg := &rpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         pass,
		HTTPPostMode: true,
		DisableTLS:   true,
	}
	client, rpcErr := rpcclient.New(connCfg, nil)
	if rpcErr != nil {
		return nil, errors.New(fmt.Sprintf("%s %s", rpcErr, ERROR_RPC_CONNECTION_FAILURE))
	}
	return client, nil
}

// Chain configuration parameters from btcsuite for main bitcoin client only
func GetChainCfgParams(name string, conf []byte) (*chaincfg.Params, error) {
	cfg, cfgErr := getCfg(name, conf)
	if cfgErr != nil {
		return nil, errors.New(fmt.Sprintf("%s %s", cfgErr, name))
	}

	chain, valueErr := cfg.getValue(RPC_CLIENT_CHAIN_NAME)
	if valueErr != nil {
		return nil, errors.New(fmt.Sprintf("%s %s", valueErr, chain))
	}
	if chain == "regtest" {
		return &chaincfg.RegressionNetParams, nil
	} else if chain == "testnet" {
		return &chaincfg.TestNet3Params, nil
	} else if chain == "main" {
		return &chaincfg.MainNetParams, nil
	}
	return nil, errors.New(BAD_DATA_CLIENT_CHAIN)
}

// Get parameter from conf file argument using base name and argument name
// We first test if this is an env variable and if not we return value as is
func GetParamFromConf(baseName string, argName string, conf []byte) (string, error) {
	cfg, cfgErr := getCfg(baseName, conf)
	if cfgErr != nil {
		return "", errors.New(fmt.Sprintf("%s %s", cfgErr, baseName))
	}

	argValue := cfg.tryGetValue(argName)
	if argValue != "" {
		argValueEnv := os.Getenv(argValue)
		if argValueEnv == "" {
			return argValue, nil
		}
		return argValueEnv, nil
	}
	return "", nil
}
