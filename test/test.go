// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package test

import (
	"context"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"mainstay/clients"
	confpkg "mainstay/config"

	"github.com/btcsuite/btcd/btcjson"
)

// For regtest attestation demonstration
const DemoInitPath = "/src/mainstay/test/demo-init.sh"

// For unit-testing
const TestInitPath = "/src/mainstay/test/test-init.sh"

var testConf = []byte(`
{
    "main": {
        "rpcurl": "localhost:18443",
        "rpcuser": "user",
        "rpcpass": "pass",
        "chain": "regtest"
    },
    "signer": {
        "signers": "127.0.0.1:5001"
    },
    "db": {
        "user":"serviceUser",
        "password":"servicePass",
        "host":"localhost",
        "port":"27017",
        "name":"mainstayX"
    }
}
`)

// test parameters for a 1-2 multisig redeemScript and P2SH address
const Address = "2MxBi6eodnuoVCw8McGrf1nuoVhastqoBXB"
const Script = "512103e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b332102f3a78a7bd6cf01c56312e7e828bef74134dfb109e59afd088526212d96518e7552ae"

// address - "2N9z6a8BQB1xWmesCJcBWZm1R3f1PZcwrGz"
// pubkey - "03e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b33"
const PrivMain = "cQca2KvrBnJJUCYa2tD4RXhiQshWLNMSK2A96ZKWo1SZkHhh3YLz"

// address - "2MyC1i1FGy6MZWyMgmZXku4gdWZxWCRa6RL"
// pubkey -  "02f3a78a7bd6cf01c56312e7e828bef74134dfb109e59afd088526212d96518e75"
const PrivClient = "cSS9R4XPpajhqy28hcfHEzEzAbyWDqBaGZR4xtV7Jg8TixSWee1x"

// test parameters for a 1-2 multisig redeemScript and P2SH address for TOPUP
const TopupAddress = "2NG3LCHuausgrYEsJQjYqhmgDVjRPYYrB5w"
const TopupScript = "512102253297770861be1e512e00329c91bc85300fa46c39d603320d1f5b5e04eaf3342103a10b8872e1aca43b6c0376a20052efa3789fae2fae82972920b8cbba5bc9f33d52ae"

// address - "2Mvi6msoTtozNPAfuSUtTKCWG8ryMZvheuF"
const TopupPrivMain = "cPLAx2s7x8jBc58Ruyp2dUsG42D5jgY6FzKcSNPiMMeNWw1h6JXX"

// address - "2NGWnYUFHaq6f5KB8qGHcjV3sp6vN5Wc2hu"
const TopupPrivClient = "cUPcXWas6iaCWFKZc2rogeY4JK2cHAtFGS2h9CmfEmL3dzehP8K7"

// Test structure
// Set up testing environment for use by regtest demo or unit tests
type Test struct {
	Config      *confpkg.Config
	OceanClient clients.SidechainClient
}

// NewTest returns a pointer to a Test instance
func NewTest(logOutput bool, isRegtest bool) *Test {
	// Run init test script that sets up bitcoin and ocean
	var initPath string
	if isRegtest { // for running the demon in regtest mode along with ocean demo
		initPath = os.Getenv("GOPATH") + DemoInitPath
	} else { // for running unit tests
		initPath = os.Getenv("GOPATH") + TestInitPath
	}

	cmd := exec.Command("/bin/sh", initPath)
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	if logOutput {
		log.Println(string(output))
	}

	// if not a regtest, then unittest
	config, configErr := confpkg.NewConfig(testConf)
	if configErr != nil {
		log.Fatal(configErr)
	}
	oceanClient := confpkg.NewClientFromConfig("ocean", true, testConf)

	// Get first unspent as initial TX for attestation chain
	unspent, errUnspent := config.MainClient().ListUnspent()
	if errUnspent != nil {
		log.Fatal(errUnspent)
	}
	var tx0 btcjson.ListUnspentResult
	for _, vout := range unspent {
		if vout.Amount > 50 { // skip regtest txs
			tx0 = vout
		}
	}

	config.SetInitTx(tx0.TxID)
	config.SetInitPK(PrivMain)
	config.SetInitScript(Script)

	// set topupScript-topupAddress same as init script/addr
	config.SetTopupScript(TopupScript)
	config.SetTopupAddress(TopupAddress)
	config.SetTopupPK(TopupPrivMain)

	config.SetRegtest(true)
	return &Test{config, oceanClient}
}

// Work on main client for regtest - block generation automatically
func DoRegtestWork(config *confpkg.Config, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	for {
		newBlockTimer := time.NewTimer(60 * time.Second)
		select {
		case <-ctx.Done():
			return
		case <-newBlockTimer.C:
			config.MainClient().Generate(1)
		}
	}
}
