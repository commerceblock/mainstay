// Package conf handles reading conf files and establishing client RPC connections.
package config

import (
    "os"

    "ocean-attestation/clients"

    "github.com/btcsuite/btcd/chaincfg"
    "github.com/btcsuite/btcd/rpcclient"
)

const MAIN_CHAIN_NAME = "main"
const SIDE_CHAIN_NAME = "ocean"
const CONF_PATH = "/src/ocean-attestation/config/conf.json"

// Config struct
// Client connections and other parameters required
// by ocean attestation service and testing
type Config struct {
    mainClient              *rpcclient.Client
    mainChainCfg            *chaincfg.Params
    oceanClient             clients.SidechainClient
}

// Get Main Client
func (c *Config) MainClient() *rpcclient.Client {
    return c.mainClient
}

// Get Ocean Client
func (c *Config) OceanClient() clients.SidechainClient {
    return c.oceanClient
}

// Get Main Client Cfg
func (c *Config) MainChainCfg() *chaincfg.Params {
    return c.mainChainCfg
}

// Return Config instance
func NewConfig(isUnitTest bool, customConf ...[]byte) *Config {
    var conf []byte
    if len(customConf) > 0 { //custom config provided
        conf = customConf[0]
    } else {
        conf = GetConfFile(os.Getenv("GOPATH") + CONF_PATH)
    }

    mainClient := GetRPC(MAIN_CHAIN_NAME, conf)
    mainClientCfg := GetChainCfgParams(MAIN_CHAIN_NAME, conf)
    oceanClient := GetSidechainClient(isUnitTest, conf)
    return &Config{mainClient, mainClientCfg, oceanClient}
}

// Return SidechainClient depending on whether unit test config or actual config
func GetSidechainClient(isUnitTest bool, conf []byte) clients.SidechainClient {
    if isUnitTest {
        return clients.NewSidechainClientFake()
    }
    return clients.NewSidechainClientOcean(GetRPC(SIDE_CHAIN_NAME, conf))
}
