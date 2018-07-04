// Struct that handles fetching transactions of the attestation
// chain by searching each main client block and trying to match
// the vin of each transaction with the vout of the previous found
package staychain

import (
    "log"
    "github.com/btcsuite/btcd/rpcclient"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
)

// ChainFetcher struct
type ChainFetcher struct {
    mainClient      *rpcclient.Client
    sideClient      *rpcclient.Client
    txid0           string
    latestTx        Tx
    latestHeight    int64
}

// Get initial tx from main client and return fetcher instance
func NewChainFetcher(main *rpcclient.Client, side *rpcclient.Client, txid string) ChainFetcher {
    txhash, errHash := chainhash.NewHashFromStr(txid)
    if errHash != nil {
        log.Fatal("Invalid tx id provided")
    }
    txraw, errGet := main.GetRawTransactionVerbose(txhash)
    if errGet != nil {
        log.Fatal("Inititial transcaction does not exist")
    }

    blockhash, _ := chainhash.NewHashFromStr(txraw.BlockHash)
    blockheader, _ := main.GetBlockHeaderVerbose(blockhash)

    return ChainFetcher{main, side, txid, Tx(*txraw), int64(blockheader.Height)}
}

// Main method that tries to fetch the next transaction in the chan
// and updates the latest main client block height that was tested
func (f *ChainFetcher) Fetch() []Tx {
    blockcount, errCount := f.mainClient.GetBlockCount()
    if errCount != nil {
        log.Fatal(errCount)
    }

    height := f.latestHeight
    for height <= blockcount { // iterate through all blocks until latest
        height += 1
        tx, found := f.txInBlock(height)
        if found { // if next tx found update latest and return
            f.latestHeight = height
            f.latestTx = tx
            return []Tx{tx}
        }
    }
    f.latestHeight = height
    return nil
}

// Search for a transaction in a block in which the vin hash
// matches the hash of the previous transcaction in the chain
func (f *ChainFetcher) txInBlock(height int64) (Tx, bool) {
    // Get block for height specified
    blockhash, errHash := f.mainClient.GetBlockHash(height)
    if errHash != nil {
        log.Fatal(errHash)
    }
    block, errBlock := f.mainClient.GetBlock(blockhash)
    if errBlock != nil {
        log.Fatal(errBlock)
    }

    // Iterate through block transactions searching for the next tx in the chain
    for _, tx := range block.Transactions {
        if tx.TxIn[0].PreviousOutPoint.Hash.String() == f.latestTx.Hash {
            txhash := tx.TxHash()
            txraw, errGet := f.mainClient.GetRawTransactionVerbose(&txhash)
            if errGet != nil {
                log.Fatal(errGet)
            }
            return Tx(*txraw), true
        }
    }
    return Tx{}, false
}
