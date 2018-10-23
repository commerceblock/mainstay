package requestapi

import (
	"log"
	"mainstay/models"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

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
		models.ROUTE_BEST_BLOCK,
		"GET",
		"/api/bestblock/",
		HandleBestBlock,
	},
	Route{
		models.ROUTE_BEST_BLOCK_HEIGHT,
		"GET",
		"/api/bestblockheight/",
		HandleBestBlockHeight,
	},
	Route{
		models.ROUTE_BLOCK,
		"GET",
		"/api/block/{blockId}",
		HandleBlock,
	},
	Route{
		models.ROUTE_COMMITMENT_SEND,
		"POST",
		"/api/commitment/send/{clientId,hash,height}",
		HandleCommitmentSend,
	},
	Route{
		models.ROUTE_INDEX,
		"GET",
		"/api/",
		HandleIndex,
	},
	Route{
		models.ROUTE_LATEST_ATTESTATION,
		"GET",
		"/api/latestattestation/",
		HandleLatestAttestation,
	},
	Route{
		models.ROUTE_SERVER_VERIFY,
		"GET",
		"/api/server/verify/{hash}",
		HandleServerVerify,
	},
	Route{
		models.ROUTE_TRANSACTION,
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
