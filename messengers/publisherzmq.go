// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package messengers

import (
	"fmt"

	zmq "github.com/pebbe/zmq4"
)

// Zmq publisher wrapper
type PublisherZmq struct {
	socket *zmq.Socket
}

// Publish message via zmq socket
func (p *PublisherZmq) SendMessage(msg []byte, topic string) {
	p.socket.SendBytes([]byte(topic), zmq.SNDMORE)
	p.socket.SendBytes(msg, 0)
}

// Close underlying zmq socket - To be used with defer
func (p *PublisherZmq) Close() {
	p.socket.Close()
	return
}

// Return underlying socket
func (p *PublisherZmq) Socket() *zmq.Socket {
	return p.socket
}

// Return new PublisherZmq instance
// Bind address provided to constructor
func NewPublisherZmq(addr string, poller *zmq.Poller) *PublisherZmq {
	//  Prepare our publisher
	publisher, _ := zmq.NewSocket(zmq.PUB)
	publisher.Bind(fmt.Sprintf("tcp://%s", addr))

	poller.Add(publisher, zmq.POLLOUT)

	return &PublisherZmq{publisher}
}
