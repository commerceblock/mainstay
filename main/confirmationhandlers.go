package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "github.com/gorilla/mux"
)

func Index(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Confirmation Service for Ocean Attestations!")
}

func Block(w http.ResponseWriter, r *http.Request) {
    blockid := mux.Vars(r)["blockId"]
    req := Request{Name: mux.CurrentRoute(r).GetName(), Id: blockid}

    if err := json.NewEncoder(w).Encode(req); err != nil {
        panic(err)
    }
}
