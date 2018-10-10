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
func (p *PublisherZmq) SendMessage(msg string, topic string) {
    p.socket.Send(topic, zmq.SNDMORE)
    p.socket.Send(msg, 0)
}

// Close underlying zmq socket - To be used with defer
func (p *PublisherZmq) Close() {
    p.socket.Close()
    return
}

// Return new PublisherZmq instance
// Bind to localhost and port provided
func NewPublisherZmq(port int) *PublisherZmq {
    //  Prepare our publisher
    publisher, _ := zmq.NewSocket(zmq.PUB)
    publisher.Bind(fmt.Sprintf("tcp://*:%d", port))

    return &PublisherZmq{publisher}
}
