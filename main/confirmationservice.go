package main

import (
    "log"
    "net/http"
    "sync"
    "context"
    "github.com/gorilla/mux"
)

type ConfirmationService struct {
    host    string
    ctx     context.Context
    wg      *sync.WaitGroup
    router  *mux.Router
}

func NewConfirmationService(host string, ctx context.Context, wg *sync.WaitGroup) *ConfirmationService {
    router := NewRouter()
    return &ConfirmationService{host, ctx, wg, router}
}

func (c *ConfirmationService) Run() {
    defer c.wg.Done()

    srv := &http.Server{
        Addr:   c.host,
        //WriteTimeout: time.Second * 15,
        //ReadTimeout:  time.Second * 15,
        //IdleTimeout:  time.Second * 60,
        Handler: c.router,
    }

    c.wg.Add(1)
    go func() { //Running server waiting for confirmation requests
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
                log.Println("Shutting down confirmation service...")
                srv.Shutdown(c.ctx)
                return
        }
    }()
}
