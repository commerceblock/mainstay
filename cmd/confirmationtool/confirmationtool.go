// Staychain confirmation tool
package main

import (
    "os"
    "log"
    "time"
    "flag"

    "ocean-attestation/config"
    "ocean-attestation/crypto"
    "ocean-attestation/staychain"

    "github.com/btcsuite/btcd/chaincfg/chainhash"
    "github.com/btcsuite/btcutil"
)

// Use staychain package to read attestations, verify and print information

const MAIN_NAME     = "bitcoin"
const SIDE_NAME     = "ocean"
const CONF_PATH     = "/src/ocean-attestation/cmd/confirmationtool/conf.json"
const FUNDING_TX    = "bf41c0da8047b1416d5ca680e2643967b27537cdf9a41527034698c336b55313"
const FIRST_TX      = "902fd11c3166eb07864a7b8ed0a3a0fbda0f4c26423b8eee4dd94420cfbae40e"

var (
    tx              string
    pk              string
    pkWIF           *btcutil.WIF
    showDetails     bool
    mainConfig      *config.Config
)

// init
func init() {
    flag.BoolVar(&showDetails, "detailed", false, "Detailed information on attestation transaction")
    flag.StringVar(&tx, "tx", "", "Tx id from which to start searching the staychain")
    flag.StringVar(&pk, "pk", "", "Private key for genesis attestation transaction")
    flag.Parse()
    if (tx == "") {
        tx = FUNDING_TX
    }
    if pk != "" {
        pkWIF = crypto.GetWalletPrivKey(pk)
    }

    confFile := config.GetConfFile(os.Getenv("GOPATH") + CONF_PATH)
    mainConfig = config.NewConfig(false, confFile)
}

// main method
func main() {
    defer mainConfig.MainClient().Shutdown()
    defer mainConfig.OceanClient().Close()

    txraw := getRawTxFromHash(tx)
    tx0raw := getRawTxFromHash(FIRST_TX)

    fetcher := staychain.NewChainFetcher(mainConfig.MainClient(), txraw)
    chain := staychain.NewChain(fetcher)
    verifier := staychain.NewChainVerifier(mainConfig.MainChainCfg(), mainConfig.OceanClient(), tx0raw)

    time.AfterFunc(5*time.Minute, func() {
        log.Println("Exit: ", chain.Close())
    })

    // await new attestations and verify
    for transaction := range chain.Updates() {
        log.Println("Verifying attestation")
        log.Printf("txid: %s\n", transaction.Txid)
        info, err := verifier.Verify(transaction)
        if err != nil {
            log.Fatal(err)
        } else {
            printAttestation(transaction, info)
            if pkWIF != nil {
                printDerivedKey(info)
            }
        }
    }
}

// Get raw transaction from a tx string hash using rpc client
func getRawTxFromHash(hashstr string) staychain.Tx {
    txhash, errHash := chainhash.NewHashFromStr(hashstr)
    if errHash != nil {
        log.Println("Invalid tx id provided")
        log.Fatal(errHash)
    }
    txraw, errGet := mainConfig.MainClient().GetRawTransactionVerbose(txhash)
    if errGet != nil {
        log.Println("Inititial transcaction does not exist")
        log.Fatal(errGet)
    }
    return staychain.Tx(*txraw)
}

// print attestation information
func printAttestation(tx staychain.Tx, info staychain.ChainVerifierInfo) {
    log.Println("Attestation Verified")
    if showDetails {
        log.Printf("%+v\n", tx)
    } else {
        log.Printf("%s blockhash: %s\n", MAIN_NAME, tx.BlockHash)
    }
    log.Printf("%s blockhash: %s\n", SIDE_NAME, info.Hash().String())
    log.Printf("%s blockheight: %d\n", SIDE_NAME, info.Height())
    log.Printf("\n")
}

// print derived private key for attestation
func printDerivedKey(info staychain.ChainVerifierInfo) {
    tweak_hash := info.Hash()
    tweaked_priv := crypto.TweakPrivKey(pkWIF, tweak_hash.CloneBytes(), mainConfig.MainChainCfg())
    log.Printf("%s privkey: %s\n", MAIN_NAME, tweaked_priv.String())
}
