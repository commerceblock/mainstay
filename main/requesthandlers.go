// Http handlers for service requests

package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "github.com/gorilla/mux"
)

func Index(w http.ResponseWriter, r *http.Request, reqs chan Request) {
    fmt.Fprintln(w, "Request Service for Ocean Attestations!")
}

func Block(w http.ResponseWriter, r *http.Request, reqs chan Request) {
    blockid := mux.Vars(r)["blockId"]
    request := Request{Name: mux.CurrentRoute(r).GetName(), Id: blockid,}

    reqs <- request // put request in channel

    response := <- reqs // wait for response from attestation service

    if err := json.NewEncoder(w).Encode(response); err != nil {
        panic(err)
    }
}

func BestBlock(w http.ResponseWriter, r *http.Request, reqs chan Request) {
    request := Request{Name: mux.CurrentRoute(r).GetName(),}

    reqs <- request // put request in channel

    response := <- reqs // wait for response from attestation service

    if err := json.NewEncoder(w).Encode(response); err != nil {
        panic(err)
    }
}
