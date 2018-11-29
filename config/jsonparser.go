// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package config

import (
	"bytes"
	"encoding/json"
	"errors"
)

// Handle reading conf files and parsing configuration options

const (
	ERROR_CONFIG_NAME_NOT_FOUND  = "config name not found"
	ERROR_CONFIG_VALUE_NOT_FOUND = "config value not found"
)

type ClientCfg map[string]interface{}

// Get config for a specific base name from conf file
func getCfg(name string, conf []byte) (ClientCfg, error) {
	file := bytes.NewReader(conf)
	dec := json.NewDecoder(file)
	var j map[string]map[string]interface{}
	err := dec.Decode(&j)
	if err != nil {
		return ClientCfg{}, errors.New(ERROR_CONFIG_NAME_NOT_FOUND)
	}
	val, ok := j[name]
	if !ok {
		return ClientCfg{}, errors.New(ERROR_CONFIG_NAME_NOT_FOUND)
	}
	return val, nil
}

// Get string values of config options for a base category
func (conf ClientCfg) getValue(key string) (string, error) {
	val, ok := conf[key]
	if !ok {
		return "", errors.New(ERROR_CONFIG_VALUE_NOT_FOUND)
	}
	str, ok := val.(string)
	if !ok {
		return "", errors.New(ERROR_CONFIG_VALUE_NOT_FOUND)
	}
	return str, nil
}

// Try get string values of config options for a base category
func (conf ClientCfg) tryGetValue(key string) string {
	val, ok := conf[key]
	if !ok {
		return ""
	}
	str, ok := val.(string)
	if !ok {
		return ""
	}
	return str
}
