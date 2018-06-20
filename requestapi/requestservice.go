// Request service routine

package requestapi

import (
    "log"
    "net/http"
    "sync"
    "context"
    "time"
    "github.com/gorilla/mux"
    "ocean-attestation/models"
)

type RequestService struct {
    ctx     context.Context
    wg      *sync.WaitGroup
    host    string
    router  *mux.Router
}

func NewRequestService(ctx context.Context, wg *sync.WaitGroup, channel *models.Channel, host string) *RequestService {
    router := NewRouter(channel)
    return &RequestService{ctx, wg, host, router}
}

func (c *RequestService) Run() {
    defer c.wg.Done()

    srv := &http.Server{
        Addr:   c.host,
        WriteTimeout: time.Second * 15,
        ReadTimeout:  time.Second * 15,
        Handler: c.router,
    }

    c.wg.Add(1)
    go func() { //Running server waiting for requests
        defer c.wg.Done()
        if err := srv.ListenAndServe(); err != nil {
            log.Println(err)
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
