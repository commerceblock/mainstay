// Http handlers for service requests

package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "github.com/gorilla/mux"
)

func Index(w http.ResponseWriter, r *http.Request, channel *Channel) {
    fmt.Fprintln(w, "Request Service for Ocean Attestations!")
}

func Block(w http.ResponseWriter, r *http.Request, channel *Channel) {
    blockid := mux.Vars(r)["blockId"]
    request := Request{Name: mux.CurrentRoute(r).GetName(), Id: blockid,}

    channel.requests <- request // put request in channel

    response := <- channel.responses // wait for response from attestation service

    if err := json.NewEncoder(w).Encode(response); err != nil {
        panic(err)
    }
}

func BestBlock(w http.ResponseWriter, r *http.Request, channel *Channel) {
    request := Request{Name: mux.CurrentRoute(r).GetName(),}

    channel.requests <- request // put request in channel

    response := <- channel.responses // wait for response from attestation service

    if err := json.NewEncoder(w).Encode(response); err != nil {
        panic(err)
    }
}

func Transaction(w http.ResponseWriter, r *http.Request, channel *Channel) {
    transactionId := mux.Vars(r)["transactionId"]
    request := Request{Name: mux.CurrentRoute(r).GetName(), Id: transactionId,}

    channel.requests <- request // put request in channel

    response := <- channel.responses // wait for response from attestation service

    if err := json.NewEncoder(w).Encode(response); err != nil {
        panic(err)
    }
}

func LatestAttestation(w http.ResponseWriter, r *http.Request, channel *Channel) {
    request := Request{Name: mux.CurrentRoute(r).GetName(),}

    channel.requests <- request // put request in channel

    response := <- channel.responses // wait for response from attestation service

    if err := json.NewEncoder(w).Encode(response); err != nil {
        panic(err)
    }
}
