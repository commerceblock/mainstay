package attestation

import (
    "log"
    "time"

    "ocean-attestation/crypto"
    "ocean-attestation/clients"

    "github.com/btcsuite/btcd/btcjson"
    "github.com/btcsuite/btcd/rpcclient"
    "github.com/btcsuite/btcutil"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
    "github.com/btcsuite/btcd/chaincfg"
)

// AttestClient structure
// Maintains RPC connections to main and side chain clients
// Handles generating staychain next address and next transaction
// and verifying that the correct chain of transactions is maintained
type AttestClient struct {
    mainClient      *rpcclient.Client
    sideClient      clients.SidechainClient
    mainChainCfg    *chaincfg.Params
    pk0             string
    txid0           string
    walletPriv      *btcutil.WIF
}

// NewAttestClient returns a pointer to a new AttestClient instance
// Initially locates the genesis transaction in the main chain wallet
// and verifies that the corresponding private key is in the wallet
func NewAttestClient(rpcMain *rpcclient.Client, rpcSide clients.SidechainClient, cfgMain *chaincfg.Params, txid string) *AttestClient {
    // Get initial private key from initial funding transaction of main client
    txhash, errHash := chainhash.NewHashFromStr(txid)
    if errHash != nil {
        log.Fatal("Invalid tx id provided")
    }
    tx, errGet := rpcMain.GetTransaction(txhash)
    if errGet != nil {
        log.Fatal("Inititial transcaction does not exist")
    }
    addr, errDec := btcutil.DecodeAddress(tx.Details[0].Address, cfgMain)
    if errDec != nil {
        log.Fatal("Failed decoding address from initial transaction")
    }
    pk, errDump := rpcMain.DumpPrivKey(addr)
    if errDump != nil {
        log.Fatal("Failed getting initial transaction private key from address")
    }

    return &AttestClient{rpcMain, rpcSide, cfgMain, pk.String(), txid, pk}
}

// Get next attestation address by tweaking initial private key with current sidechain block hash
func (w *AttestClient) getNextAttestationAddr() (chainhash.Hash, btcutil.Address) {
    hash, err := w.sideClient.GetBestBlockHash()
    if err != nil {
        log.Fatal(err)
    }

    // Tweak priv key with the latest ocean hash
    tweakedWalletPriv := crypto.TweakPrivKey(w.walletPriv, hash.CloneBytes(), w.mainChainCfg)
    addr := crypto.GetAddressFromPrivKey(tweakedWalletPriv, w.mainChainCfg)

    // Import tweaked priv key to wallet
    importErr := w.mainClient.ImportPrivKeyRescan(tweakedWalletPriv, hash.String(), false)
    if importErr != nil {
        log.Fatal(importErr)
    }

    return *hash, addr
}

// Generate a new transaction paying to the tweaked address, add fees and send the transaction through the wallet client
func (w *AttestClient) sendAttestation(paytoaddr btcutil.Address, txunspent btcjson.ListUnspentResult, useDefaultFee bool) (chainhash.Hash) {
    inputs := []btcjson.TransactionInput{{Txid: txunspent.TxID, Vout: txunspent.Vout},}

    amounts := map[btcutil.Address]btcutil.Amount{paytoaddr: btcutil.Amount(txunspent.Amount*100000000)}
    msgtx, err := w.mainClient.CreateRawTransaction(inputs, amounts, nil)
    if err != nil {
        log.Fatal(err)
    }

    feePerByte := GetFee(useDefaultFee)
    fee := int64(feePerByte * msgtx.SerializeSize())
    msgtx.TxOut[0].Value -= fee

    signedmsgtx, issigned, err := w.mainClient.SignRawTransaction(msgtx)
    if err != nil || !issigned {
        log.Fatal(err)
    }

    txhash, err := w.mainClient.SendRawTransaction(signedmsgtx, false)
    if err != nil {
        log.Fatal(err)
    }
    return *txhash
}

// Verify that an unspent vout is on the tip of the subchain attestations
func (w *AttestClient) verifyTxOnSubchain(txid chainhash.Hash) bool {
    if (txid.String() == w.txid0) { // genesis transaction
        return true
    } else { //might be better to store subchain on init and no need to parse all transactions every time
        txraw, err := w.mainClient.GetRawTransaction(&txid)
        if err != nil {
            return false
        }

        prevtxid := txraw.MsgTx().TxIn[0].PreviousOutPoint.Hash
        return w.verifyTxOnSubchain(prevtxid)
    }
    return false
}

// Find the latest unspent vout that is on the tip of subchain attestations
func (w *AttestClient) findLastUnspent() (bool, btcjson.ListUnspentResult) {
    unspent, err := w.mainClient.ListUnspent()
    if err != nil {
        log.Fatal(err)
    }
    if (len(unspent) > 0) {
        for _, vout := range unspent {
            txhash, _ := chainhash.NewHashFromStr(vout.TxID)
            if (w.verifyTxOnSubchain(*txhash)) { //theoretically only one unspent vout, but check anyway
                return true, vout
            }
        }
    }
    return false, btcjson.ListUnspentResult{}
}

// Find any previously unconfirmed transactions in order to start attestation from there
func (w *AttestClient) getUnconfirmedTx() (bool, Attestation) {
    mempool, err := w.mainClient.GetRawMempool()
    if err != nil {
        log.Fatal(err)
    }
    for _, hash := range mempool {
        if (w.verifyTxOnSubchain(*hash)) {
            return true, Attestation{*hash, w.getTxAttestedHash(*hash), false, time.Now()}
        }
    }
    return false, Attestation{chainhash.Hash{}, chainhash.Hash{}, true, time.Now()}
}

// Find the attested sidechain hash from a transaction, by testing for all sidechain hashes
func (w *AttestClient) getTxAttestedHash(txid chainhash.Hash) chainhash.Hash {
    // Get latest block and block height from sidechain
    latesthash, err := w.sideClient.GetBestBlockHash()
    if err != nil {
        log.Fatal(err)
    }
    latestheight, err := w.sideClient.GetBlockHeight(latesthash)
    if err != nil {
        log.Fatal(err)
    }

    // Get address from transaction
    tx, err := w.mainClient.GetTransaction(&txid)
    if err != nil {
        log.Fatal(err)
    }
    addr := tx.Details[0].Address

    // Check first if the attestation came from the latest block
    if (crypto.IsAddrTweakedFromHash(addr, latesthash.CloneBytes(), w.walletPriv, w.mainChainCfg)) {
        return *latesthash
    }

    // Iterate backwards through all sidechain hashes to find the block hash that was attested
    for h := latestheight - 1; h >= 0; h-- {
        hash, err := w.sideClient.GetBlockHash(int64(h))
        if err != nil {
            log.Fatal(err)
        }
        if (crypto.IsAddrTweakedFromHash(addr, hash.CloneBytes(), w.walletPriv, w.mainChainCfg)) {
            return *hash
        }
    }

    return chainhash.Hash{}
}
