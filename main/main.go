package main

import (
	"log"
	"context"
	"ocean-attestation/conf"
	"github.com/btcsuite/btcd/rpcclient"
	"os"
	"os/signal"
	"sync"
)

var txid0 string
var pk0 string
var mainClient *rpcclient.Client
var oceanClient *rpcclient.Client
var server *Server

func initialise() {
    args := os.Args[1:]
    if (len(args) != 2) {
    	log.Fatal("Invalid args - Provide txid0 and pk0")
    }
    txid0 = args[0]
    pk0 = args[1]

	mainClient 	= conf.GetRPC("main")
	oceanClient = conf.GetRPC("ocean")
}

func main() {
	initialise()
	defer mainClient.Shutdown()
	defer oceanClient.Shutdown()

	wg := &sync.WaitGroup{}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	server := NewServer(ctx, wg, mainClient, txid0, pk0)

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
