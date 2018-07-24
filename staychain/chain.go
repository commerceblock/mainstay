// Package staychain provides utilities for fetching attestations and verifying them.
package staychain

import (
    "log"
    "time"

    "github.com/btcsuite/btcd/btcjson"
)

// Sleep time till next attestation
const SLEEP_TIME = 5*time.Minute

type Tx btcjson.TxRawResult

// Chain structure
// Struct that builds the staychain from the initial transaction,
// adds fetched attestations to a channel on which clients can
// subscribe to and then waits for the next attestation to happen
type Chain struct {
    updates     chan Tx
    closing     chan chan error
    fetcher     ChainFetcher
}

// Return a new Chain instance that continuously fetches attestations
func NewChain(fetcher ChainFetcher) *Chain {
    c := &Chain{
        updates:    make(chan Tx),
        fetcher:    fetcher,
    }
    go c.fetch()
    return c
}

// Return the updates channel for external client use
func (c *Chain) Updates() <-chan Tx {
    return c.updates
}

// Send a closing signal from external client
func (c *Chain) Close() error {
    errc := make(chan error)
    c.closing <- errc
    return <-errc
}

// Fetch chain attestations using c.fetcher and add to updates
func (c *Chain) fetch() {
    var pending []Tx // appended by fetch; consumed by send
    var next time.Time
    var err error
    for {
        var fetchDelay time.Duration // initally 0 (no delay)
        if now := time.Now(); next.After(now) {
            fetchDelay = next.Sub(now)
        }
        startFetch := time.After(fetchDelay)

        var first Tx
        var updates chan Tx
        if len(pending) > 0 {
            first = pending[0]
            updates = c.updates // enable send case
        }

        select {
            case <-startFetch:
                var fetched []Tx
                fetched = c.fetcher.Fetch()
                if len(fetched) == 0 {
                    log.Printf("All attestations fetched. Sleeping for %s...\n", SLEEP_TIME.String())
                    next = time.Now().Add(SLEEP_TIME)
                    break
                }
                pending = append(pending, fetched...)
            case errc := <-c.closing:
                errc <- err
                close(c.updates)
                return
            case updates <- first:
                pending = pending[1:]
        }
    }
}
