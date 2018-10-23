// Package requestapi implements the request api service that listens to attestation requests.
package requestapi

import (
	"context"
	"log"
	"mainstay/models"
	"mainstay/server"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// RequestService struct
// Handles setting a request router and handling api requests
type RequestService struct {
	ctx     context.Context
	wg      *sync.WaitGroup
	host    string
	router  *mux.Router
	server  *server.Server
	channel *models.Channel
}

// NewRequestService returns a pointer to a RequestService instance
func NewRequestService(ctx context.Context, wg *sync.WaitGroup, server *server.Server, host string) *RequestService {
	channel := models.NewChannel()
	router := NewRouter(channel)
	return &RequestService{ctx, wg, host, router, server, channel}
}

// Main Run method
func (c *RequestService) Run() {
	defer c.wg.Done()

	srv := &http.Server{
		Addr:         c.host,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		Handler:      c.router,
	}

	c.wg.Add(1)
	go func() { //Running server waiting for requests
		defer c.wg.Done()
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c.wg.Add(1)
	go func() { //Waiting for requests from the request service and pass to server for response
		defer c.wg.Done()
		for {
			select {
			case <-c.ctx.Done():
				return
			case req := <-c.channel.Requests:
				c.server.RequestChan() <- models.RequestWithResponseChannel{req, c.channel.Responses}
			}
		}
	}()

	c.wg.Add(1)
	go func() { //Waiting for cancellation signal to shut down server
		defer c.wg.Done()
		select {
		case <-c.ctx.Done():
			log.Println("Shutting down request service...")
			srv.Shutdown(c.ctx)
			return
		}
	}()
}
