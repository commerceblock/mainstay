package models

type RequestGet_s struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type RequestPost_s struct {
	Url    string `json:"url"`
	Hash   string `json:"hach"`
	Header string `json:"header"`
	Name   string `json:"name"`
	Height string `json:"height"`
	Id     string `json:"id"`
}
