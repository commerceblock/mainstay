//
// Client RPC connectivity and client related functionality
//

package conf

import (
    "log"
    "github.com/btcsuite/btcd/rpcclient"
)

// Get RPC connection for a client from the conf file
func GetRPC(name string) *rpcclient.Client{
    cfg := getCfg(name)
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
