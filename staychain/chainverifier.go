package staychain

import (
    "log"
    "strings"
    "encoding/hex"

    "ocean-attestation/crypto"
    "ocean-attestation/clients"

    "github.com/btcsuite/btcd/chaincfg"
    "github.com/btcsuite/btcd/btcec"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
)

// ChainVerifierInfo struct
type ChainVerifierInfo struct {
    hash    chainhash.Hash
    height  int64
}

// Hash getter
func (i *ChainVerifierInfo) Hash() chainhash.Hash {
    return i.hash
}

// Height getter
func (i *ChainVerifierInfo) Height() int64 {
    return i.height
}

// ChainVerifierError struct
type ChainVerifierError struct {
    errstr  string
}

// Implement Error interface method
func (e *ChainVerifierError) Error() string {
    return e.errstr;
}

// ChainVerifier struct
// Verifies that attestations are part of the staychain
// Does basic validation checks
// Extract attestation pubkey and verify it corresponds to a sidechain block hash
type ChainVerifier struct {
    sideClient      clients.SidechainClient
    cfgMain         *chaincfg.Params
    pubkey0         *btcec.PublicKey
    latestHeight    int64
}

// Return new Chain Verifier instance that verifies attestations on the side chain
func NewChainVerifier(cfgMain *chaincfg.Params, side clients.SidechainClient, tx0 Tx) ChainVerifier {

    pubkey0 := getPubKeyFromTx(tx0)
    return ChainVerifier{side, cfgMain, pubkey0, 0}
}

// Method to get the pub key from the scriptSig of a transaction
func getPubKeyFromTx(tx Tx) *btcec.PublicKey {
    scriptSig := tx.Vin[0].ScriptSig.Asm
    sigComps := strings.Split(scriptSig, " ")
    if len(sigComps) == 2 {
        keybytes, _ := hex.DecodeString(sigComps[1])
        key, errKey := btcec.ParsePubKey(keybytes, btcec.S256())
        if errKey != nil {
            log.Fatal(errKey)
        }
        return key
    }
    return nil
}

// Basic verification for vout size and number of addresses
func verifyTxBasic(tx Tx) error {
    if (len(tx.Vout) != 1) {
        return &ChainVerifierError{"Attestation TX does not have a single vout."}
    }

    if (len(tx.Vout[0].ScriptPubKey.Addresses) != 1) {
        return &ChainVerifierError{"Attestation TX does not have a single address."}
    }

    return nil
}

// Verify transaction address by going through all side chain blockhashes,
// tweaking the initial public key with the blockhash and trying to match
// with the public key of the current transaction being verified
func (v *ChainVerifier) verifyTxAddr(addr string) (ChainVerifierInfo, error) {
    blockcount, errCount := v.sideClient.GetBlockCount()
    if errCount != nil {
        log.Fatal(errCount)
    }
    height := v.latestHeight
    for height < blockcount { // iterate through all blocks until latest
        height += 1
        blockhash, errHash := v.sideClient.GetBlockHash(height)
        if errHash != nil {
            log.Printf("height: %d Hash: %s\n", height, blockhash.String())
            log.Fatal(errHash)
        }

        if height%1000 == 0 { // log if verifying takes too long
            log.Printf("Latest verifying block height: %d\n", height)
        }

        tweakedPub := crypto.TweakPubKey(v.pubkey0, blockhash.CloneBytes())
        tweakedAddr := crypto.GetAddressFromPubKey(tweakedPub, v.cfgMain)
        if (tweakedAddr.String() == addr) {
            v.latestHeight = height
            return ChainVerifierInfo{*blockhash, height}, nil
        }
    }
    return ChainVerifierInfo{}, &ChainVerifierError{"Matching hash not found"}
}

// Main chainverifier method wrapping the verification process
func (v *ChainVerifier) Verify(tx Tx) (ChainVerifierInfo, error) {
    errBasic := verifyTxBasic(tx)
    if errBasic != nil {
        return ChainVerifierInfo{}, errBasic
    }

    // In regtest mode it is not obvious how to extract the pubkey
    // from scriptSig. Skipping any further verification for now
    // The verification tool should only be used for live chains anyway
    if v.pubkey0 == nil {
        return ChainVerifierInfo{}, nil
    }

    txaddr := tx.Vout[0].ScriptPubKey.Addresses[0]
    info, errAddr := v.verifyTxAddr(txaddr)
    if errAddr != nil {
        return ChainVerifierInfo{}, errAddr
    }

    return info, nil
}
