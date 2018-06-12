package main

import (
    "log"
    "context"
    "flag"
    "ocean-attestation/conf"
    "github.com/btcsuite/btcd/rpcclient"
    "os"
    "os/signal"
    "sync"
)

var (
    txid0, pk0 *string
    confirmHost string
    mainClient, oceanClient *rpcclient.Client
)

func initialise() {
    txid0 = flag.String("tx", "", "Tx id for genesis attestation transaction")
    pk0   = flag.String("pk", "", "Private key used for genesis attestation transaction")
    flag.Parse()

    confirmHost = "127.0.0.1:8080"

    mainClient  = conf.GetRPC("main")
    oceanClient = conf.GetRPC("ocean")
}

func main() {
    initialise()
    defer mainClient.Shutdown()
    defer oceanClient.Shutdown()

    wg := &sync.WaitGroup{}
    ctx := context.Background()
    ctx, cancel := context.WithCancel(ctx)

    requests := make(chan Request)
    confirm := NewConfirmationService(ctx, wg, requests, confirmHost)
    attest := NewAttestService(ctx, wg, requests, mainClient, oceanClient, *txid0, *pk0)

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
    go confirm.Run()
    wg.Add(1)
    go attest.Run()
    wg.Wait()
}
