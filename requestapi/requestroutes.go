// Routing for http requests to request service

package requestapi

import (
    "net/http"
    "time"
    "log"
    "github.com/gorilla/mux"
    "ocean-attestation/models"
)

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
        "/",
        Index,
    },
    Route{
        "Block",
        "GET",
        "/block/{blockId}",
        Block,
    },
    Route{
        "BestBlock",
        "GET",
        "/bestblock/",
        BestBlock,
    },
    Route{
        "Transaction",
        "GET",
        "/transaction/{transactionId}",
        Transaction,
    },
    Route{
        "LatestAttestation",
        "GET",
        "/latestattestation/",
        LatestAttestation,
    },
}

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

func makeHandler(fn func (http.ResponseWriter, *http.Request, *models.Channel), channel *models.Channel) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        fn(w, r, channel)
        log.Printf("%s\t%s\t%s", r.Method, r.RequestURI, time.Since(start),)
    }
}
