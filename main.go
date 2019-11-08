// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// Package main implements attestation and request services.
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"sync"

	"mainstay/attestation"
	"mainstay/config"
	"mainstay/db"
	"mainstay/test"
	"mainstay/log"
)

var (
	tx0         string
	script0     string
	chaincodes  string
	addrTopup   string
	scriptTopup string
	isRegtest   bool
	mainConfig  *config.Config
)

func parseFlags() {
	flag.BoolVar(&isRegtest, "regtest", false, "Use regtest wallet configuration instead of user wallet")
	flag.StringVar(&tx0, "tx", "", "Tx id for genesis attestation transaction")
	flag.StringVar(&script0, "script", "", "Redeem script in case multisig is used")
	flag.StringVar(&chaincodes, "chaincodes", "", "Chaincodes for multisig pubkeys")
	flag.StringVar(&addrTopup, "addrTopup", "", "Address for topup transaction")
	flag.StringVar(&scriptTopup, "scriptTopup", "", "Redeem script for topup")
	flag.Parse()
}

func init() {
	parseFlags()

	if isRegtest {
		test := test.NewTest(true, true)
		mainConfig = test.Config
		log.Infof("Running regtest mode with -tx=%s\n", mainConfig.InitTx())
	} else {
		var mainConfigErr error
		mainConfig, mainConfigErr = config.NewConfig()
		if mainConfigErr != nil {
			log.Error(mainConfigErr)
		}

		// if either tx or script not set throw error
		if tx0 == "" || script0 == "" || chaincodes == "" {
			if mainConfig.InitTx() == "" || mainConfig.InitScript() == "" || len(mainConfig.InitChaincodes()) == 0 {
				flag.PrintDefaults()
				log.Error(`Need to provide all -tx, -script and -chaincode arguments.
                    To use test configuration set the -regtest flag.`)
			}
		} else {
			mainConfig.SetInitTx(tx0)
			mainConfig.SetInitScript(script0)

			chaincodesList := strings.Split(chaincodes, ",") // string to string slice
			for i := range chaincodesList {                  // trim whitespace
				chaincodesList[i] = strings.TrimSpace(chaincodesList[i])
			}
			mainConfig.SetInitChaincodes(chaincodesList)
		}
		if addrTopup != "" && scriptTopup != "" {
			mainConfig.SetTopupAddress(addrTopup)
			mainConfig.SetTopupScript(scriptTopup)
		}
		mainConfig.SetRegtest(isRegtest)
	}
}

func main() {
	defer mainConfig.MainClient().Shutdown()

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	dbInterface := db.NewDbMongo(ctx, mainConfig.DbConfig())
	server := attestation.NewAttestServer(dbInterface)
	signer := attestation.NewAttestSignerZmq(mainConfig.SignerConfig())
	attestService := attestation.NewAttestService(ctx, wg, server, signer, mainConfig)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	wg.Add(1)
	go func() {
		defer cancel()
		defer wg.Done()
		select {
		case sig := <-c:
			log.Warnf("Got %s signal. Aborting...\n", sig)
		case <-ctx.Done():
			signal.Stop(c)
		}
	}()

	wg.Add(1)
	go attestService.Run()

	// In regtest demo mode do block generation work
	// Also auto commitment to ClientCommitment to
	// allow easier testing without db intervention
	if isRegtest {
		wg.Add(1)
		go test.DoRegtestWork(dbInterface, mainConfig, wg, ctx)
	}
	wg.Wait()
}
