//
// Handle reading conf files and parsing 
// configuration options
//

package conf

import (
    "bytes"
    "encoding/json"
    "log"
    //"os"
    //"path/filepath"
)

var confjson = []byte(`
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

type ClientCfg map[string]interface{}

// Get config for a specific client from conf file
func getCfg(name string) ClientCfg {
    /* // need specific file path
    dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
    file, err := os.Open(dir + "/conf.json")
    if err != nil {
        log.Fatal(err)
    }
    */
    file := bytes.NewReader(confjson)
    dec := json.NewDecoder(file)
    var j map[string]map[string]interface{}
    err := dec.Decode(&j)
    if err != nil {
        log.Fatal(err)
    }
    val, ok := j[name]
    if !ok {
        log.Fatal(err)
    }
    return val
}

// Get string values of config options for a client
func (conf ClientCfg) getValue(key string) string {
    val, ok := conf[key]
    if !ok {
        log.Fatal("%s not found in conf file", key)
    }
    str, ok := val.(string)
    if !ok {
        log.Fatal("%s not string in conf file", key)
    }
    return str
}
