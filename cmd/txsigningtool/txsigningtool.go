package main

import (
    "fmt"
    "log"
    "bytes"
    "flag"
    "os"
    "os/exec"
    "time"
    "encoding/hex"

    "mainstay/messengers"
    confpkg "mainstay/config"
    "mainstay/attestation"
    "mainstay/crypto"

    "github.com/btcsuite/btcd/chaincfg/chainhash"
    "github.com/btcsuite/btcd/btcjson"
    "github.com/btcsuite/btcd/wire"
    "github.com/btcsuite/btcd/txscript"
    zmq "github.com/pebbe/zmq4"
)

var (
    tx0             string
    pk0             string
    script          string
    isRegtest       bool
    sub             *messengers.SubscriberZmq
    pub             *messengers.PublisherZmq
    attestedHash    chainhash.Hash
    nextHash        chainhash.Hash
    poller          *zmq.Poller
    client          *attestation.AttestClient
)

const CONF_PATH = "/src/mainstay/cmd/txsigningtool/conf.json"
const DEMO_CONF_PATH = "/src/mainstay/cmd/txsigningtool/demo-conf.json"
const DEMO_INIT_PATH = "/src/mainstay/cmd/txsigningtool/demo-init-signingtool.sh"

func parseFlags() {
    flag.BoolVar(&isRegtest, "regtest", false, "Use regtest wallet configuration")
    flag.StringVar(&tx0, "tx", "", "Tx id for genesis attestation transaction")
    flag.StringVar(&pk0, "pk", "", "Client pk for genesis attestation transaction")
    flag.StringVar(&script, "script", "", "Redeem script in case multisig is used")
    flag.Parse()

    if (tx0 == "" || pk0 == "") && !isRegtest {
        flag.PrintDefaults()
        log.Fatalf("Need to provide both -tx and -pk argument. To use test configuration set the -regtest flag.")
    }
}

func init() {
    parseFlags()
    var config *confpkg.Config

    if isRegtest {
        // btc regtest node setup
        cmd := exec.Command("/bin/sh", os.Getenv("GOPATH") + DEMO_INIT_PATH)
        output, err := cmd.Output()
        if err != nil {
            log.Fatal(err)
        }
        log.Printf("%s\n", output)

        confFile := confpkg.GetConfFile(os.Getenv("GOPATH") + DEMO_CONF_PATH)
        config = confpkg.NewConfig(false, confFile)
        tx0 = "cfda78d3ad0510bd49f971b28fd3c58d38c2a8dd79dd97c9a76ece8002507697"
        pk0 = "cSS9R4XPpajhqy28hcfHEzEzAbyWDqBaGZR4xtV7Jg8TixSWee1x"
        script = "512103e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b332102f3a78a7bd6cf01c56312e7e828bef74134dfb109e59afd088526212d96518e7552ae"
    } else {
        confFile := confpkg.GetConfFile(os.Getenv("GOPATH") + CONF_PATH)
        config = confpkg.NewConfig(false, confFile)
    }

    config.SetInitTX(tx0)
    config.SetInitPK(pk0)
    config.SetMultisigScript(script)
    client = attestation.NewAttestClient(config)

    // comms setup
    poller = zmq.NewPoller()
    topics := []string{confpkg.TOPIC_NEW_HASH, confpkg.TOPIC_NEW_TX, confpkg.TOPIC_CONFIRMED_HASH}
    sub = messengers.NewSubscriberZmq(fmt.Sprintf("127.0.0.1:%d",confpkg.MAIN_PUBLISHER_PORT), topics, poller)
    pub = messengers.NewPublisherZmq(5001, poller)
}

func main() {
    for {
        sockets, _ := poller.Poll(-1)
        for _, socket := range sockets {
            if sub.Socket() == socket.Socket {
                topic, msg := sub.ReadMessage()
                switch topic {
                case confpkg.TOPIC_NEW_TX:
                    processTx(msg)
                case confpkg.TOPIC_NEW_HASH:
                    nextHash = processHash(msg)
                    fmt.Printf("nexthash %s\n", nextHash.String())
                case confpkg.TOPIC_CONFIRMED_HASH:
                    attestedHash = processHash(msg)
                    fmt.Printf("attestedhash %s\n", attestedHash.String())
                }
            }
        }
        time.Sleep(1 * time.Second)
    }
}

func processHash(msg []byte) chainhash.Hash {
    hash, hashErr := chainhash.NewHash(msg)
    if hashErr != nil {
        log.Fatal(hashErr)
    }
    return *hash
}

func verifyTx(tx wire.MsgTx) bool {
    nextKey := client.GetNextAttestationKey(nextHash)
    nextAddr, _ := client.GetNextAttestationAddr(nextKey, nextHash)
    // exactr addr from unsigned tx and verify addresses match
    _, txScriptAddrs, _, err := txscript.ExtractPkScriptAddrs(tx.TxOut[0].PkScript, client.MainChainCfg)
    if err != nil {
        log.Fatal(err)
    }
    txAddr := txScriptAddrs[0]
    if txAddr.String() == nextAddr.String() {
        fmt.Printf("tx address %s verified\n", txAddr.String())
        return true
    }
    fmt.Printf("tx address %s not verified\n", txAddr.String())
    return false
}

func processTx(msg []byte) {
    // parse received tx into useful format
    var msgTx wire.MsgTx
    if err := msgTx.Deserialize(bytes.NewReader(msg)); err != nil {
        log.Fatal(err)
    }

    // verify transaction first
    if !verifyTx(msgTx) {
        return
    }

    // Get previous pk - redeem script
    key, redeemScript := client.GetKeyAndScriptFromHash(attestedHash)

    // sign tx and send signature to main attestation client
    prevTxId := msgTx.TxIn[0].PreviousOutPoint.Hash
    prevTx, errRaw := client.MainClient.GetRawTransaction(&prevTxId)
    if errRaw!= nil {
        log.Println(errRaw)
        return
    }

    // Sign transaction
    rawTxInput := btcjson.RawTxInput{prevTxId.String(), 0, hex.EncodeToString(prevTx.MsgTx().TxOut[0].PkScript), redeemScript}
    signedMsgTx, _, errSign := client.MainClient.SignRawTransaction3(&msgTx, []btcjson.RawTxInput{rawTxInput}, []string{key.String()})
    if errSign != nil{
        log.Println(errSign)
        return
    }

    scriptSig := signedMsgTx.TxIn[0].SignatureScript
    if len(scriptSig) > 0 {
        sigs, _ := crypto.ParseScriptSig(scriptSig)
        pub.SendMessage(sigs[0], confpkg.TOPIC_SIGS)
    }
}
