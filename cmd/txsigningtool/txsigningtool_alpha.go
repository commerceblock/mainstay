package main

import (
    "fmt"
    "log"
    "bytes"

    "ocean-attestation/messengers"
    "ocean-attestation/config"

    "github.com/btcsuite/btcd/chaincfg/chainhash"
    "github.com/btcsuite/btcd/wire"
)

func main() {
    // subscribe topics
    topics := []string{config.TOPIC_NEW_HASH, config.TOPIC_NEW_TX}
    sub := messengers.NewSubscriberZmq(fmt.Sprintf("127.0.0.1:%d",config.MAIN_PUBLISHER_PORT), topics)

    for {
        topic, msg := sub.ReadMessage()

        switch topic {
        case config.TOPIC_NEW_TX:
            processTx(msg)
        case config.TOPIC_NEW_HASH:
            processHash(msg)
        }
    }
}

func processHash(msg []byte) {
    hash, hashErr := chainhash.NewHash(msg)
    if hashErr != nil {
        log.Fatal(hashErr)
    }
    fmt.Printf("hash %s\n", hash.String())
}

func processTx(msg []byte) {
    var msgTx wire.MsgTx
    if err := msgTx.Deserialize(bytes.NewReader(msg)); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("tx %s\n", msgTx.TxHash().String())
}
