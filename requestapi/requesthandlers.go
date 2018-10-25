package requestapi

import (
	"encoding/json"
	"fmt"
	"mainstay/models"
	"net/http"

	"github.com/gorilla/mux"
)

// Http handlers for service requests

// Index request handler
func HandleIndex(w http.ResponseWriter, r *http.Request, channel *models.Channel) {
	fmt.Fprintln(w, "Request Service for Ocean Attestations!")
}

// Is Block Attested request handler
func HandleVerifyBlock(w http.ResponseWriter, r *http.Request, channel *models.Channel) {
	blockid := mux.Vars(r)["blockId"]
	request := models.RequestGet_s{Name: mux.CurrentRoute(r).GetName(), Id: blockid}
	channel.RequestGet <- request   // put request in channel
	response := <-channel.Responses // wait for response from attestation service
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

// Best Block request handler
func HandleBestBlock(w http.ResponseWriter, r *http.Request, channel *models.Channel) {
	request := models.RequestGet_s{Name: mux.CurrentRoute(r).GetName()}
	channel.RequestGet <- request   // put request in channel
	response := <-channel.Responses // wait for response from attestation service
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

// Best Block Height request handler
func HandleBestBlockHeight(w http.ResponseWriter, r *http.Request, channel *models.Channel) {
	request := models.RequestGet_s{Name: mux.CurrentRoute(r).GetName()}
	channel.RequestGet <- request   // put request in channel
	response := <-channel.Responses // wait for response from attestation service
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

// TODO: Add comment
func HandleCommitmentSend(w http.ResponseWriter, r *http.Request, channel *models.Channel) {
	clientid := r.Header.Get("CLIENT-ID")
	hash := r.Header.Get("HASH")
	height := r.Header.Get("HEIGHT")
	request := models.RequestPost_s{Name: mux.CurrentRoute(r).GetName(), ClientId: clientid, Hash: hash, Height: height}
	channel.RequestPost <- request  // put request in channel
	response := <-channel.Responses // wait for response from attestation service
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

// Latest Attestation request handler
func HandleLatestAttestation(w http.ResponseWriter, r *http.Request, channel *models.Channel) {
	request := models.RequestGet_s{Name: mux.CurrentRoute(r).GetName()}
	channel.RequestGet <- request   // put request in channel
	response := <-channel.Responses // wait for response from attestation service
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}
