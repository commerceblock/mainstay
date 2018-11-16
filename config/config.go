package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"mainstay/clients"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
)

// config name consts
const (
	MAIN_CHAIN_NAME = "main"

	CONF_PATH = "/src/mainstay/config/conf.json"

	MISC_NAME = "misc"
)

// zmq config consts
const (
	MISC_MULTISIGNODES_NAME = "multisignodes"

	MAIN_PUBLISHER_PORT  = 5000
	MAIN_LISTENER_PORT   = 6000
	TOPIC_NEW_HASH       = "H"
	TOPIC_NEW_TX         = "T"
	TOPIC_CONFIRMED_HASH = "C"
	TOPIC_SIGS           = "S"
)

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
	dbConfig       DbConfig
	feesConfig     FeesConfig
	timingConfig   TimingConfig
}

// Get Main Client
func (c Config) MainClient() *rpcclient.Client {
	return c.mainClient
}

// Get Main Client Cfg
func (c Config) MainChainCfg() *chaincfg.Params {
	return c.mainChainCfg
}

// Get Tx Signers host names
func (c Config) MultisigNodes() []string {
	return c.multisigNodes
}

// Get Database configuration
func (c Config) DbConfig() DbConfig {
	return c.dbConfig
}

// Get Fees configuration
func (c Config) FeesConfig() FeesConfig {
	return c.feesConfig
}

// Get Timing configuration
func (c Config) TimingConfig() TimingConfig {
	return c.timingConfig
}

// Set timing configuration
func (c *Config) SetTimingConfig(timingConfig TimingConfig) {
	c.timingConfig = timingConfig
}

// Get init TX
func (c Config) InitTX() string {
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
func NewConfig(customConf ...[]byte) (*Config, error) {
	var conf []byte
	if len(customConf) > 0 { //custom config provided
		conf = customConf[0]
	} else {
		var confErr error
		conf, confErr = GetConfFile(os.Getenv("GOPATH") + CONF_PATH)
		if confErr != nil {
			return nil, confErr
		}
	}

	// get main rpc client
	mainClient, rpcErr := GetRPC(MAIN_CHAIN_NAME, conf)
	if rpcErr != nil {
		return nil, rpcErr
	}

	// get main rpc client chain parameters
	mainClientCfg, paramsErr := GetChainCfgParams(MAIN_CHAIN_NAME, conf)
	if paramsErr != nil {
		return nil, paramsErr
	}

	// get multisig node hosts
	multisigNodesVal, multisigErr := GetParamFromConf(MISC_NAME, MISC_MULTISIGNODES_NAME, conf)
	if multisigErr != nil {
		return nil, multisigErr
	}
	multisignodes := strings.Split(multisigNodesVal, ",")

	// get db connectivity details
	dbConnectivity, dbErr := GetDbConfig(conf)
	if dbErr != nil {
		return nil, dbErr
	}

	feesConfig := GetFeesConfig(conf)
	timingConfig := GetTimingConfig(conf)

	return &Config{
		mainClient:     mainClient,
		mainChainCfg:   mainClientCfg,
		multisigNodes:  multisignodes,
		initTX:         "",
		initPK:         "",
		multisigScript: "",
		dbConfig:       dbConnectivity,
		feesConfig:     feesConfig,
		timingConfig:   timingConfig,
	}, nil
}

// Return SidechainClient depending on whether unit test config or actual config
func NewClientFromConfig(chainName string, isTest bool, customConf ...[]byte) clients.SidechainClient {
	// mock side client rpc for unit-test / regtest
	if isTest {
		return clients.NewSidechainClientFake()
	}

	var conf []byte
	if len(customConf) > 0 { //custom config provided
		conf = customConf[0]
	} else {
		var confErr error
		conf, confErr = GetConfFile(os.Getenv("GOPATH") + CONF_PATH)
		if confErr != nil {
			log.Fatal(confErr)
		}
	}

	// get side client rpc
	sideClient, rpcErr := GetRPC(chainName, conf)
	if rpcErr != nil {
		log.Fatal(rpcErr)
	}
	return clients.NewSidechainClientOcean(sideClient)
}

// db config parameter names
const (
	DB_USER_NAME     = "user"
	DB_PASSWORD_NAME = "password"
	DB_HOST_NAME     = "host"
	DB_PORT_NAME     = "port"
	DB_NAME_NAME     = "name"
	DB_NAME          = "db"
)

// DbConfig struct
// Database connectivity details
type DbConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}

// Return DbConfig from conf options
func GetDbConfig(conf []byte) (DbConfig, error) {

	// db connectivity parameters

	user, userErr := GetParamFromConf(DB_NAME, DB_USER_NAME, conf)
	if userErr != nil {
		return DbConfig{}, userErr
	}

	password, passwordErr := GetParamFromConf(DB_NAME, DB_PASSWORD_NAME, conf)
	if passwordErr != nil {
		return DbConfig{}, passwordErr
	}

	host, hostErr := GetParamFromConf(DB_NAME, DB_HOST_NAME, conf)
	if hostErr != nil {
		return DbConfig{}, hostErr
	}

	port, portErr := GetParamFromConf(DB_NAME, DB_PORT_NAME, conf)
	if portErr != nil {
		return DbConfig{}, portErr
	}

	name, nameErr := GetParamFromConf(DB_NAME, DB_NAME_NAME, conf)
	if nameErr != nil {
		return DbConfig{}, nameErr
	}

	return DbConfig{
		User:     user,
		Password: password,
		Host:     host,
		Port:     port,
		Name:     name,
	}, nil
}

// fee config parameter names
const (
	FEES_NAME               = "fees"
	FEES_MIN_FEE_NAME       = "minFee"
	FEES_MAX_FEE_NAME       = "maxFee"
	FEES_FEE_INCREMENT_NAME = "feeIncrement"
)

// FeeConfig struct
// Configuration on fee limits for attestation service
type FeesConfig struct {
	MinFee       int
	MaxFee       int
	FeeIncrement int
}

// Return FeeConfig from conf options
func GetFeesConfig(conf []byte) FeesConfig {
	// try getting all config parameters
	// all are optional so if no value is found
	// we set to invalid value

	minFeeStr := TryGetParamFromConf(FEES_NAME, FEES_MIN_FEE_NAME, conf)
	var minFee int
	minFeeInt, minFeeErr := strconv.Atoi(minFeeStr)
	if minFeeErr != nil {
		minFee = -1
	} else {
		minFee = minFeeInt
	}

	maxFeeStr := TryGetParamFromConf(FEES_NAME, FEES_MAX_FEE_NAME, conf)
	var maxFee int
	maxFeeInt, maxFeeErr := strconv.Atoi(maxFeeStr)
	if maxFeeErr != nil {
		maxFee = -1
	} else {
		maxFee = maxFeeInt
	}

	feeIncrementStr := TryGetParamFromConf(FEES_NAME, FEES_FEE_INCREMENT_NAME, conf)
	var feeIncrement int
	feeIncrementInt, feeIncrementErr := strconv.Atoi(feeIncrementStr)
	if feeIncrementErr != nil {
		feeIncrement = -1
	} else {
		feeIncrement = feeIncrementInt
	}

	return FeesConfig{
		MinFee:       minFee,
		MaxFee:       maxFee,
		FeeIncrement: feeIncrement,
	}
}

// timing config parameter names
const (
	TIMING_NAME                            = "timing"
	TIMING_NEW_ATTESTATION_MINUTES_NAME    = "newAttestationMinutes"
	TIMING_HANDLE_UNCONFIRMED_MINUTES_NAME = "handleUnconfirmedMinutes"
)

// Timing config struct
// Configuration on wait time duration for various things in attestation service
type TimingConfig struct {
	NewAttestationMinutes    int
	HandleUnconfirmedMinutes int
}

// Return TimingConfig from conf options
func GetTimingConfig(conf []byte) TimingConfig {
	attMinStr := TryGetParamFromConf(TIMING_NAME, TIMING_NEW_ATTESTATION_MINUTES_NAME, conf)
	var attMin int
	attMinInt, attMinIntErr := strconv.Atoi(attMinStr)
	if attMinIntErr != nil {
		attMin = -1
	} else {
		attMin = attMinInt
	}

	uncMinStr := TryGetParamFromConf(TIMING_NAME, TIMING_HANDLE_UNCONFIRMED_MINUTES_NAME, conf)
	var uncMin int
	uncMinInt, uncMinIntErr := strconv.Atoi(uncMinStr)
	if uncMinIntErr != nil {
		uncMin = -1
	} else {
		uncMin = uncMinInt
	}

	return TimingConfig{
		NewAttestationMinutes:    attMin,
		HandleUnconfirmedMinutes: uncMin,
	}
}
