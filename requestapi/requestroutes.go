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
		"Index",
		"GET",
		"/api/",
		Index,
	},
	Route{
		"Block",
		"GET",
		"/api/block/{blockId}",
		Block,
	},
	Route{
		"BestBlock",
		"GET",
		"/api/bestblock/",
		BestBlock,
	},
	Route{
		"BestBlockHeight",
		"GET",
		"/api/bestblockheight/",
		BestBlockHeight,
	},
	Route{
		"Transaction",
		"GET",
		"/api/transaction/{transactionId}",
		Transaction,
	},
	Route{
		"LatestAttestation",
		"GET",
		"/api/latestattestation/",
		LatestAttestation,
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
