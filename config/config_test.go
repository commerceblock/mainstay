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
	var configErr error
	var testConf = []byte(`
    {
    }
    `)
	_, configErr = NewConfig(testConf)
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ERROR_CONFIG_NAME_NOT_FOUND, MAIN_CHAIN_NAME)), configErr)

	testConf = []byte(`
    {
        "main": {
        }
    }
    `)
	_, configErr = NewConfig(testConf)
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ERROR_CONFIG_VALUE_NOT_FOUND, RPC_CLIENT_URL_NAME)), configErr)

	testConf = []byte(`
    {
        "main": {
            "rpcurl": ""
        }
    }
    `)
	_, configErr = NewConfig(testConf)
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ERROR_CONFIG_VALUE_NOT_FOUND, RPC_CLIENT_USER_NAME)), configErr)

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "",
            "rpcuser": ""
        }
    }
    `)
	_, configErr = NewConfig(testConf)
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ERROR_CONFIG_VALUE_NOT_FOUND, RPC_CLIENT_PASS_NAME)), configErr)

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "",
            "rpcuser": "",
            "rpcpass": ""
        }
    }
    `)
	_, configErr = NewConfig(testConf)
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ERROR_CONFIG_VALUE_NOT_FOUND, RPC_CLIENT_CHAIN_NAME)), configErr)

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
	_, configErr = NewConfig(testConf)
	assert.Equal(t, nil, configErr)

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
	_, configErr = NewConfig(testConf)
	assert.Equal(t, nil, configErr)

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
	_, configErr = NewConfig(testConf)
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ERROR_CONFIG_VALUE_NOT_FOUND, DB_PASSWORD_NAME)), configErr)

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
	_, configErr = NewConfig(testConf)
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ERROR_CONFIG_VALUE_NOT_FOUND, DB_HOST_NAME)), configErr)

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
	_, configErr = NewConfig(testConf)
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ERROR_CONFIG_VALUE_NOT_FOUND, DB_PORT_NAME)), configErr)

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
	_, configErr = NewConfig(testConf)
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ERROR_CONFIG_VALUE_NOT_FOUND, DB_NAME_NAME)), configErr)

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
	_, configErr = NewConfig(testConf)
	assert.Equal(t, nil, configErr)
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
            "signers": "127.0.0.1:12345,127.0.0.1:12346"
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
	assert.Equal(t, []string{"127.0.0.1:12345", "127.0.0.1:12346"}, config.SignerConfig().Signers)
	assert.Equal(t, DbConfig{
		User:     "username1",
		Password: "password2",
		Host:     "localhost",
		Port:     "27017",
		Name:     "mainstay",
	}, config.DbConfig())
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
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ERROR_CONFIG_VALUE_NOT_FOUND, SIGNER_SIGNERS_NAME)), configErr)

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "",
            "rpcuser": "",
            "rpcpass": "",
            "chain": ""
        },
        "signer": {
            "signers": "host"
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, nil, configErr)
	assert.Equal(t, []string{"host"}, config.SignerConfig().Signers)

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "",
            "rpcuser": "",
            "rpcpass": "",
            "chain": ""
        },
        "signer": {
            "signers": "host",
            "publisher": "*:5000"
        }
    }
    `)
	config, configErr = NewConfig(testConf)
	assert.Equal(t, nil, configErr)
	assert.Equal(t, "*:5000", config.SignerConfig().Publisher)
}
