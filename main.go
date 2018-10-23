// Package main implements attestation and request services.
package main

import (
	"context"
	"flag"
	"log"
	"mainstay/attestation"
	"mainstay/config"
	"mainstay/requestapi"
	"mainstay/server"
	"mainstay/test"
	"os"
	"os/signal"
	"sync"
	"time"
)

const DEFAULT_API_HOST = "localhost:8080"

var (
	tx0        string
	pk0        string
	script     string
	isRegtest  bool
	apiHost    string
	mainConfig *config.Config
)

func parseFlags() {
	flag.BoolVar(&isRegtest, "regtest", false, "Use regtest wallet configuration instead of user wallet")
	flag.StringVar(&tx0, "tx", "", "Tx id for genesis attestation transaction")
	flag.StringVar(&pk0, "pk", "", "Main client pk for genesis attestation transaction")
	flag.StringVar(&script, "script", "", "Redeem script in case multisig is used")
	flag.Parse()

	if (tx0 == "" || pk0 == "") && !isRegtest {
		flag.PrintDefaults()
		log.Fatalf("Need to provide both -tx and -pk argument. To use test configuration set the -regtest flag.")
	}
}

func init() {
	parseFlags()

	if isRegtest {
		test := test.NewTest(true, true)
		mainConfig = test.Config
		log.Printf("Running regtest mode with -tx=%s\n", mainConfig.InitTX())
	} else {
		mainConfig = config.NewConfig(false)
		mainConfig.SetInitTX(tx0)
		mainConfig.SetInitPK(pk0)
		mainConfig.SetMultisigScript(script)
	}

	apiHost = os.Getenv("API_HOST")
	if apiHost == "" {
		apiHost = DEFAULT_API_HOST
	}
}

func main() {
	defer mainConfig.MainClient().Shutdown()
	defer mainConfig.OceanClient().Close()

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	server := server.NewServer(ctx, wg, mainConfig)
	attestService := attestation.NewAttestService(ctx, wg, server, mainConfig)
	requestService := requestapi.NewRequestService(ctx, wg, server, apiHost)

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
	go server.Run()
	wg.Add(1)
	go attestService.Run()
	wg.Add(1)
	go requestService.Run()

	if isRegtest { // In regtest demo mode generate main client blocks automatically
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				newBlockTimer := time.NewTimer(60 * time.Second)
				select {
				case <-ctx.Done():
					return
				case <-newBlockTimer.C:
					mainConfig.MainClient().Generate(1)
				}
			}
		}()
	}
	wg.Wait()
}
