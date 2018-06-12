package main

import (
    "net/http"
    "github.com/gorilla/mux"
)

type Route struct {
    Name        string
    Method      string
    Pattern     string
    HandlerFunc func(http.ResponseWriter, *http.Request, chan Request)
}

type Routes []Route

var routes = Routes{
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
}

func NewRouter(reqs chan Request) *mux.Router {
    router := mux.NewRouter().StrictSlash(true)
    for _, route := range routes {
        handlerFunc := makeHandler(route.HandlerFunc, reqs)
        router.
            Methods(route.Method).
            Path(route.Pattern).
            Name(route.Name).
            Handler(handlerFunc)
    }
    return router
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, chan Request), reqs chan Request) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        fn(w, r, reqs)
    }
}
