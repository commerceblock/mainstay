package models

type Channel struct {
	Requests  chan Request
	Responses chan interface{}
}

func NewChannel() *Channel {
	channel := &Channel{}
	channel.Requests = make(chan Request)
	channel.Responses = make(chan interface{})
	return channel
}
