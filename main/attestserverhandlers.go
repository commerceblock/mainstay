// Response handlers for attestation server

package main

import (
    "log"
    "strconv"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
)

func (s *AttestServer) BlockResponse(req Request) Request{
    // Will store height in the future so no need for this
    currentheader, err := s.sideClient.GetBlockHeaderVerbose(&s.confirmed[len(s.confirmed)-1].clientHash)
    log.Println(&s.confirmed[len(s.confirmed)-1].clientHash)
    if err != nil {
        log.Fatal(err)
    }
    currentheight := currentheader.Height

    var height int32
    if (len(req.Id) == 64) {
        hash, err := chainhash.NewHashFromStr(req.Id)
        if err != nil {
            req.Error = "Invalid request for block hash: " + req.Id
            return req
        }
        header, err := s.sideClient.GetBlockHeaderVerbose(hash)
        log.Println(hash)
        if err != nil {
            req.Error = "Invalid request block does not exist: " + req.Id
            return req
        }
        height = header.Height
    } else {
        h, err := strconv.Atoi(req.Id)
        if err != nil {
            req.Error = "Invalid request for block: " + req.Id
            return req
        }
        height = int32(h)
    }

    req.Attested = currentheight >= height
    return req
}

func (s *AttestServer) BestBlockResponse(req Request) Request{
    req.Id = s.confirmed[len(s.confirmed)-1].clientHash.String()
    req.Attested = true
    return req
}
