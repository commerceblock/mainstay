package requestapi

import (
	"log"
	"mainstay/models"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const GET = "GET"
const POST = "POST"

const ROUTE_BEST_BLOCK = "/api/bestblock/"
const ROUTE_BEST_BLOCK_HEIGHT = "/api/bestblockheight/"
const ROUTE_BLOCK = "/api/block/{blockId}"
const ROUTE_COMMITMENT_SEND = "/api/commitment/send/"
const ROUTE_INDEX = "/api/"
const ROUTE_LATEST_ATTESTATION = "/api/latestattestation/"
const ROUTE_SERVER_VERIFY = "/api/server/verify/{hash}"
const ROUTE_TRANSACTION = "/api/transaction/{transactionId}"

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
		models.NAME_BEST_BLOCK,
		GET,
		ROUTE_BEST_BLOCK,
		HandleBestBlock,
	},
	Route{
		models.NAME_BEST_BLOCK_HEIGHT,
		GET,
		ROUTE_BEST_BLOCK_HEIGHT,
		HandleBestBlockHeight,
	},
	Route{
		models.NAME_BLOCK,
		GET,
		ROUTE_BLOCK,
		HandleBlock,
	},
	Route{
		models.NAME_COMMITMENT_SEND,
		POST,
		ROUTE_COMMITMENT_SEND,
		HandleCommitmentSend,
	},
	Route{
		models.NAME_INDEX,
		GET,
		ROUTE_INDEX,
		HandleIndex,
	},
	Route{
		models.NAME_LATEST_ATTESTATION,
		GET,
		ROUTE_LATEST_ATTESTATION,
		HandleLatestAttestation,
	},
	Route{
		models.NAME_SERVER_VERIFY,
		GET,
		ROUTE_SERVER_VERIFY,
		HandleServerVerify,
	},
	Route{
		models.NAME_TRANSACTION,
		GET,
		ROUTE_TRANSACTION,
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
