package requestapi

import (
	"log"
	"mainstay/models"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const ROUTE_BEST_BLOCK = "BestBlock"
const ROUTE_BEST_BLOCK_HEIGHT = "BestBlockHeight"
const ROUTE_BLOCK = "Block"
const ROUTE_COMMITMENT_SEND = "CommitmentSend"
const ROUTE_INDEX = "Index"
const ROUTE_LATEST_ATTESTATION = "LatestAttestation"
const ROUTE_SERVER_VERIFY = "HandleServerVerify"
const ROUTE_TRANSACTION = "Transaction"

// Route structure
// Routing for http requests to request service
type Route struct {
	name        string
	method      string
	pattern     string
	handlerFunc func(http.ResponseWriter, *http.Request, *models.Channel)
}

var routes = []Route{
	Route{
		ROUTE_BEST_BLOCK,
		"GET",
		"/api/bestblock/",
		HandleBestBlock,
	},
	Route{
		ROUTE_BEST_BLOCK_HEIGHT,
		"GET",
		"/api/bestblockheight/",
		HandleBestBlockHeight,
	},
	Route{
		ROUTE_BLOCK,
		"GET",
		"/api/block/{blockId}",
		HandleBlock,
	},
	// Route{
	// 	ROUTE_COMMITMENT_SEND,
	// 	"POST",
	// 	"/api/commitment/send/{clientId,hash,height}",
	// 	HandleCommitmentSend,
	// },
	Route{
		ROUTE_INDEX,
		"GET",
		"/api/",
		HandleIndex,
	},
	Route{
		ROUTE_LATEST_ATTESTATION,
		"GET",
		"/api/latestattestation/",
		HandleLatestAttestation,
	},
	Route{
		ROUTE_SERVER_VERIFY,
		"GET",
		"/api/server/verify/{hash}",
		HandleServerVerify,
	},
	Route{
		ROUTE_TRANSACTION,
		"GET",
		"/api/transaction/{transactionId}",
		HandleTransaction,
	},
}

// NewRouter returns pointer to mux router instance
func NewRouter(channel *models.Channel) *mux.Router {
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
func makeHandler(fn func(http.ResponseWriter, *http.Request, *models.Channel), channel *models.Channel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		fn(w, r, channel)
		log.Printf("%s\t%s\t%s", r.Method, r.RequestURI, time.Since(start))
	}
}
