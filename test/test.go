// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package test

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"mainstay/clients"
	confpkg "mainstay/config"
	"mainstay/models"
	"mainstay/server"
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
        "signers": "127.0.0.1:5001,127.0.0.1:5002"
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
const Address = "2N8AAQy6SH5HGoAtzwr5xp4LTicqJ3fic8d"
const Script = "512103e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b3321037361a2dba6a9e82faaf5465c36937adba283c878c506000b8479894c6f9cbae752ae"
const InitChaincodes = "1a090f710e47968aee906804f211cf10cde9a11e14908ca0f78cc55dd190ceaa,0a090f710e47968aee906804f211cf10cde9a11e14908ca0f78cc55dd190ceaa"

// pubkey hsm -  "037361a2dba6a9e82faaf5465c36937adba283c878c506000b8479894c6f9cbae7"
// pubkey main - "03e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b33"
const PrivMain = "cQca2KvrBnJJUCYa2tD4RXhiQshWLNMSK2A96ZKWo1SZkHhh3YLz"

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

	// Get transaction for Address as initial TX for attestation chain
	unspent, errUnspent := config.MainClient().ListTransactions("*")
	if errUnspent != nil {
		log.Fatal(errUnspent)
	}
	var txid string
	for _, vout := range unspent {
		if vout.Address == Address {
			txid = vout.TxID
		}
	}

	config.SetInitTx(txid)
	config.SetInitPK(PrivMain)
	config.SetInitScript(Script)
	config.SetInitChaincodes(strings.Split(InitChaincodes, ","))

	// set topupScript-topupAddress same as init script/addr
	config.SetTopupScript(TopupScript)
	config.SetTopupAddress(TopupAddress)
	config.SetTopupPK(TopupPrivMain)

	config.SetRegtest(true)
	return &Test{config, oceanClient}
}

// Work on main client for regtest
// Do block generation automatically
// Do auto commitment for position 0
func DoRegtestWork(dbMongo *server.DbMongo, config *confpkg.Config, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	doCommit := false
	for {
		newBlockTimer := time.NewTimer(60 * time.Second)
		select {
		case <-ctx.Done():
			return
		case <-newBlockTimer.C:
			// generate and get hash
			hash, genErr := config.MainClient().Generate(1)
			if genErr != nil {
				log.Println(genErr)
			}

			// every other block generation commit
			// dummy block hash as commitment for
			// client position 0 in ClientCommitment
			if doCommit {
				newClientCommitment := models.ClientCommitment{
					Commitment:     *hash[0],
					ClientPosition: 0}

				saveErr := dbMongo.SaveClientCommitment(newClientCommitment)
				if saveErr != nil {
					log.Println(saveErr)
				}
				doCommit = false
			} else {
				doCommit = true
			}
		}
	}
}
