package models

type RequestGet_s struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type RequestPost_s struct {
	Name     string `json:"name"`
	ClientId string `json:"clientId"`
	Hash     string `json:"hash"`
	Height   string `json:"height"`
}
