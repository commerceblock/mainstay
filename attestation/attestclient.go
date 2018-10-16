package attestation

import (
    "log"

    "ocean-attestation/crypto"
    "ocean-attestation/clients"
    confpkg "ocean-attestation/config"

    "github.com/btcsuite/btcd/btcjson"
    "github.com/btcsuite/btcd/rpcclient"
    "github.com/btcsuite/btcutil"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
    "github.com/btcsuite/btcd/chaincfg"
    "github.com/btcsuite/btcd/wire"
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
    script0         string
    walletPriv      *btcutil.WIF
}

// NewAttestClient returns a pointer to a new AttestClient instance
// Initially locates the genesis transaction in the main chain wallet
// and verifies that the corresponding private key is in the wallet
func NewAttestClient(config *confpkg.Config) *AttestClient {
    // Get initial private key from initial funding transaction of main client
    pk := config.InitPK()
    pkWif := crypto.GetWalletPrivKey(pk)
    importErr := config.MainClient().ImportPrivKeyRescan(pkWif, "init", false)
    if importErr != nil {
        log.Fatal(importErr)
    }

    return &AttestClient{config.MainClient(), config.OceanClient(), config.MainChainCfg(), pk, config.InitTX(), config.MultisigScript(), pkWif}
}

// Get latest client hash to use for attestation key tweaking
func (w *AttestClient) getNextAttestationHash() (chainhash.Hash) {
    hash, err := w.sideClient.GetBestBlockHash()
    if err != nil {
        log.Fatal(err)
    }
    return *hash
}

// Get next attestation key by tweaking with latest hash
func (w *AttestClient) getNextAttestationKey(hash chainhash.Hash) *btcutil.WIF {
    // Tweak priv key with the latest ocean hash
    tweakedWalletPriv := crypto.TweakPrivKey(w.walletPriv, hash.CloneBytes(), w.mainChainCfg)

    // Import tweaked priv key to wallet
    importErr := w.mainClient.ImportPrivKeyRescan(tweakedWalletPriv, hash.String(), false)
    if importErr != nil {
        log.Fatal(importErr)
    }

    return tweakedWalletPriv
}

// Get next attestation address from private key
func (w *AttestClient) getNextAttestationAddr(key *btcutil.WIF) btcutil.Address {

    // tweak all keys
    // verify mine is correct

    addr := crypto.GetAddressFromPrivKey(key, w.mainChainCfg)

    // importaddress addr
    // importaddress script

    return addr
}

// Generate a new transaction paying to the tweaked address and add fees
func (w *AttestClient) createAttestation(paytoaddr btcutil.Address, txunspent btcjson.ListUnspentResult, useDefaultFee bool) *wire.MsgTx {
    inputs := []btcjson.TransactionInput{{Txid: txunspent.TxID, Vout: txunspent.Vout},}

    amounts := map[btcutil.Address]btcutil.Amount{paytoaddr: btcutil.Amount(txunspent.Amount*100000000)}
    msgtx, err := w.mainClient.CreateRawTransaction(inputs, amounts, nil)
    if err != nil {
        log.Fatal(err)
    }

    feePerByte := GetFee(useDefaultFee)
    fee := int64(feePerByte * msgtx.SerializeSize())
    msgtx.TxOut[0].Value -= fee

    return msgtx
}

// Sign and send the latest attestation transaction
func (w *AttestClient) signAndSendAttestation(msgtx *wire.MsgTx, txunspent btcjson.ListUnspentResult, sigs []string) chainhash.Hash {

    // Get redeem script from unspent - Priv key should already have been imported
    redeemScript := ""
    if txunspent.RedeemScript != "" {
        redeemScript = txunspent.RedeemScript
    }

    rawtxinput := btcjson.RawTxInput{txunspent.TxID, txunspent.Vout, txunspent.ScriptPubKey, redeemScript}
    signedmsgtx, issigned, err := w.mainClient.SignRawTransaction2(msgtx, []btcjson.RawTxInput{rawtxinput})
    if err != nil{
        log.Fatal(err)
    } else if !issigned {
        log.Printf("incomplete signing")
    }

    // add sigs

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
            return true, *NewAttestation(*hash, w.getTxAttestedHash(*hash), ASTATE_UNCONFIRMED)
        }
    }
    return false, *NewAttestation(chainhash.Hash{}, chainhash.Hash{}, ASTATE_NEW_ATTESTATION)
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
