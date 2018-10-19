package messengers

import (
	"fmt"
	"strings"

	zmq "github.com/pebbe/zmq4"
)

// Zmq subscriber wrapper
// Extends Subscriber interface
type SubscriberZmq struct {
	socket *zmq.Socket
}

// Read topic-msg from zmq socket
func (s *SubscriberZmq) ReadMessage() (string, []byte) {

	//  Read envelope with address
	address, _ := s.socket.RecvBytes(0)
	//  Read message contents
	contents, _ := s.socket.RecvBytes(0)

	return string(address), contents
}

// Close underlying zmq socket - To be used with defer
func (s *SubscriberZmq) Close() {
	s.socket.Close()
}

// Return underlying socket
func (s *SubscriberZmq) Socket() *zmq.Socket {
	return s.socket
}

// Return new SubscriberZmq instance
// Connect to address provided and subscribe to topics
func NewSubscriberZmq(address string, topics []string, poller *zmq.Poller) *SubscriberZmq {

	// Get host/port
	addrComp := strings.Split(address, ":")

	//  Prepare our subscriber
	subscriber, _ := zmq.NewSocket(zmq.SUB)
	subscriber.Connect(fmt.Sprintf("tcp://%s:%s", addrComp[0], addrComp[1]))

	for _, topic := range topics {
		subscriber.SetSubscribe(topic)
	}

	poller.Add(subscriber, zmq.POLLIN)

	return &SubscriberZmq{subscriber}
}
