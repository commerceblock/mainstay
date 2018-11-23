// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package messengers

import (
	"fmt"

	zmq "github.com/pebbe/zmq4"
)

// Zmq router wrapper
type DealerZmq struct {
	socket *zmq.Socket
}

// Read message from router socket
func (d *DealerZmq) ReadMessage() []byte {
	msg, _ := d.socket.RecvBytes(0)
	return msg
}

// Send message from router socket to identity
func (d *DealerZmq) SendMessage(msg []byte) {
	d.socket.SendBytes([]byte(""), zmq.SNDMORE) // inform listener
	d.socket.SendBytes(msg, 0)                  // send message
}

// Close underlying zmq socket - To be used with defer
func (d *DealerZmq) Close() {
	d.socket.Close()
	return
}

// Return underlying socket
func (d *DealerZmq) Socket() *zmq.Socket {
	return d.socket
}

// Return new DealerZmq instance
// Bind to localhost and port provided and set identity
func NewDealerZmq(port int, identity string, poller *zmq.Poller) *DealerZmq {
	//  Prepare our router
	dealer, _ := zmq.NewSocket(zmq.DEALER)
	dealer.SetIdentity(identity) // unique client identity
	dealer.Connect(fmt.Sprintf("tcp://localhost:%d", port))

	if poller != nil {
		poller.Add(dealer, zmq.POLLIN)
	}

	return &DealerZmq{dealer}
}
