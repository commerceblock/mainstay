// Client RPC connectivity and client related functionality

package conf

import (
    "log"
    "github.com/btcsuite/btcd/rpcclient"
)

var myConf = []byte(`
{
    "main": {
        "rpcurl": "localhost:18000",
        "rpcuser": "user",
        "rpcpass": "pass"
    },
    "ocean": {
        "rpcurl": "localhost:18001",
        "rpcuser": "user",
        "rpcpass": "pass"
    }
}
`)

// Get RPC connection for a client from a conf file
func GetRPC(name string, customConf ...[]byte) *rpcclient.Client{
    conf := myConf
    if len(customConf) > 0 { //custom config provided
        conf = customConf[0]
    }
    cfg := getCfg(name, conf)
    connCfg := &rpcclient.ConnConfig{
        Host:         cfg.getValue("rpcurl"),
        User:         cfg.getValue("rpcuser"),
        Pass:         cfg.getValue("rpcpass"),
        HTTPPostMode: true,
        DisableTLS:   true,
    }
    client, err := rpcclient.New(connCfg, nil)
    if err != nil {
        log.Fatal(err)
    }
    return client
}
