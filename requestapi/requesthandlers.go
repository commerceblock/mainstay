package requestapi

import (
	"encoding/json"
	"fmt"
	"net/http"
    "log"

	"github.com/gorilla/mux"
)

// Http handlers for service requests

// Index request handler
func HandleIndex(w http.ResponseWriter, r *http.Request, channel *Channel) {
	fmt.Fprintln(w, "Request Service for Ocean Attestations!")
}

// Is Block Attested request handler
func HandleVerifyBlock(w http.ResponseWriter, r *http.Request, channel *Channel) {
	blockid := mux.Vars(r)["blockId"]

	request := ServerVerifyBlockRequest{Id: blockid}
    request.SetRequestType(mux.CurrentRoute(r).GetName())

	channel.Requests <- request   // put request in channel
	response := <-channel.Responses // wait for response from attestation service
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

// Best Block request handler
func HandleBestBlock(w http.ResponseWriter, r *http.Request, channel *Channel) {
	request := BaseRequest{}
    request.SetRequestType(mux.CurrentRoute(r).GetName())

	channel.Requests <- request   // put request in channel
	response := <-channel.Responses // wait for response from attestation service
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

// Best Block Height request handler
func HandleBestBlockHeight(w http.ResponseWriter, r *http.Request, channel *Channel) {
	request := BaseRequest{}
    request.SetRequestType(mux.CurrentRoute(r).GetName())

	channel.Requests <- request   // put request in channel
	response := <-channel.Responses // wait for response from attestation service
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

// TODO: Add comment
func HandleCommitmentSend(w http.ResponseWriter, r *http.Request, channel *Channel) {
	clientid := r.Header.Get("CLIENT-ID")
	hash := r.Header.Get("HASH")
	height := r.Header.Get("HEIGHT")

	request := ServerCommitmentSendRequest{ClientId: clientid, Hash: hash, Height: height}
    request.SetRequestType(mux.CurrentRoute(r).GetName())

	channel.Requests <- request  // put request in channel
	response := <-channel.Responses // wait for response from attestation service
    log.Printf("%v\n", response)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

// Latest Attestation request handler
func HandleLatestAttestation(w http.ResponseWriter, r *http.Request, channel *Channel) {
	request := BaseRequest{}
    request.SetRequestType(mux.CurrentRoute(r).GetName())

	channel.Requests <- request   // put request in channel
	response := <-channel.Responses // wait for response from attestation service
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}
