package main

import (
	"context"
	"flag"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"log"
	"ocean-attestation/attestation"
	"ocean-attestation/conf"
	"ocean-attestation/models"
	"ocean-attestation/requestapi"
	"ocean-attestation/test"
	"os"
	"os/signal"
	"sync"
)

const API_HOST = "127.0.0.1:8080"
const MAIN_NAME = "main"
const OCEAN_NAME = "ocean"

var (
	genesisTX, genesisPK    string
	mainClient, oceanClient *rpcclient.Client
	mainChainCfg            *chaincfg.Params
	isTest                  bool
)

func parseFlags() {
	flag.BoolVar(&isTest, "test", false, "Use test wallet configuration instead of user wallet")
	flag.StringVar(&genesisTX, "tx", "", "Tx id for genesis attestation transaction")
	flag.StringVar(&genesisPK, "pk", "", "Private key used for genesis attestation transaction")
	flag.Parse()

	if (genesisTX == "" || genesisPK == "") && !isTest {
		flag.PrintDefaults()
		log.Fatalf("Provide both -tx and -pk arguments. To use test configuration set the -test flag.")
	}
}

func init() {
	parseFlags()
	if isTest { // Use configuration applied in unit tests
		test := test.NewTest(true)
		mainClient = test.Btc
		oceanClient = test.Ocean
		mainChainCfg = test.BtcConfig
        genesisTX = test.Tx0hash
        genesisPK = test.Tx0pk
        log.Printf("Running test mode with -tx=%s and -pk=%s\n", genesisTX, genesisPK)
	} else {
		mainClient = conf.GetRPC(MAIN_NAME)
		oceanClient = conf.GetRPC(OCEAN_NAME)
		mainChainCfg = conf.GetChainCfgParams(MAIN_NAME)
	}
}

func main() {
	defer mainClient.Shutdown()
	defer oceanClient.Shutdown()

	wg := &sync.WaitGroup{}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	channel := models.NewChannel()
	requestService := requestapi.NewRequestService(ctx, wg, channel, API_HOST)
	attestService := attestation.NewAttestService(ctx, wg, channel, mainClient, oceanClient, mainChainCfg, genesisTX, genesisPK)

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
	go requestService.Run()
	wg.Add(1)
	go attestService.Run()
	wg.Wait()
}
