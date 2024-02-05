// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package config

import (
	"errors"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/assert"
)

// Test various Config error cases
func TestConfigErrors(t *testing.T) {
	var testConf = []byte(`
    {
    }
    `)
	config, configErr := NewConfig(testConf)
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ErroConfigNameNotFound, MainChainName)), configErr)

	testConf = []byte(`
    {
        "main": {
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ErrorConfigValueNotFound, RpcClientUrlName)), configErr)

	testConf = []byte(`
    {
        "main": {
            "rpcurl": ""
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ErrorConfigValueNotFound, RpcClientUserName)), configErr)

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "",
            "rpcuser": ""
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ErrorConfigValueNotFound, RpcClientPassName)), configErr)

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "",
            "rpcuser": "",
            "rpcpass": ""
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ErrorConfigValueNotFound, RpcClientChainName)), configErr)

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "",
            "rpcuser": "",
            "rpcpass": "",
            "chain": "testnet"
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, nil, configErr)
	assert.Equal(t, &chaincfg.TestNet3Params, config.MainChainCfg())

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "",
            "rpcuser": "",
            "rpcpass": "",
            "chain": "allaloum"
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, nil, configErr)
	assert.Equal(t, &chaincfg.MainNetParams, config.MainChainCfg())

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "",
            "rpcuser": "",
            "rpcpass": "",
            "chain": "testnet"
        },
        "db": {
            "user": ""
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ErrorConfigValueNotFound, DbPasswordName)), configErr)

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "",
            "rpcuser": "",
            "rpcpass": "",
            "chain": "testnet"
        },
        "db": {
            "user": "",
            "password": ""
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ErrorConfigValueNotFound, DbHostName)), configErr)

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "",
            "rpcuser": "",
            "rpcpass": "",
            "chain": "testnet"
        },
        "db": {
            "user": "",
            "password": "",
            "host": ""
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ErrorConfigValueNotFound, DbPortName)), configErr)

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "",
            "rpcuser": "",
            "rpcpass": "",
            "chain": "testnet"
        },
        "db": {
            "user": "",
            "password": "",
            "host": "",
            "port": ""
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ErrorConfigValueNotFound, DbNameName)), configErr)

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "",
            "rpcuser": "",
            "rpcpass": "",
            "chain": "testnet"
        },
        "db": {
            "user": "",
            "password": "",
            "host": "",
            "port": "",
            "name": ""
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, nil, configErr)
	assert.Equal(t, &chaincfg.TestNet3Params, config.MainChainCfg())
}

// Test actual Config parses correct values
func TestConfigActual(t *testing.T) {
	var configErr error
	var config *Config
	var testConf = []byte(`
    {
        "main": {
            "rpcurl": "localhost:18443",
            "rpcuser": "user",
            "rpcpass": "pass",
            "chain": "regtest"
        },
        "signer": {
            "url": "127.0.0.1:8000"
        },
        "db": {
            "user":"username1",
            "password":"password2",
            "host":"localhost",
            "port":"27017",
            "name":"mainstay"
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, nil, configErr)

	assert.Equal(t, true, config.MainClient() != nil)
	assert.Equal(t, &chaincfg.RegressionNetParams, config.MainChainCfg())
	assert.Equal(t, "127.0.0.1:8000", config.SignerConfig().Url)
	assert.Equal(t, DbConfig{
		User:     "username1",
		Password: "password2",
		Host:     "localhost",
		Port:     "27017",
		Name:     "mainstay",
	}, config.DbConfig())
}

// Test config for Optional staychain parameters
func TestConfigStaychain(t *testing.T) {
	var configErr error
	var config *Config
	var testConf = []byte(`
    {
        "main": {
            "rpcurl": "",
            "rpcuser": "",
            "rpcpass": "",
            "chain": ""
        },
        "staychain": {
            "initTx": "87e56bda501ba6a022f12e178e9f1ac03fb2c07f04e1dfa62ac9e1d83cd840e1",
            "initChaincode": "0a090f710e47968aee906804f211cf10cde9a11e14908ca0f78cc55dd190ceaa",
            "initPK": "cQca2KvrBnJJUCYa2tD4RXhiQshWLNMSK2A96ZKWo1SZkHhh3YLz",
            "topupAddress": "2MxBi6eodnuoVCw8McGrf1nuoVhastqoBXB",
            "topupPK": "cQca2KvrBnJJUCYa2tD4RXhiQshWLNMSK2A96ZKWo1SZkHhh3YLa",
            "regtest": "1"
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, nil, configErr)
	assert.Equal(t, &chaincfg.MainNetParams, config.MainChainCfg())

	assert.Equal(t, "87e56bda501ba6a022f12e178e9f1ac03fb2c07f04e1dfa62ac9e1d83cd840e1", config.InitTx())
	assert.Equal(t, "cQca2KvrBnJJUCYa2tD4RXhiQshWLNMSK2A96ZKWo1SZkHhh3YLz", config.InitPK())
	assert.Equal(t, "0a090f710e47968aee906804f211cf10cde9a11e14908ca0f78cc55dd190ceaa", config.InitChaincode())
	assert.Equal(t, "2MxBi6eodnuoVCw8McGrf1nuoVhastqoBXB", config.TopupAddress())
	assert.Equal(t, "cQca2KvrBnJJUCYa2tD4RXhiQshWLNMSK2A96ZKWo1SZkHhh3YLa", config.TopupPK())
	assert.Equal(t, true, config.Regtest())

	config.SetRegtest(false)
	assert.Equal(t, false, config.Regtest())

	config.SetInitTx("aa")
	assert.Equal(t, "aa", config.InitTx())

	config.SetInitPK("PKPKPK")
	assert.Equal(t, "PKPKPK", config.InitPK())

	config.SetInitChaincode("chaincode1")
	assert.Equal(t, "chaincode1", config.InitChaincode())

	config.SetTopupAddress("cc")
	assert.Equal(t, "cc", config.TopupAddress())

	config.SetTopupPK("TOPUPPKPK")
	assert.Equal(t, "TOPUPPKPK", config.TopupPK())

	config.SetInitChaincode("chaincode3")
	assert.Equal(t, "chaincode3", config.InitChaincode())

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "",
            "rpcuser": "",
            "rpcpass": "",
            "chain": ""
        },
        "staychain": {
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, nil, configErr)

	assert.Equal(t, "", config.InitTx())
	assert.Equal(t, "", config.TopupAddress())
	assert.Equal(t, false, config.Regtest())
}

// Test config for Optional fees parameters
func TestConfigFees(t *testing.T) {
	var configErr error
	var config *Config
	var testConf = []byte(`
    {
        "main": {
            "rpcurl": "localhost:18443",
            "rpcuser": "user",
            "rpcpass": "pass",
            "chain": "regtest"
        },
        "fees": {
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, nil, configErr)
	assert.Equal(t, FeesConfig{-1, -1, -1}, config.FeesConfig())

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "localhost:18443",
            "rpcuser": "user",
            "rpcpass": "pass",
            "chain": "regtest"
        },
        "fees": {
            "minFee": "1"
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, nil, configErr)
	assert.Equal(t, FeesConfig{1, -1, -1}, config.FeesConfig())

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "localhost:18443",
            "rpcuser": "user",
            "rpcpass": "pass",
            "chain": "regtest"
        },
        "fees": {
            "minFee": "invalid"
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, nil, configErr)
	assert.Equal(t, FeesConfig{-1, -1, -1}, config.FeesConfig())

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "localhost:18443",
            "rpcuser": "user",
            "rpcpass": "pass",
            "chain": "regtest"
        },
        "fees": {
            "maxFee": "10",
            "minFee": "5",
            "feeIncrement": "11",
            "something-else": "nice-value"
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, nil, configErr)
	assert.Equal(t, FeesConfig{5, 10, 11}, config.FeesConfig())
}

// Test config for Optional timing parameters
func TestConfigTiming(t *testing.T) {
	var configErr error
	var config *Config
	var testConf = []byte(`
    {
        "main": {
            "rpcurl": "localhost:18443",
            "rpcuser": "user",
            "rpcpass": "pass",
            "chain": "regtest"
        },
        "timing": {
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, nil, configErr)
	assert.Equal(t, TimingConfig{-1, -1}, config.TimingConfig())

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "localhost:18443",
            "rpcuser": "user",
            "rpcpass": "pass",
            "chain": "regtest"
        },
        "timing": {
            "newAttestationMinutes": "0"
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, nil, configErr)
	assert.Equal(t, TimingConfig{0, -1}, config.TimingConfig())

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "localhost:18443",
            "rpcuser": "user",
            "rpcpass": "pass",
            "chain": "regtest"
        },
        "timing": {
            "handleUnconfirmedMinutes": "0"
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, nil, configErr)
	assert.Equal(t, TimingConfig{-1, 0}, config.TimingConfig())

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "localhost:18443",
            "rpcuser": "user",
            "rpcpass": "pass",
            "chain": "regtest"
        },
        "timing": {
            "newAttestationMinutes": "10",
            "handleUnconfirmedMinutes": "60"
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, nil, configErr)
	assert.Equal(t, TimingConfig{10, 60}, config.TimingConfig())
}

// Test config for Optional signer parameters
func TestConfigSigner(t *testing.T) {
	var config *Config
	var configErr error
	var testConf = []byte(`
    {
        "main": {
            "rpcurl": "",
            "rpcuser": "",
            "rpcpass": "",
            "chain": ""
        },
        "signer": {
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ErrorConfigValueNotFound, Url)), configErr)

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "",
            "rpcuser": "",
            "rpcpass": "",
            "chain": ""
        },
        "signer": {
            "url": "host"
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, nil, configErr)
	assert.Equal(t, "host", config.SignerConfig().Url)
}
