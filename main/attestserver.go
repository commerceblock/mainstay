// Attestation Server storing attestation subchain and responding to requests

package main

import (
    "log"
    "time"
    "github.com/btcsuite/btcd/rpcclient"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
)

type AttestServer struct {
    latest      *Attestation
    confirmed   []Attestation
    sideClient  *rpcclient.Client
}

func NewAttestServer(rpcSide *rpcclient.Client, latest *Attestation, tx0 string, genesis string) *AttestServer{
    tx0hash, err := chainhash.NewHashFromStr(tx0)
    if err != nil {
        log.Fatal("*AttestServer* Incorrect tx hash for initial transaction")
    }
    genesishash, err := chainhash.NewHashFromStr(genesis)
    if err != nil {
        log.Fatal("*AttestServer* Incorrect genesis hash for client sidechain")
    }
    return &AttestServer{latest, []Attestation{Attestation{*tx0hash, *genesishash, true, time.Now()}}, rpcSide}
}

func (s *AttestServer) AddConfirmed(tx Attestation) {
    s.confirmed = append(s.confirmed, tx)
    log.Printf("**AttestServer** Added confirmed transaction %s\n", tx.txid)
}

func (s *AttestServer) Respond(req Request) Request{
    switch req.Name {
    case "Block":
        return s.BlockResponse(req)
    case "BestBlock":
        return s.BestBlockResponse(req)
    default:
        req.Error = "**AttestServer** Non supported request of type " + req.Name
        return req
    }
}
