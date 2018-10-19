package clients

import (
	"errors"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// SidechainClientFake structure
// Implements fake implementation of SidechainClient for unit-testing
type SidechainClientFake struct {
	height int32
}

// Fake chain of blocks for testing
var blocks = []string{
	"1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"2a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"3a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"4a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"5a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"6a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"7a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"8a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"9a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"aa39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"ba39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"ca39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"da39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"ea39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"fa39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
}

// Fake txs for each fake block
var blockTxs = [][]string{
	{"195df0a1426b6050368d144e7a7886d654c4ed055cd0dd1280d2d9fc6aedea04"},
	{"295df0a1426b6050368d144e7a7886d654c4ed055cd0dd1280d2d9fc6aedea04"},
	{"395df0a1426b6050368d144e7a7886d654c4ed055cd0dd1280d2d9fc6aedea04"},
	{"495df0a1426b6050368d144e7a7886d654c4ed055cd0dd1280d2d9fc6aedea04"},
	{"595df0a1426b6050368d144e7a7886d654c4ed055cd0dd1280d2d9fc6aedea04"},
	{"695df0a1426b6050368d144e7a7886d654c4ed055cd0dd1280d2d9fc6aedea04"},
	{"795df0a1426b6050368d144e7a7886d654c4ed055cd0dd1280d2d9fc6aedea04"},
	{"895df0a1426b6050368d144e7a7886d654c4ed055cd0dd1280d2d9fc6aedea04"},
	{"995df0a1426b6050368d144e7a7886d654c4ed055cd0dd1280d2d9fc6aedea04"},
	{"a95df0a1426b6050368d144e7a7886d654c4ed055cd0dd1280d2d9fc6aedea04"},
	{"b95df0a1426b6050368d144e7a7886d654c4ed055cd0dd1280d2d9fc6aedea04"},
	{"c95df0a1426b6050368d144e7a7886d654c4ed055cd0dd1280d2d9fc6aedea04"},
	{"d95df0a1426b6050368d144e7a7886d654c4ed055cd0dd1280d2d9fc6aedea04"},
	{"e95df0a1426b6050368d144e7a7886d654c4ed055cd0dd1280d2d9fc6aedea04"},
	{"f95df0a1426b6050368d144e7a7886d654c4ed055cd0dd1280d2d9fc6aedea04"},
}

// SidechainClientFake returns new instance of a fake SidechainClient
func NewSidechainClientFake() *SidechainClientFake {
	return &SidechainClientFake{0}
}

// Close function - inherit - do nothing
func (f *SidechainClientFake) Close() {
	return
}

// Generate function increments the latest height in the fake client
func (f *SidechainClientFake) Generate(height int32) {
	f.height += height
}

// GetBlockCount returns number of blocks in fake client
func (f *SidechainClientFake) GetBlockCount() (int64, error) {
	return int64(len(blocks)), nil
}

// GetBestBlockHash returns latest block from fake blocks
func (f *SidechainClientFake) GetBestBlockHash() (*chainhash.Hash, error) {
	hashstr := blocks[f.height]
	hash, err := chainhash.NewHashFromStr(hashstr)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

// GetBlockHeight returns block height of block from fake blocks
func (f *SidechainClientFake) GetBlockHeight(hash *chainhash.Hash) (int32, error) {
	hashstr := hash.String()

	for i, val := range blocks {
		if val == hashstr {
			return int32(i), nil
		}
	}

	return -1, errors.New("Block not found")
}

// GetBlockHash returns block hash of block from fake blocks
func (f *SidechainClientFake) GetBlockHash(height int64) (*chainhash.Hash, error) {
	hashstr := blocks[height]
	hash, err := chainhash.NewHashFromStr(hashstr)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

// GetTxBlockHash returns block hash of fake block for fake tx
func (f *SidechainClientFake) GetTxBlockHash(hash *chainhash.Hash) (string, error) {
	hashstr := hash.String()

	for i, _ := range blocks {
		for _, valtx := range blockTxs[i] {
			if valtx == hashstr {
				return blocks[i], nil
			}
		}
	}

	return "", errors.New("Tx not found")
}

// GetBlockTxs returns the fake txs for a fake block hash
func (f *SidechainClientFake) GetBlockTxs(hash *chainhash.Hash) ([]string, error) {
	blockheight, err := f.GetBlockHeight(hash)
	if err != nil {
		return []string{}, err
	}

	return blockTxs[blockheight], nil
}
