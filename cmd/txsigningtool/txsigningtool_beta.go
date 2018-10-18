package main

import (
    "fmt"
    "log"
    "bytes"

    "mainstay/messengers"
    "mainstay/config"

    "github.com/btcsuite/btcd/chaincfg/chainhash"
    "github.com/btcsuite/btcd/wire"
    zmq "github.com/pebbe/zmq4"
)

var (
    sub         *messengers.SubscriberZmq
    pub         *messengers.PublisherZmq
    latestHash  chainhash.Hash
    poller      *zmq.Poller
)

func main() {
    poller = zmq.NewPoller()

    topics := []string{config.TOPIC_NEW_HASH, config.TOPIC_NEW_TX}
    sub = messengers.NewSubscriberZmq(fmt.Sprintf("127.0.0.1:%d",config.MAIN_PUBLISHER_PORT), topics, poller)
    pub = messengers.NewPublisherZmq(5001, poller)

    for {
        sockets, _ := poller.Poll(-1)
        for _, socket := range sockets {
            if sub.Socket() == socket.Socket {
                topic, msg := sub.ReadMessage()
                switch topic {
                case config.TOPIC_NEW_TX:
                    processTx(msg)
                case config.TOPIC_NEW_HASH:
                    processHash(msg)
                }
            }
        }
    }
}

func processHash(msg []byte) {
    hash, hashErr := chainhash.NewHash(msg)
    if hashErr != nil {
        log.Fatal(hashErr)
    }
    fmt.Printf("hash %s\n", hash.String())
    latestHash = *hash
}

func processTx(msg []byte) {
    var msgTx wire.MsgTx
    if err := msgTx.Deserialize(bytes.NewReader(msg)); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("tx %s\n", msgTx.TxHash().String())

    pub.SendMessage([]byte("sig :D"), config.TOPIC_SIGS)
}
