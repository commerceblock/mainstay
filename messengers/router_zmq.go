// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package messengers

import (
	"fmt"

	zmq "github.com/pebbe/zmq4"
)

// Zmq router wrapper
type RouterZmq struct {
	socket *zmq.Socket
}

// Read message from router socket
func (r *RouterZmq) ReadMessage() []byte {
	msg, _ := r.socket.RecvBytes(0)
	return msg
}

// Send message from router socket to identity
func (r *RouterZmq) SendMessage(identity []byte, msg []byte) {
	r.socket.SendBytes(identity, zmq.SNDMORE)
	r.socket.SendBytes([]byte(""), zmq.SNDMORE)
	r.socket.SendBytes(msg, 0)
}

// Close underlying zmq socket - To be used with defer
func (r *RouterZmq) Close() {
	r.socket.Close()
	return
}

// Return underlying socket
func (r *RouterZmq) Socket() *zmq.Socket {
	return r.socket
}

// Return new RouterZmq instance
// Bind to localhost and port provided
func NewRouterZmq(port int, poller *zmq.Poller) *RouterZmq {
	//  Prepare our router
	router, _ := zmq.NewSocket(zmq.ROUTER)
	router.Bind(fmt.Sprintf("tcp://*:%d", port))

	if poller != nil {
		poller.Add(router, zmq.POLLIN)
	}

	return &RouterZmq{router}
}
