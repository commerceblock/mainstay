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
const Address = "bcrt1q7h6ue5w39ramd4ux6gtxh6swnrefpcfgt7vl64"
const InitChaincodes = "14df7ece79e83f0f479a37832d770294014edc6884b0c8bfa2e0aaf51fb00229,14df7ece79e83f0f479a37832d770294014edc6884b0c8bfa2e0aaf51fb00229"

const PrivMain = "cRb1ZU6gsHeifDnBRyRMfbMayWpnNtpKcNs7Z9XzqE87ZwW6Vqx8"

const PublicKey = "021037eb111561e14bcdfc0676af222041b2e4f4c2ac7309ec4763a1d7e61fa24f"

// test parameters for TOPUP
const TopupAddress = "bcrt1qzkm8f3lu6kljfs875mddl9rs72ses5v49vplrl"

const TopupPrivMain = "cVAAcEr9kGxG7Mo9nZnDsJ2aeQ4DP7588UuJqVAuuk3YT8YU6rEq"

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
	initPath := os.Getenv("GOPATH") + TestInitPath

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
	config.SetInitChaincodes(strings.Split(InitChaincodes, ","))
	config.SetInitPublicKey(PublicKey)

	// custom config for top up process
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
