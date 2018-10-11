package messengers

import (
    "fmt"

    zmq "github.com/pebbe/zmq4"
)

// Zmq publisher wrapper
// Extends Publisher interface
type PublisherZmq struct {
    socket  *zmq.Socket
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
// Bind to localhost and port provided
func NewPublisherZmq(port int, poller *zmq.Poller) *PublisherZmq {
    //  Prepare our publisher
    publisher, _ := zmq.NewSocket(zmq.PUB)
    publisher.Bind(fmt.Sprintf("tcp://*:%d", port))

    poller.Add(publisher, zmq.POLLOUT)

    return &PublisherZmq{publisher}
}
