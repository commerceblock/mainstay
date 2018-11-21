// Package requestapi implements the request api service that listens to attestation requests.
package requestapi

// import (
// 	"context"
// 	"log"
// 	"net/http"
// 	"sync"
// 	"time"

// 	"github.com/gorilla/mux"
// )

// // RequestService struct
// // Handles setting a request router and handling api requests
// type RequestService struct {
// 	ctx           context.Context
// 	wg            *sync.WaitGroup
// 	host          string
// 	router        *mux.Router
// 	channel       *Channel
// 	serverChannel chan RequestWithResponseChannel
// }

// // NewRequestService returns a pointer to a RequestService instance
// func NewRequestService(ctx context.Context, wg *sync.WaitGroup, serverChannel chan RequestWithResponseChannel, host string) *RequestService {
// 	channel := NewChannel()
// 	router := NewRouter(channel)
// 	return &RequestService{ctx, wg, host, router, channel, serverChannel}
// }

// // Main Run method
// func (c *RequestService) Run() {
// 	defer c.wg.Done()

// 	srv := &http.Server{
// 		Addr:         c.host,
// 		WriteTimeout: time.Second * 15,
// 		ReadTimeout:  time.Second * 15,
// 		Handler:      c.router,
// 	}

// 	c.wg.Add(1)
// 	go func() { //Running server waiting for requests
// 		defer c.wg.Done()
// 		if err := srv.ListenAndServe(); err != nil {
// 			log.Println(err)
// 		}
// 	}()

// 	c.wg.Add(1)
// 	go func() { //Waiting for requests from the request service and pass to server for response
// 		defer c.wg.Done()
// 		for {
// 			select {
// 			case <-c.ctx.Done():
// 				return
// 				// receive requests from http server and pass on to main attestation Server
// 				// provide a response channel along with the request, in order to pick up
// 				// the response and serve it to the clients that sent the request
// 			case req := <-c.channel.Requests:
// 				c.serverChannel <- RequestWithResponseChannel{req, c.channel.Responses}
// 			}
// 		}
// 	}()

// 	c.wg.Add(1)
// 	go func() { //Waiting for cancellation signal to shut down server
// 		defer c.wg.Done()
// 		select {
// 		case <-c.ctx.Done():
// 			log.Println("Shutting down request service...")
// 			srv.Shutdown(c.ctx)
// 			return
// 		}
// 	}()
// }
