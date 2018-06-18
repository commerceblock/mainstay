// Client RPC connectivity and client related functionality

package conf

import (
    "log"
    "github.com/btcsuite/btcd/rpcclient"
)

var myConf = []byte(`
{
    "main": {
        "rpcurl": "localhost:8333",
        "rpcuser": "bitcoinrpc",
        "rpcpass": "bitcoinrpcpass"
    },
    "ocean": {
        "rpcurl": "localhost:18886",
        "rpcuser": "bitcoinrpc",
        "rpcpass": "acc1e7a299bc49449912e235b54dbce5"
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
