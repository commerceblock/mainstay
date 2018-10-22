package models

type Response struct {
	Request Request
	Error   string `json:"error"`
}
