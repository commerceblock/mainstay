// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package test

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"mainstay/clients"
	confpkg "mainstay/config"
	"mainstay/db"
	"mainstay/log"
	"mainstay/models"
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
        "url": "127.0.0.1:8000"
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
const Address = "2N74sgEvpJRwBZqjYUEXwPfvuoLZnRaF1xJ"
const Script = "512103e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b33210325bf82856a8fdcc7a2c08a933343d2c6332c4c252974d6b09b6232ea4080462652ae"
const InitChaincodes = "14df7ece79e83f0f479a37832d770294014edc6884b0c8bfa2e0aaf51fb00229,14df7ece79e83f0f479a37832d770294014edc6884b0c8bfa2e0aaf51fb00229"

// pubkey hsm -  "0325bf82856a8fdcc7a2c08a933343d2c6332c4c252974d6b09b6232ea40804626"
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
		log.Error(err)
	}
	if logOutput {
		log.Infoln(string(output))
	}

	// if not a regtest, then unittest
	config, configErr := confpkg.NewConfig(testConf)
	if configErr != nil {
		log.Error(configErr)
	}
	oceanClient := confpkg.NewClientFromConfig("ocean", true, testConf)

	// Get transaction for Address as initial TX for attestation chain
	unspent, errUnspent := config.MainClient().ListTransactions("*")
	if errUnspent != nil {
		log.Error(errUnspent)
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

	// custom config for top up process
	config.SetTopupScript(TopupScript)
	config.SetTopupAddress(TopupAddress)
	config.SetTopupPK(TopupPrivMain)

	config.SetRegtest(true)
	return &Test{config, oceanClient}
}

// Work on main client for regtest
// Do block generation automatically
// Do auto commitment for position 0
func DoRegtestWork(dbMongo *db.DbMongo, config *confpkg.Config, wg *sync.WaitGroup, ctx context.Context) {
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
				log.Infoln(genErr)
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
					log.Infoln(saveErr)
				}
				doCommit = false
			} else {
				doCommit = true
			}
		}
	}
}

// For unit-testing
const TestInitPathMulti = "/src/mainstay/test/test-init-multi.sh"

// Test Multi structure
// Set up testing environment for use by regtest demo or unit tests
// Use multiple configs to allow multiple transaction signers for testing
type TestMulti struct {
	Configs     []*confpkg.Config
	OceanClient clients.SidechainClient
}

// test parameters for a 2-3 multisig redeemScript and P2SH address
const AddressMulti = "2N53Hkuyz8gSM2swdAvt7yqzxH8vVCKxgvK"
const ScriptMulti = "522103dfc3e2f3d0a3ebf9265d30a87764206d2ee0198820eee200eee4fb3f18eaac43210375f474311ba6248dc7ea1d4044114ee8e8c9cad3974ce2ae5a44dfaa285f3f372103cf016cd19049437c1cfa241bcf1baac58e22c71cae2dc06cb15259ee2f61bb2b53ae"
const InitChaincodesMulti = "14df7ece79e83f0f479a37832d770294014edc6884b0c8bfa2e0aaf51fb00229,24df7ece79e83f0f479a37832d770294014edc6884b0c8bfa2e0aaf51fb00229,34df7ece79e83f0f479a37832d770294014edc6884b0c8bfa2e0aaf51fb00229"
const PrivsMulti = "cUY3m2QRr8tGypHsY8UdPH7W7QtpZPJEe4CWsv4HoK1721cHKxQx,cNtt35LyNnFTTJGnFC1fpH5FFJLtfXUYGYJM9ZZLpt5Yp5fSPYWV,cRUq5ww43yFUSdTyAWrbXxxr838qr7oHiiQKRnPGDgJjbCRZacjo"

// NewTestMulti returns a pointer to a TestMulti instance
func NewTestMulti() *TestMulti {

	initPath := os.Getenv("GOPATH") + TestInitPathMulti
	cmd := exec.Command("/bin/sh", initPath)
	_, err := cmd.Output()
	if err != nil {
		log.Error(err)
	}

	// get config
	config, configErr := confpkg.NewConfig(testConf)
	if configErr != nil {
		log.Error(configErr)
	}

	// Get transaction for Address as initial TX for attestation chain
	unspent, errUnspent := config.MainClient().ListTransactions("*")
	if errUnspent != nil {
		log.Error(errUnspent)
	}
	var txid string
	for _, vout := range unspent {
		if vout.Address == AddressMulti {
			txid = vout.TxID
		}
	}

	var configs []*confpkg.Config
	chaincodesList := strings.Split(InitChaincodesMulti, ",")
	privsList := strings.Split(PrivsMulti, ",")

	// config for each private key
	for _, priv := range privsList {
		// get config
		config, configErr := confpkg.NewConfig(testConf)
		if configErr != nil {
			log.Error(configErr)
		}

		config.SetInitTx(txid)
		config.SetInitPK(priv)
		config.SetInitScript(ScriptMulti)
		config.SetInitChaincodes(chaincodesList)

		// use same as init for topup for ease
		config.SetTopupScript(ScriptMulti)
		config.SetTopupAddress(AddressMulti)
		config.SetTopupPK(priv)

		config.SetRegtest(true)

		configs = append(configs, config)
	}

	oceanClient := confpkg.NewClientFromConfig("ocean", true, testConf)

	return &TestMulti{configs, oceanClient}
}
