package models

type Channel struct {
	RequestGet  chan RequestGet_s
	RequestPost chan RequestPost_s
	Responses   chan interface{}
}

func NewChannel() *Channel {
	channel := &Channel{}
	channel.RequestGet = make(chan RequestGet_s)
	channel.RequestPost = make(chan RequestPost_s)
	channel.Responses = make(chan interface{})
	return channel
}
