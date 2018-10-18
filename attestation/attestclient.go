package attestation

import (
    "log"

    "mainstay/crypto"
    "mainstay/clients"
    confpkg "mainstay/config"

    "github.com/btcsuite/btcd/btcjson"
    "github.com/btcsuite/btcd/rpcclient"
    "github.com/btcsuite/btcutil"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
    "github.com/btcsuite/btcd/chaincfg"
    "github.com/btcsuite/btcd/wire"
    "github.com/btcsuite/btcd/btcec"
    "github.com/btcsuite/btcd/txscript"
)

// AttestClient structure
// Maintains RPC connections to main and side chain clients
// Handles generating staychain next address and next transaction
// and verifying that the correct chain of transactions is maintained
type AttestClient struct {
    MainClient      *rpcclient.Client
    sideClient      clients.SidechainClient
    MainChainCfg    *chaincfg.Params
    pk0             string
    txid0           string
    script0         string
    pubkeys         []*btcec.PublicKey
    numOfSigs       int
    WalletPriv      *btcutil.WIF
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

    multisig := config.MultisigScript()
    if multisig != "" { // if multisig attestation, parse pubkeys
        pubkeys, numOfSigs := crypto.ParseRedeemScript(config.MultisigScript())
        return &AttestClient{config.MainClient(), config.OceanClient(), config.MainChainCfg(), pk, config.InitTX(), multisig, pubkeys, numOfSigs, pkWif}
    }
    return &AttestClient{config.MainClient(), config.OceanClient(), config.MainChainCfg(), pk, config.InitTX(), multisig, []*btcec.PublicKey{}, 1, pkWif}
}

// Get next attestation key by tweaking with latest hash
func (w *AttestClient) GetNextAttestationKey(hash chainhash.Hash) *btcutil.WIF {
    // Tweak priv key with the latest ocean hash
    tweakedWalletPriv := crypto.TweakPrivKey(w.WalletPriv, hash.CloneBytes(), w.MainChainCfg)

    // Import tweaked priv key to wallet
    importErr := w.MainClient.ImportPrivKeyRescan(tweakedWalletPriv, hash.String(), false)
    if importErr != nil {
        log.Fatal(importErr)
    }

    return tweakedWalletPriv
}

// Get next attestation address from private key
func (w *AttestClient) GetNextAttestationAddr(key *btcutil.WIF, hash chainhash.Hash) (btcutil.Address, string) {

    myAddr := crypto.GetAddressFromPrivKey(key, w.MainChainCfg)

    // In multisig case tweak all initial pubkeys and import
    // a multisig address to the main client wallet
    if len(w.pubkeys) > 0 {
        var tweakedPubs []*btcec.PublicKey
        myFound := false
        hashBytes := hash.CloneBytes()
        for _, pub := range w.pubkeys {
            tweakedPub := crypto.TweakPubKey(pub, hashBytes)
            tweakedPubs = append(tweakedPubs, tweakedPub)
            if myAddr.String() ==
            crypto.GetAddressFromPubKey(tweakedPub, w.MainChainCfg).String() {
                myFound = true
            }
        }
        if !myFound {
            log.Fatal("Client address missing from tweaked pubkey addresses")
        }

        multisigAddr, redeemScript := crypto.CreateMultisig(tweakedPubs, w.numOfSigs, w.MainChainCfg)

        importErr := w.MainClient.ImportAddress(multisigAddr.String())
        if importErr != nil {
            log.Fatal(importErr)
        }
        // importaddress for P2SH no currently supported by btcd code
        // importErr2 := w.MainClient.ImportAddress(redeemScript)
        // if importErr2 != nil {
        //     log.Printf("import error")
        //     log.Fatal(importErr2)
        // }

        return multisigAddr, redeemScript
    }

    return myAddr, ""
}

// Generate a new transaction paying to the tweaked address and add fees
func (w *AttestClient) createAttestation(paytoaddr btcutil.Address, txunspent btcjson.ListUnspentResult, useDefaultFee bool) *wire.MsgTx {
    inputs := []btcjson.TransactionInput{{Txid: txunspent.TxID, Vout: txunspent.Vout},}

    amounts := map[btcutil.Address]btcutil.Amount{paytoaddr: btcutil.Amount(txunspent.Amount*100000000)}
    msgtx, err := w.MainClient.CreateRawTransaction(inputs, amounts, nil)
    if err != nil {
        log.Fatal(err)
    }

    feePerByte := GetFee(useDefaultFee)
    fee := int64(feePerByte * msgtx.SerializeSize())
    msgtx.TxOut[0].Value -= fee

    return msgtx
}

// Given a hash return the corresponding client private key and redeemscript
func (w *AttestClient) GetKeyAndScriptFromHash(hash chainhash.Hash) (btcutil.WIF, string) {
    var key btcutil.WIF
    var redeemScript string
    if !hash.IsEqual(&chainhash.Hash{}) {
        key = *crypto.TweakPrivKey(w.WalletPriv, hash.CloneBytes(), w.MainChainCfg)
        _, redeemScript = w.GetNextAttestationAddr(&key, hash)
    } else {
        key = *w.WalletPriv
        redeemScript = w.script0
    }
    return key, redeemScript
}

// Sign and send the latest attestation transaction
func (w *AttestClient) signAndSendAttestation(msgtx *wire.MsgTx, txunspent btcjson.ListUnspentResult, sigs []string, hash chainhash.Hash) chainhash.Hash {

    // Calculate private key and redeemScript from hash
    key, redeemScript := w.GetKeyAndScriptFromHash(hash)
    // Can't get redeem script from unspent as importaddress P2SH not supported
    // if txunspent.RedeemScript != "" {
    //     redeemScript = txunspent.RedeemScript
    // }

    rawtxinput := btcjson.RawTxInput{txunspent.TxID, txunspent.Vout, txunspent.ScriptPubKey, redeemScript}
    signedmsgtx, issigned, err := w.MainClient.SignRawTransaction3(msgtx, []btcjson.RawTxInput{rawtxinput}, []string{key.String()})
    if err != nil{
        log.Fatal(err)
    } else if !issigned {
        log.Printf("incomplete signing")
    }

    // combine client sig with sigs

    txhash, err := w.MainClient.SendRawTransaction(signedmsgtx, false)
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
        txraw, err := w.MainClient.GetRawTransaction(&txid)
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
    unspent, err := w.MainClient.ListUnspent()
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
    mempool, err := w.MainClient.GetRawMempool()
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
    tx, err := w.MainClient.GetRawTransaction(&txid)
    if err != nil {
        log.Fatal(err)
    }
    _, addrs, _, errExtract := txscript.ExtractPkScriptAddrs(tx.MsgTx().TxOut[0].PkScript, w.MainChainCfg)
    if errExtract != nil {
        log.Fatal(errExtract)
    }
    addr := addrs[0]

    tweakedPriv := crypto.TweakPrivKey(w.WalletPriv, latesthash.CloneBytes(), w.MainChainCfg)
    addrTweaked, _ := w.GetNextAttestationAddr(tweakedPriv, *latesthash)
    // Check first if the attestation came from the latest block
    if (addr.String() == addrTweaked.String()) {
        return *latesthash
    }

    // Iterate backwards through all sidechain hashes to find the block hash that was attested
    for h := latestheight - 1; h >= 0; h-- {
        hash, err := w.sideClient.GetBlockHash(int64(h))
        if err != nil {
            log.Fatal(err)
        }
        tweakedPriv := crypto.TweakPrivKey(w.WalletPriv, hash.CloneBytes(), w.MainChainCfg)
        addrTweaked, _ := w.GetNextAttestationAddr(tweakedPriv, *hash)
        if (addr.String() == addrTweaked.String()) {
            return *hash
        }
    }

    return chainhash.Hash{}
}
