// Wallet transaction generation and general handling

package attestation

import (
    "log"
    "time"
    "ocean-attestation/crypto"
    "github.com/btcsuite/btcd/btcjson"
    "github.com/btcsuite/btcd/rpcclient"
    "github.com/btcsuite/btcutil"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
    "github.com/btcsuite/btcd/chaincfg"
)

const FEE_PER_BYTE = 20 // satoshis

type AttestClient struct {
    mainClient  *rpcclient.Client
    sideClient  *rpcclient.Client
    mainChainCfg *chaincfg.Params
    pk0         string
    txid0       string
    walletPriv  *btcutil.WIF
}

func NewAttestClient(rpcMain *rpcclient.Client, rpcSide *rpcclient.Client, cfgMain *chaincfg.Params, pk string, tx string) *AttestClient {
    if (false && len(pk) != 64 /*need to validate key properly*/) {
        log.Fatal("*AttestClient* Incorrect key size")
    }
    return &AttestClient{rpcMain, rpcSide, cfgMain, pk, tx, crypto.GetWalletPrivKey(pk)}
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
    importErr := w.mainClient.ImportPrivKey(tweakedWalletPriv)
    if importErr != nil {
        log.Fatal(importErr)
    }

    return *hash, addr
}

// Generate a new transaction paying to the tweaked address, add fees and send the transaction through the wallet client
func (w *AttestClient) sendAttestation(paytoaddr btcutil.Address, txunspent btcjson.ListUnspentResult) (chainhash.Hash) {
    inputs := []btcjson.TransactionInput{{Txid: txunspent.TxID, Vout: txunspent.Vout},}

    amounts := map[btcutil.Address]btcutil.Amount{paytoaddr: btcutil.Amount(txunspent.Amount*100000000)}
    msgtx, err := w.mainClient.CreateRawTransaction(inputs, amounts, nil)
    if err != nil {
        log.Fatal(err)
    }

    fee := int64(FEE_PER_BYTE * msgtx.SerializeSize())
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

// Find any previously unconfirmed transactions and start attestation from there
func (w *AttestClient) getUnconfirmedTx(tx *Attestation) {
    mempool, err := w.mainClient.GetRawMempool()
    if err != nil {
        log.Fatal(err)
    }
    for _, hash := range mempool {
        if (w.verifyTxOnSubchain(*hash)) {
            *tx = Attestation{*hash, chainhash.Hash{}, false, time.Now()}
            return
        }
    }
}
