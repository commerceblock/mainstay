package requestapi

import (
	"encoding/json"
	"fmt"
	"mainstay/models"
	"net/http"

	"github.com/gorilla/mux"
)

// Http handlers for service requests

// Is Block Attested request handler
func HandleBlock(w http.ResponseWriter, r *http.Request, channel *models.Channel) {
	blockid := mux.Vars(r)["blockId"]
	request := models.Request{Name: mux.CurrentRoute(r).GetName(), Id: blockid}
	channel.Requests <- request     // put request in channel
	response := <-channel.Responses // wait for response from attestation service
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

// Best Block request handler
func HandleBestBlock(w http.ResponseWriter, r *http.Request, channel *models.Channel) {
	request := models.Request{Name: mux.CurrentRoute(r).GetName()}
	channel.Requests <- request     // put request in channel
	response := <-channel.Responses // wait for response from attestation service
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

// Best Block Height request handler
func HandleBestBlockHeight(w http.ResponseWriter, r *http.Request, channel *models.Channel) {
	request := models.Request{Name: mux.CurrentRoute(r).GetName()}
	channel.Requests <- request     // put request in channel
	response := <-channel.Responses // wait for response from attestation service
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

// Index request handler
func HandleIndex(w http.ResponseWriter, r *http.Request, channel *models.Channel) {
	fmt.Fprintln(w, "Request Service for Ocean Attestations!")
}

// Latest Attestation request handler
func HandleLatestAttestation(w http.ResponseWriter, r *http.Request, channel *models.Channel) {
	request := models.Request{Name: mux.CurrentRoute(r).GetName()}
	channel.Requests <- request     // put request in channel
	response := <-channel.Responses // wait for response from attestation service
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

// Is Transaction Attested request handler
func HandleTransaction(w http.ResponseWriter, r *http.Request, channel *models.Channel) {
	transactionId := mux.Vars(r)["transactionId"]
	request := models.Request{Name: mux.CurrentRoute(r).GetName(), Id: transactionId}
	channel.Requests <- request     // put request in channel
	response := <-channel.Responses // wait for response from attestation service
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}
