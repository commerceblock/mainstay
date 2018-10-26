package requestapi

// Channel implementation with channel for Requests and channel for Responses
type Channel struct {
	Requests  chan Request
	Responses chan Response
}

// Return new Channel instance
func NewChannel() *Channel {
	channel := &Channel{}
	channel.Requests = make(chan Request)
	channel.Responses = make(chan Response)
	return channel
}

// Channel struct used to pass interface channel for responses along with request
type RequestWithInterfaceChannel struct {
	Request  Request
	Response chan interface{}
}

// Channel struct used to pass response channel for responses along with request
type RequestWithResponseChannel struct {
	Request  Request
	Response chan Response
}
