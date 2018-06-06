package main

import (
	"log"
	"time"
	"sync"
	"context"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/davecgh/go-spew/spew"
)

type Server struct {
	ctx			context.Context
	wg     		*sync.WaitGroup
	serverRpc 	*rpcclient.Client
	txid0		string
	pk0			string
}

func NewServer(ctx context.Context, wg *sync.WaitGroup, rpc *rpcclient.Client, tx string, pk string) *Server{
	if (len(tx) != 64) {
		log.Fatal("Incorrect txid size")
	}
	return &Server{ctx, wg, rpc, tx, pk}
}

func (s *Server) Run() {
	defer s.wg.Done()

	for {
		select {
		    case <-s.ctx.Done():
            	log.Println("Shutting down server...")
		        return
	        default:
	        	s.doAttestation()
	        	time.Sleep(time.Millisecond * 1000)
	    }
	}
}

func (s *Server) findLastUnspent() (bool, btcjson.ListUnspentResult) {
	// Get the latest unspent and generate a new transaction
	unspent, err := s.serverRpc.ListUnspent()
	if err != nil {
		log.Fatal(err)
	}
	if (len(unspent) > 0) {
		tx := unspent[0]
		if (tx.TxID == s.txid0) { // genesis transaction
			log.Printf("found genesis transaction")
			log.Printf("First utxo:\n%v", spew.Sdump(tx))
			return true, tx
		}
	}
	return false, btcjson.ListUnspentResult{}
}

func (s *Server) doAttestation(){
	log.Println("Doing attestation...")
	success, tx := s.findLastUnspent()
	if (success) {
		newTransaction(tx, s.serverRpc)

		for {
			// wait for confirmation
			time.Sleep(time.Millisecond * 1000)
			return
		}
	}
}
