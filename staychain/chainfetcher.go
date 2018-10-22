package staychain

import (
	"log"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
)

// ChainFetcher struct
// Struct that handles fetching transactions of the attestation
// chain by searching each main client block and trying to match
// the vin of each transaction with the vout of the previous found
type ChainFetcher struct {
	mainClient   *rpcclient.Client
	txid0        string
	latestTx     Tx
	latestHeight int64
}

// Get initial tx from main client and return fetcher instance
func NewChainFetcher(main *rpcclient.Client, tx Tx) ChainFetcher {

	blockhash, _ := chainhash.NewHashFromStr(tx.BlockHash)
	blockheader, _ := main.GetBlockHeaderVerbose(blockhash)

	return ChainFetcher{main, tx.Txid, tx, int64(blockheader.Height)}
}

// Main method that tries to fetch the next transaction in the chan
// and updates the latest main client block height that was tested
func (f *ChainFetcher) Fetch() []Tx {
	blockcount, errCount := f.mainClient.GetBlockCount()
	if errCount != nil {
		log.Fatal(errCount)
	}

	height := f.latestHeight
	for height < blockcount { // iterate through all blocks until latest
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
		if tx.TxIn[0].PreviousOutPoint.Hash.String() == f.latestTx.Txid {
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
