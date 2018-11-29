// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

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
	CONF_PATH                   = "/src/mainstay/config/conf.json"
	MAIN_CHAIN_NAME             = "main"
	STAYCHAIN_NAME              = "staychain"
	STAYCHAIN_REGTEST_NAME      = "regtest"
	STAYCHAIN_INIT_TX_NAME      = "initTx"
	STAYCHAIN_INIT_SCRIPT_NAME  = "initScript"
	STAYCHAIN_TOPUP_TX_NAME     = "topupTx"
	STAYCHAIN_TOPUP_SCRIPT_NAME = "topupScript"
)

// Config struct
// Client connections and other parameters required
// by ocean attestation service and testing
type Config struct {
	// main bitcoin rpc connectivity
	mainClient   *rpcclient.Client
	mainChainCfg *chaincfg.Params

	// core staychain config parameters
	regtest     bool
	initTX      string
	initPK      string
	initScript  string
	topupTx     string
	topupScript string

	// additional parameter categories
	signerConfig SignerConfig
	dbConfig     DbConfig
	feesConfig   FeesConfig
	timingConfig TimingConfig
}

// Get Main Client
func (c Config) MainClient() *rpcclient.Client {
	return c.mainClient
}

// Get Main Client Cfg
func (c Config) MainChainCfg() *chaincfg.Params {
	return c.mainChainCfg
}

// Get Signer configuration
func (c Config) SignerConfig() SignerConfig {
	return c.signerConfig
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

// Get regtest flag
func (c Config) Regtest() bool {
	return c.regtest
}

// Set regtest flag
func (c *Config) SetRegtest(regtest bool) {
	c.regtest = regtest
}

// Get init TX
func (c Config) InitTx() string {
	return c.initTX
}

// Set init TX
func (c *Config) SetInitTx(tx string) {
	c.initTX = tx
}

// Get topup TX
func (c Config) TopupTx() string {
	return c.topupTx
}

// Set topup TX
func (c *Config) SetTopupTx(tx string) {
	c.topupTx = tx
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
func (c *Config) InitScript() string {
	return c.initScript
}

// Set init PK
func (c *Config) SetInitScript(script string) {
	c.initScript = script
}

// Get topup PK
func (c *Config) TopupScript() string {
	return c.topupScript
}

// Set init PK
func (c *Config) SetTopupScript(script string) {
	c.topupScript = script
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

	// get db connectivity details
	dbConnectivity, dbErr := GetDbConfig(conf)
	if dbErr != nil {
		return nil, dbErr
	}

	feesConfig := GetFeesConfig(conf)
	timingConfig := GetTimingConfig(conf)

	signerConfig, signerConfigErr := GetSignerConfig(conf)
	if signerConfigErr != nil {
		return nil, signerConfigErr
	}

	// get staychain config parameters
	// most of these can be overriden from command line
	regtestStr := TryGetParamFromConf(STAYCHAIN_NAME, STAYCHAIN_REGTEST_NAME, conf)
	initTxStr := TryGetParamFromConf(STAYCHAIN_NAME, STAYCHAIN_INIT_TX_NAME, conf)
	initScriptStr := TryGetParamFromConf(STAYCHAIN_NAME, STAYCHAIN_INIT_SCRIPT_NAME, conf)
	topupTxStr := TryGetParamFromConf(STAYCHAIN_NAME, STAYCHAIN_TOPUP_TX_NAME, conf)
	topupScriptStr := TryGetParamFromConf(STAYCHAIN_NAME, STAYCHAIN_TOPUP_SCRIPT_NAME, conf)

	return &Config{
		mainClient:   mainClient,
		mainChainCfg: mainClientCfg,
		regtest:      (regtestStr == "1"),
		initTX:       initTxStr,
		initPK:       "",
		initScript:   initScriptStr,
		topupTx:      topupTxStr,
		topupScript:  topupScriptStr,
		signerConfig: signerConfig,
		dbConfig:     dbConnectivity,
		feesConfig:   feesConfig,
		timingConfig: timingConfig,
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
// If DB_NAME exists in the config, then all fields are compulsory
// IF DB_NAME does not exist, then all config fields are empty
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
// All Fees Config fields are optional
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
// All Timing Config fields are optional
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

// signer config parameter names
const (
	SIGNER_NAME           = "signer"
	SIGNER_PUBLISHER_NAME = "publisher"
	SIGNER_SIGNERS_NAME   = "signers"
)

// Signer config struct
// Configuration on communication between service and signers
// Configure host addresses and zmq TOPIC config
type SignerConfig struct {
	// main publisher address
	Publisher string

	// signer addresses
	Signers []string
}

// Return SignerConfig from conf options
// If SIGNER_NAME exists in conf, SIGNER_SIGNERS_NAME is compsulsory
// Every other Signer Config field is optional
func GetSignerConfig(conf []byte) (SignerConfig, error) {
	// get signer node addresses
	signersStr, signersErr := GetParamFromConf(SIGNER_NAME, SIGNER_SIGNERS_NAME, conf)
	if signersErr != nil {
		return SignerConfig{}, signersErr
	}
	signers := strings.Split(signersStr, ",")

	publisher := TryGetParamFromConf(SIGNER_NAME, SIGNER_PUBLISHER_NAME, conf)

	return SignerConfig{
		Publisher: publisher,
		Signers:   signers,
	}, nil
}
