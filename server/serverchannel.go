package server

// Channel struct used to pass response channel for responses along with request
type RequestWithResponseChannel struct {
	Request  Request
	Response chan Response
}
