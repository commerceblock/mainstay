// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package config

import (
	"os"
	"strconv"
	"strings"

	"mainstay/clients"
	"mainstay/log"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
)

// config name consts
const (
	ConfPath                     = "/src/mainstay/config/conf.json"
	MainChainName                = "main"
	StaychainName                = "staychain"
	StaychainRegtestName         = "regtest"
	StaychainInitTxName          = "initTx"
	StaychainInitScriptName      = "initScript"
	StaychainInitPkName          = "initPK"
	StaychainInitChaincodesName  = "initChaincodes"
	StaychainTopupAddressName    = "topupAddress"
	StaychainTopupScriptName     = "topupScript"
	StaychainTopupPkName         = "topupPK"
	StayChainTopupChaincodesName = "topupChaincodes"
)

// Config struct
// Client connections and other parameters required
// by ocean attestation service and testing
type Config struct {
	// main bitcoin rpc connectivity
	mainClient   *rpcclient.Client
	mainChainCfg *chaincfg.Params

	// core staychain config parameters
	regtest         bool
	initTX          string
	initPK          string
	initScript      string
	initChaincodes  []string
	topupAddress    string
	topupScript     string
	topupPK         string
	topupChaincodes []string

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

// Get topup Address
func (c Config) TopupAddress() string {
	return c.topupAddress
}

// Set topup Address
func (c *Config) SetTopupAddress(addr string) {
	c.topupAddress = addr
}

// Get init PK
func (c *Config) InitPK() string {
	return c.initPK
}

// Set init PK
func (c *Config) SetInitPK(pk string) {
	c.initPK = pk
}

// Get Init Script
func (c *Config) InitScript() string {
	return c.initScript
}

// Set Init Script
func (c *Config) SetInitScript(script string) {
	c.initScript = script
}

// Get Init Chaincodes
func (c *Config) InitChaincodes() []string {
	return c.initChaincodes
}

// Set Init Chaincodes
func (c *Config) SetInitChaincodes(chaincodes []string) {
	c.initChaincodes = chaincodes
}

// Get topup Script
func (c *Config) TopupScript() string {
	return c.topupScript
}

// Set topup Script
func (c *Config) SetTopupScript(script string) {
	c.topupScript = script
}

// Get Topup Chaincodes
func (c *Config) TopupChaincodes() []string {
	return c.topupChaincodes
}

// Set Topup Chaincodes
func (c *Config) SetTopupChaincodes(chaincodes []string) {
	c.topupChaincodes = chaincodes
}

// Get topup PK
func (c *Config) TopupPK() string {
	return c.topupPK
}

// Set topup PK
func (c *Config) SetTopupPK(pk string) {
	c.topupPK = pk
}

// Return Config instance
func NewConfig(customConf ...[]byte) (*Config, error) {
	var conf []byte
	if len(customConf) > 0 { //custom config provided
		conf = customConf[0]
	} else {
		var confErr error
		conf, confErr = GetConfFile(os.Getenv("GOPATH") + ConfPath)
		if confErr != nil {
			return nil, confErr
		}
	}

	// get main rpc client
	mainClient, rpcErr := GetRPC(MainChainName, conf)
	if rpcErr != nil {
		return nil, rpcErr
	}

	// get main rpc client chain parameters
	mainClientCfg, paramsErr := GetChainCfgParams(MainChainName, conf)
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
	regtestStr := TryGetParamFromConf(StaychainName, StaychainRegtestName, conf)
	initTxStr := TryGetParamFromConf(StaychainName, StaychainInitTxName, conf)
	initScriptStr := TryGetParamFromConf(StaychainName, StaychainInitScriptName, conf)
	initPKStr := TryGetParamFromConf(StaychainName, StaychainInitPkName, conf)
	topupAddrStr := TryGetParamFromConf(StaychainName, StaychainTopupAddressName, conf)
	topupScriptStr := TryGetParamFromConf(StaychainName, StaychainTopupScriptName, conf)
	topupPKStr := TryGetParamFromConf(StaychainName, StaychainTopupPkName, conf)

	initChaincodesStr := TryGetParamFromConf(StaychainName, StaychainInitChaincodesName, conf)
	initChaincodes := strings.Split(initChaincodesStr, ",") // string to string slice
	for i := range initChaincodes {                         // trim whitespace
		initChaincodes[i] = strings.TrimSpace(initChaincodes[i])
	}
	topupChaincodesStr := TryGetParamFromConf(StaychainName, StayChainTopupChaincodesName, conf)
	topupChaincodes := strings.Split(topupChaincodesStr, ",") // string to string slice
	for i := range topupChaincodes {                          // trim whitespace
		topupChaincodes[i] = strings.TrimSpace(topupChaincodes[i])
	}

	return &Config{
		mainClient:      mainClient,
		mainChainCfg:    mainClientCfg,
		regtest:         (regtestStr == "1"),
		initTX:          initTxStr,
		initPK:          initPKStr,
		initScript:      initScriptStr,
		initChaincodes:  initChaincodes,
		topupAddress:    topupAddrStr,
		topupScript:     topupScriptStr,
		topupPK:         topupPKStr,
		topupChaincodes: topupChaincodes,
		signerConfig:    signerConfig,
		dbConfig:        dbConnectivity,
		feesConfig:      feesConfig,
		timingConfig:    timingConfig,
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
		conf, confErr = GetConfFile(os.Getenv("GOPATH") + ConfPath)
		if confErr != nil {
			log.Error(confErr)
		}
	}

	// get side client rpc
	sideClient, rpcErr := GetRPC(chainName, conf)
	if rpcErr != nil {
		log.Error(rpcErr)
	}
	return clients.NewSidechainClientOcean(sideClient)
}

// db config parameter names
const (
	DbUserName     = "user"
	DbPasswordName = "password"
	DbHostName     = "host"
	DbPortName     = "port"
	DbNameName     = "name"
	DbName         = "db"
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
// If DbName exists in the config, then all fields are compulsory
// IF DbName does not exist, then all config fields are empty
func GetDbConfig(conf []byte) (DbConfig, error) {

	// db connectivity parameters

	user, userErr := GetParamFromConf(DbName, DbUserName, conf)
	if userErr != nil {
		return DbConfig{}, userErr
	}

	password, passwordErr := GetParamFromConf(DbName, DbPasswordName, conf)
	if passwordErr != nil {
		return DbConfig{}, passwordErr
	}

	host, hostErr := GetParamFromConf(DbName, DbHostName, conf)
	if hostErr != nil {
		return DbConfig{}, hostErr
	}

	port, portErr := GetParamFromConf(DbName, DbPortName, conf)
	if portErr != nil {
		return DbConfig{}, portErr
	}

	name, nameErr := GetParamFromConf(DbName, DbNameName, conf)
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
	FeesName             = "fees"
	FeesMinFeeName       = "minFee"
	FeesMaxFeeName       = "maxFee"
	FeesFeeIncrementName = "feeIncrement"
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

	minFeeStr := TryGetParamFromConf(FeesName, FeesMinFeeName, conf)
	var minFee int
	minFeeInt, minFeeErr := strconv.Atoi(minFeeStr)
	if minFeeErr != nil {
		minFee = -1
	} else {
		minFee = minFeeInt
	}

	maxFeeStr := TryGetParamFromConf(FeesName, FeesMaxFeeName, conf)
	var maxFee int
	maxFeeInt, maxFeeErr := strconv.Atoi(maxFeeStr)
	if maxFeeErr != nil {
		maxFee = -1
	} else {
		maxFee = maxFeeInt
	}

	feeIncrementStr := TryGetParamFromConf(FeesName, FeesFeeIncrementName, conf)
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
	TimingName                         = "timing"
	TimingNewAttestationMinutesName    = "newAttestationMinutes"
	TimingHandleUnconfirmedMinutesName = "handleUnconfirmedMinutes"
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
	attMinStr := TryGetParamFromConf(TimingName, TimingNewAttestationMinutesName, conf)
	var attMin int
	attMinInt, attMinIntErr := strconv.Atoi(attMinStr)
	if attMinIntErr != nil {
		attMin = -1
	} else {
		attMin = attMinInt
	}

	uncMinStr := TryGetParamFromConf(TimingName, TimingHandleUnconfirmedMinutesName, conf)
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
	SignerName          = "signer"
	SignerPublisherName = "publisher"
	SignerSignersName   = "signers"
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
// If SignerName exists in conf, SignerSignersName is compsulsory
// Every other Signer Config field is optional
func GetSignerConfig(conf []byte) (SignerConfig, error) {
	// get signer node addresses
	signersStr, signersErr := GetParamFromConf(SignerName, SignerSignersName, conf)
	if signersErr != nil {
		return SignerConfig{}, signersErr
	}
	signers := strings.Split(signersStr, ",")
	for i := range signers {
		signers[i] = strings.TrimSpace(signers[i])
	}
	publisher := TryGetParamFromConf(SignerName, SignerPublisherName, conf)

	return SignerConfig{
		Publisher: publisher,
		Signers:   signers,
	}, nil
}
