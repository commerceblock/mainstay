// Package main implements attestation and request services.
package main

import (
	"context"
	"flag"
    "os"
    "os/signal"
    "sync"
    "log"
    "time"

    "ocean-attestation/attestation"
    "ocean-attestation/config"
    "ocean-attestation/models"
    "ocean-attestation/requestapi"
    "ocean-attestation/test"
)

const DEFAULT_API_HOST = "localhost:8080"

var (
	genesisTX               string
    genesisPK               string
	isRegtest               bool
    apiHost                 string
    mainConfig              *config.Config
)

func parseFlags() {
	flag.BoolVar(&isRegtest, "regtest", false, "Use regtest wallet configuration instead of user wallet")
	flag.StringVar(&genesisTX, "tx", "", "Tx id for genesis attestation transaction")
	flag.Parse()

	if (genesisTX == "") && !isRegtest {
		flag.PrintDefaults()
		log.Fatalf("Need to provide -tx argument. To use test configuration set the -regtest flag.")
	}
}

func init() {
	parseFlags()

    if isRegtest {
        test := test.NewTest(true, true)
        mainConfig = test.Config
        genesisTX = test.Tx0hash
        log.Printf("Running regtest mode with -tx=%s\n", genesisTX)
    } else {
        mainConfig = config.NewConfig(false)
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

	channel := models.NewChannel()
	requestService := requestapi.NewRequestService(ctx, wg, channel, apiHost)
	attestService := attestation.NewAttestService(ctx, wg, channel, mainConfig, genesisTX)

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

	wg.Add(1)
	go requestService.Run()
	wg.Add(1)
	go attestService.Run()
	wg.Wait()
}
