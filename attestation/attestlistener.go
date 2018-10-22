package attestation

import (
	"log"
	"mainstay/clients"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type Listener struct {
	clients.SidechainClient
}

func (l *Listener) GetNextHash() chainhash.Hash {
	hash, err := l.SidechainClient.GetBestBlockHash()
	if err != nil {
		log.Fatal(err)
	}
	return *hash
}

func NewListener(client clients.SidechainClient) *Listener {
	return &Listener{client}
}
