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
	assert.Equal(t, errors.New(BAD_DATA_CLIENT_CHAIN), configErr)

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "",
            "rpcuser": "",
            "rpcpass": "",
            "chain": "testnet"
        },
        "misc": {

        }
    }
    `)
	_, configErr = NewConfig(testConf)
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ERROR_CONFIG_VALUE_NOT_FOUND, MISC_MULTISIGNODES_NAME)), configErr)

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "",
            "rpcuser": "",
            "rpcpass": "",
            "chain": "testnet"
        },
        "misc": {
            "multisignodes": "host"
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
        "misc": {
            "multisignodes": "host"
        },
        "db": {
        }
    }
    `)
	_, configErr = NewConfig(testConf)
	assert.Equal(t, errors.New(fmt.Sprintf("%s: %s", ERROR_CONFIG_VALUE_NOT_FOUND, DB_USER_NAME)), configErr)

	testConf = []byte(`
    {
        "main": {
            "rpcurl": "",
            "rpcuser": "",
            "rpcpass": "",
            "chain": "testnet"
        },
        "misc": {
            "multisignodes": "host"
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
        "misc": {
            "multisignodes": "host"
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
        "misc": {
            "multisignodes": "host"
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
        "misc": {
            "multisignodes": "host"
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
        "misc": {
            "multisignodes": "host"
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
        "misc": {
            "multisignodes": "127.0.0.1:12345,127.0.0.1:12346"
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
	assert.Equal(t, []string{"127.0.0.1:12345", "127.0.0.1:12346"}, config.MultisigNodes())
	assert.Equal(t, DbConfig{
		User:     "username1",
		Password: "password2",
		Host:     "localhost",
		Port:     "27017",
		Name:     "mainstay",
	}, config.DbConfig())
}
