// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	"fmt"

	confpkg "mainstay/config"
	"mainstay/crypto"
	"mainstay/messengers"

	zmq "github.com/pebbe/zmq4"
)

// zmq communication consts
const (
	DEFAULT_MAIN_PUBLISHER_PORT = 5000 // port used by main signer publisher

	// predefined topics for publishing/subscribing via zmq
	TOPIC_NEW_HASH       = "H"
	TOPIC_NEW_TX         = "T"
	TOPIC_CONFIRMED_HASH = "C"
	TOPIC_SIGS           = "S"
)

// AttestSignerZmq struct
//
// Implements AttestSigner interface and uses communication
// via zmq to publish data and listen to subscriptions and
// send commitments/new tx and receive signatures
type AttestSignerZmq struct {
	// zmq publisher interface used to publish hashes and txes to signers
	publisher *messengers.PublisherZmq

	// zmq subscribe interface to signers to receive tx signatures
	subscribers []*messengers.SubscriberZmq
}

// poller to add all subscriber/publisher sockets
var poller *zmq.Poller

// Return new AttestSignerZmq instance
func NewAttestSignerZmq(config confpkg.SignerConfig) AttestSignerZmq {
	// get publisher addr from config, if set
	publisherAddr := fmt.Sprintf("*:%d", DEFAULT_MAIN_PUBLISHER_PORT)
	if config.Publisher != "" {
		publisherAddr = config.Publisher
	}

	// Initialise publisher for sending new hashes and txs
	// and subscribers to receive sig responses
	poller = zmq.NewPoller()
	publisher := messengers.NewPublisherZmq(publisherAddr, poller)
	var subscribers []*messengers.SubscriberZmq
	subtopics := []string{TOPIC_SIGS}
	for _, nodeaddr := range config.Signers {
		subscribers = append(subscribers, messengers.NewSubscriberZmq(nodeaddr, subtopics, poller))
	}

	return AttestSignerZmq{publisher, subscribers}
}

// Use zmq publisher to send confirmed hash
func (z AttestSignerZmq) SendConfirmedHash(hash []byte) {
	z.publisher.SendMessage(hash, TOPIC_CONFIRMED_HASH)
}

// Use zmq publisher to send new hash
func (z AttestSignerZmq) SendNewHash(hash []byte) {
	z.publisher.SendMessage(hash, TOPIC_NEW_HASH)
}

// Use zmq publisher to send new tx
func (z AttestSignerZmq) SendNewTx(tx []byte) {
	z.publisher.SendMessage(tx, TOPIC_NEW_TX)
}

// Split incoming message to a slice of messages by
// parsing message sizes and reading correct bytes
func getSplitMsgFromMsg(msg []byte) [][]byte {
	var msgs [][]byte

	if len(msg) == 0 {
		return [][]byte{}
	}

	it := 0
	for {
		partMsgSize := msg[it]
		partMsg := msg[it+1 : it+1+int(partMsgSize)]
		msgs = append(msgs, partMsg)
		it += 1 + int(partMsgSize)
		if len(msg) <= it {
			break
		}
	}
	return msgs
}

// Parse all received messages and create a sigs slice
// input:
// x dimension: subscriber
// y dimension: list of signatures of subscriber (one for each tx input)
// z dimension: slice of bytes (signature)
// output:
// x dimension: number of transaction inputs
// y dimension: number of signatures per input
func getSigsFromMsgs(msgs [][][]byte, numOfInputs int) [][]crypto.Sig {
	if numOfInputs == 0 {
		return [][]crypto.Sig{}
	}

	sigs := make([][]crypto.Sig, numOfInputs)
	for i := 0; i < numOfInputs; i++ {
		for _, msgSplit := range msgs {
			if len(msgSplit) > i {
				sigs[i] = append(sigs[i], crypto.Sig(msgSplit[i]))
			}
		}
	}
	return sigs
}

// Update num of transaction inputs from latest msg
func updateNumOfTxInputs(msgSplit [][]byte, numOfInputs int) int {
	if len(msgSplit) > numOfInputs {
		numOfInputs = len(msgSplit)
	}
	return numOfInputs
}

// Listen to zmq subscribers to receive tx signatures
func (z AttestSignerZmq) GetSigs() [][]crypto.Sig {
	var msgs [][][]byte
	numOfTxInputs := 0
	sockets, _ := poller.Poll(-1)
	for _, socket := range sockets {
		for _, sub := range z.subscribers {
			if sub.Socket() == socket.Socket {
				_, msg := sub.ReadMessage()
				var msgSplit [][]byte
				msgSplit = getSplitMsgFromMsg(msg)
				numOfTxInputs = updateNumOfTxInputs(msgSplit, numOfTxInputs)
				msgs = append(msgs, msgSplit)
			}
		}
	}
	return getSigsFromMsgs(msgs, numOfTxInputs)
}
