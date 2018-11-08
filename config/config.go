package config

import (
	"os"
	"strings"

	"mainstay/clients"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
)

const MAIN_CHAIN_NAME = "main"
const SIDE_CHAIN_NAME = "ocean"
const CONF_PATH = "/src/mainstay/config/conf.json"

const MAIN_PUBLISHER_PORT = 5000
const MAIN_LISTENER_PORT = 6000
const TOPIC_NEW_HASH = "H"
const TOPIC_NEW_TX = "T"
const TOPIC_CONFIRMED_HASH = "C"
const TOPIC_SIGS = "S"

// Config struct
// Client connections and other parameters required
// by ocean attestation service and testing
type Config struct {
	mainClient     *rpcclient.Client
	mainChainCfg   *chaincfg.Params
	multisigNodes  []string
	initTX         string
	initPK         string
	multisigScript string
	dbConnectivity DbConnectivity
}

// Get Main Client
func (c *Config) MainClient() *rpcclient.Client {
	return c.mainClient
}

// Get Main Client Cfg
func (c *Config) MainChainCfg() *chaincfg.Params {
	return c.mainChainCfg
}

// Get Tx Signers host names
func (c *Config) MultisigNodes() []string {
	return c.multisigNodes
}

// Get Tx Signers host names
func (c *Config) DbConnectivity() DbConnectivity {
	return c.dbConnectivity
}

// Get init TX
func (c *Config) InitTX() string {
	return c.initTX
}

// Set init TX
func (c *Config) SetInitTX(tx string) {
	c.initTX = tx
}

// Get init PK
func (c *Config) InitPK() string {
	return c.initPK
}

// Set init PK
func (c *Config) SetInitPK(pk string) {
	c.initPK = pk
}

// Get init PK
func (c *Config) MultisigScript() string {
	return c.multisigScript
}

// Set init PK
func (c *Config) SetMultisigScript(script string) {
	c.multisigScript = script
}

// Return Config instance
func NewConfig(customConf ...[]byte) *Config {
	var conf []byte
	if len(customConf) > 0 { //custom config provided
		conf = customConf[0]
	} else {
		conf = GetConfFile(os.Getenv("GOPATH") + CONF_PATH)
	}

	mainClient := GetRPC(MAIN_CHAIN_NAME, conf)
	mainClientCfg := GetChainCfgParams(MAIN_CHAIN_NAME, conf)

	multisignodes := strings.Split(GetEnvFromConf("misc", "multisignodes", conf), ",")

	dbConnectivity := GetDbConnectivity(conf)
	return &Config{mainClient, mainClientCfg, multisignodes, "", "", "", dbConnectivity}
}

// Return SidechainClient depending on whether unit test config or actual config
func NewClientFromConfig(isTest bool, customConf ...[]byte) clients.SidechainClient {
	if isTest {
		return clients.NewSidechainClientFake()
	}

	var conf []byte
	if len(customConf) > 0 { //custom config provided
		conf = customConf[0]
	} else {
		conf = GetConfFile(os.Getenv("GOPATH") + CONF_PATH)
	}
	return clients.NewSidechainClientOcean(GetRPC(SIDE_CHAIN_NAME, conf))
}

// DbDetails struct
// Database connectivity details
type DbConnectivity struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}

// Return DbConnectivity from conf options
func GetDbConnectivity(conf []byte) DbConnectivity {
	return DbConnectivity{
		User:     GetEnvFromConf("db", "user", conf),
		Password: GetEnvFromConf("db", "password", conf),
		Host:     GetEnvFromConf("db", "host", conf),
		Port:     GetEnvFromConf("db", "port", conf),
		Name:     GetEnvFromConf("db", "name", conf),
	}
}
