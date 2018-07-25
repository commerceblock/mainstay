package clients

import (
    "github.com/btcsuite/btcd/chaincfg/chainhash"
)

// SidechainClient interface
// Implements the interface for sidechain clients
// Current logic includes getting latest block from sidechain
type SidechainClient interface {
    GetBestBlockHash()              (*chainhash.Hash, error)
    GetBlockHeight(*chainhash.Hash) (int32, error)
    GetBlockHash(int64)             (*chainhash.Hash, error)
    GetTxBlockHash(*chainhash.Hash) (string, error)
    GetBlockCount()                 (int64, error)
    Close()
}
