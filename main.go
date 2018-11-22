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
	tx0        string
	script     string
	isRegtest  bool
	mainConfig *config.Config
)

func parseFlags() {
	flag.BoolVar(&isRegtest, "regtest", false, "Use regtest wallet configuration instead of user wallet")
	flag.StringVar(&tx0, "tx", "", "Tx id for genesis attestation transaction")
	flag.StringVar(&script, "script", "", "Redeem script in case multisig is used")
	flag.Parse()

	if (tx0 == "" || script == "") && !isRegtest {
		flag.PrintDefaults()
		log.Fatalf("Need to provide both -tx and -script argument. To use test configuration set the -regtest flag.")
	}
}

func init() {
	parseFlags()

	if isRegtest {
		test := test.NewTest(true, true)
		mainConfig = test.Config
		log.Printf("Running regtest mode with -tx=%s\n", mainConfig.InitTX())
	} else {
		var mainConfigErr error
		mainConfig, mainConfigErr = config.NewConfig()
		if mainConfigErr != nil {
			log.Fatal(mainConfigErr)
		}
		mainConfig.SetInitTX(tx0)
		mainConfig.SetMultisigScript(script)
	}
}

func main() {
	defer mainConfig.MainClient().Shutdown()

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	dbInterface := server.NewDbMongo(ctx, mainConfig.DbConfig())
	server := server.NewServer(dbInterface)
	attestService := attestation.NewAttestService(ctx, wg, server, mainConfig, isRegtest)

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
