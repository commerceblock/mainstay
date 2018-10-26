package requestapi

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const GET = "GET"
const POST = "POST"

const (
	ROUTE_INDEX              = "/"
	ROUTE_VERIFY_BLOCK       = "/api/verifyblock/{blockId}"
	ROUTE_BEST_BLOCK         = "/api/bestblock/"
	ROUTE_BEST_BLOCK_HEIGHT  = "/api/bestblockheight/"
	ROUTE_COMMITMENT_SEND    = "/api/commitment/send/"
	ROUTE_LATEST_ATTESTATION = "/api/latestattestation/"
)

// Route structure
// Routing for http requests to request service
type Route struct {
	name        string
	method      string
	pattern     string
	handlerFunc func(http.ResponseWriter, *http.Request, *Channel)
}

var routes = []Route{
	Route{
		SERVER_INDEX,
		GET,
		ROUTE_INDEX,
		HandleIndex,
	},
	Route{
		SERVER_BEST_BLOCK,
		GET,
		ROUTE_BEST_BLOCK,
		HandleBestBlock,
	},
	Route{
		SERVER_BEST_BLOCK_HEIGHT,
		GET,
		ROUTE_BEST_BLOCK_HEIGHT,
		HandleBestBlockHeight,
	},
	Route{
		SERVER_VERIFY_BLOCK,
		GET,
		ROUTE_VERIFY_BLOCK,
		HandleVerifyBlock,
	},
	Route{
		SERVER_LATEST_ATTESTATION,
		GET,
		ROUTE_LATEST_ATTESTATION,
		HandleLatestAttestation,
	},
	Route{
		SERVER_COMMITMENT_SEND,
		POST,
		ROUTE_COMMITMENT_SEND,
		HandleCommitmentSend,
	},
}

// NewRouter returns pointer to mux router instance
func NewRouter(channel *Channel) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		handlerFunc := makeHandler(route.handlerFunc, channel) // pass channel to request handler
		router.
			Methods(route.method).
			Path(route.pattern).
			Name(route.name).
			Handler(handlerFunc)
	}
	return router
}

// make custom handler to pass communication channel between server and api
func makeHandler(fn func(http.ResponseWriter, *http.Request, *Channel), channel *Channel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		fn(w, r, channel)
		log.Printf("%s\t%s\t%s", r.Method, r.RequestURI, time.Since(start))
	}
}
