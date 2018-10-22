package clients

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
)

// SidechainClientOcean structure
// Ocean implementation for the sidechain client interface
type SidechainClientOcean struct {
	rpc *rpcclient.Client
}

// NewSidechainClientOcean returns new instance of SideChainClient for Ocean
func NewSidechainClientOcean(rpc *rpcclient.Client) *SidechainClientOcean {
	return &SidechainClientOcean{rpc}
}

// Close function shuts down the rpc connection to Ocean
func (o *SidechainClientOcean) Close() {
	o.rpc.Shutdown()
	return
}

// GetBlockCount Ocean implementation using underlying rpc client
func (o *SidechainClientOcean) GetBlockCount() (int64, error) {
	blockcount, err := o.rpc.GetBlockCount()
	if err != nil {
		return -1, err
	}
	return blockcount, nil
}

// GetBestBlockHash Ocean implementation using underlying rpc client
func (o *SidechainClientOcean) GetBestBlockHash() (*chainhash.Hash, error) {
	latesthash, err := o.rpc.GetBestBlockHash()
	if err != nil {
		return nil, err
	}
	return latesthash, nil
}

// GetBlockHeight Ocean implementation using underlying rpc client
func (o *SidechainClientOcean) GetBlockHeight(hash *chainhash.Hash) (int32, error) {
	latestheader, err := o.rpc.GetBlockHeaderVerbose(hash)
	if err != nil {
		return -1, err
	}
	return latestheader.Height, nil
}

// GetBlockHash Ocean implementation using underlying rpc client
func (o *SidechainClientOcean) GetBlockHash(height int64) (*chainhash.Hash, error) {
	hash, err := o.rpc.GetBlockHash(int64(height))
	if err != nil {
		return nil, err
	}
	return hash, nil
}

// GetTxBlockHash Ocean implementation using underlying rpc client
func (o *SidechainClientOcean) GetTxBlockHash(hash *chainhash.Hash) (string, error) {
	tx, err := o.rpc.GetRawTransactionVerbose(hash)
	if err != nil {
		return "", err
	}
	return tx.BlockHash, nil
}
