package main

type Request struct {
    Name      string    `json:"name"`
    Id        string    `json:"hash"`
    Attested  bool      `json:"attested"`
}
