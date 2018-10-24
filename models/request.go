package models

type RequestGet_s struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type RequestPost_s struct {
	Url    string `json:"url"`
	Header string `json:"header"`
	Name   string `json:"name"`
	Id     string `json:"id"`
}
