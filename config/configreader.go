// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

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
	RpcClientUrlName   = "rpcurl"
	RpcClientUserName  = "rpcuser"
	RpcClientPassName  = "rpcpass"
	RpcClientChainName = "chain"

	ErrorRpcConnectionFailure = "failed connecting to rpc client"

	ErrorBadDataClientChain = "invalid value for client chain. 'main', 'testnet' and 'regtest' allowed only"
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
		return nil, errors.New(fmt.Sprintf("%s: %s", cfgErr, name))
	}

	// get client url value
	urlValue, urlValueErr := cfg.getValue(RpcClientUrlName)
	if urlValueErr != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s", urlValueErr, RpcClientUrlName))
	}
	host := os.Getenv(urlValue)
	if host == "" {
		host = urlValue
	}

	// get client user value
	userValue, userValueErr := cfg.getValue(RpcClientUserName)
	if userValueErr != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s", userValueErr, RpcClientUserName))
	}
	user := os.Getenv(userValue)
	if user == "" {
		user = userValue
	}

	// get client password value
	passValue, passValueErr := cfg.getValue(RpcClientPassName)
	if passValueErr != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s", passValueErr, RpcClientPassName))
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
		return nil, errors.New(fmt.Sprintf("%s: %s", rpcErr, ErrorRpcConnectionFailure))
	}
	return client, nil
}

// Chain configuration parameters from btcsuite for main bitcoin client only
func GetChainCfgParams(name string, conf []byte) (*chaincfg.Params, error) {
	cfg, cfgErr := getCfg(name, conf)
	if cfgErr != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s", cfgErr, name))
	}

	// error if RpcClientChainName not found in main config
	chainValue, chainValueErr := cfg.getValue(RpcClientChainName)
	if chainValueErr != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s", chainValueErr, RpcClientChainName))
	}

	// try get env or keep current value
	chain := os.Getenv(chainValue)
	if chain == "" {
		chain = chainValue
	}

	if chain == "regtest" {
		return &chaincfg.RegressionNetParams, nil
	} else if chain == "testnet" {
		return &chaincfg.TestNet3Params, nil
	}
	// mainnet returned unless specified otherwise
	return &chaincfg.MainNetParams, nil
}

// Get parameter from conf file argument using base name and argument name
// If base name does not exist we don't try to get the values from conf
// We first test if this is an env variable and if not we return value as is
func GetParamFromConf(baseName string, argName string, conf []byte) (string, error) {
	cfg, cfgErr := getCfg(baseName, conf)
	if cfgErr != nil {
		return "", nil
	}

	argValue, valueErr := cfg.getValue(argName)
	if valueErr != nil {
		return "", errors.New(fmt.Sprintf("%s: %s", valueErr, argName))
	}

	argValueEnv := os.Getenv(argValue)
	if argValueEnv == "" {
		return argValue, nil
	}
	return argValueEnv, nil
}

// Get parameter from conf file argument using base name and argument name
// We first test if this is an env variable and if not we return value as is
func TryGetParamFromConf(baseName string, argName string, conf []byte) string {
	cfg, cfgErr := getCfg(baseName, conf)
	if cfgErr != nil {
		return ""
	}

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
