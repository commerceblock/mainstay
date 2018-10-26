package requestapi

type Channel struct {
	Requests    chan Request
	Responses   chan Response
}

func NewChannel() *Channel {
	channel := &Channel{}
	channel.Requests = make(chan Request)
	channel.Responses = make(chan Response)
	return channel
}

type RequestWithInterfaceChannel struct {
    Request  Request
    Response chan interface{}
}

type RequestWithResponseChannel struct {
    Request  Request
    Response chan Response
}
