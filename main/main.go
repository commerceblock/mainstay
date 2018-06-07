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
    mainClient, oceanClient *rpcclient.Client
    server *Server
)

func initialise() {
    txid0 = flag.String("tx", "looool", "Tx id for genesis attestation transaction")
    pk0   = flag.String("pk", "", "Private key used for genesis attestation transaction")
    flag.Parse()

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

    server := NewServer(ctx, wg, mainClient, oceanClient, *txid0, *pk0)

    c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt)

    wg.Add(1)
    go func() {
        select {
            case sig := <-c:
                log.Printf("Got %s signal. Aborting...\n", sig)
                cancel()
                wg.Done()
            case <-ctx.Done():
                signal.Stop(c)
                cancel()
                wg.Done()
        }
    }()

    wg.Add(1)
    go server.Run()
    wg.Wait()
}
