// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// Package main implements attestation and request services.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"

	"mainstay/attestation"
	"mainstay/config"
	"mainstay/server"
	"mainstay/test"
)

var (
	tx0         string
	script0     string
	txTopup     string
	scriptTopup string
	isRegtest   bool
	mainConfig  *config.Config
)

func parseFlags() {
	flag.BoolVar(&isRegtest, "regtest", false, "Use regtest wallet configuration instead of user wallet")
	flag.StringVar(&tx0, "tx", "", "Tx id for genesis attestation transaction")
	flag.StringVar(&script0, "script", "", "Redeem script in case multisig is used")
	flag.StringVar(&txTopup, "txTopup", "", "Tx id for topup transaction")
	flag.StringVar(&scriptTopup, "scriptTopup", "", "Redeem script for topup")
	flag.Parse()
}

func init() {
	parseFlags()

	if isRegtest {
		test := test.NewTest(true, true)
		mainConfig = test.Config
		log.Printf("Running regtest mode with -tx=%s\n", mainConfig.InitTx())
	} else {
		var mainConfigErr error
		mainConfig, mainConfigErr = config.NewConfig()
		if mainConfigErr != nil {
			log.Fatal(mainConfigErr)
		}

		// if either tx or script not set throw error
		if tx0 == "" || script0 == "" {
			if mainConfig.InitTx() == "" || mainConfig.InitScript() == "" {
				flag.PrintDefaults()
				log.Fatalf(`Need to provide both -tx and -script argument.
                    To use test configuration set the -regtest flag.`)
			}
		} else {
			mainConfig.SetInitTx(tx0)
			mainConfig.SetInitScript(script0)
		}
		if txTopup != "" && scriptTopup != "" {
			mainConfig.SetTopupTx(txTopup)
			mainConfig.SetTopupScript(scriptTopup)
		}
		mainConfig.SetRegtest(isRegtest)
	}
}

func main() {
	defer mainConfig.MainClient().Shutdown()

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	dbInterface := server.NewDbMongo(ctx, mainConfig.DbConfig())
	server := server.NewServer(dbInterface)
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
			log.Printf("Got %s signal. Aborting...\n", sig)
		case <-ctx.Done():
			signal.Stop(c)
		}
	}()

	wg.Add(1)
	go attestService.Run()

	if isRegtest { // In regtest demo mode do block generation work
		wg.Add(1)
		go test.DoRegtestWork(mainConfig, wg, ctx)
	}
	wg.Wait()
}
