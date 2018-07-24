package conf

import (
    "bytes"
    "encoding/json"
    "log"
)

// Handle reading conf files and parsing configuration options
type ClientCfg map[string]interface{}

// Get config for a specific client from conf file
func getCfg(name string, conf []byte) ClientCfg {
    file := bytes.NewReader(conf)
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
        log.Fatalf("%s not found in conf file", key)
    }
    str, ok := val.(string)
    if !ok {
        log.Fatalf("%s not string in conf file", key)
    }
    return str
}
