package main

import "time"

type Request struct {
    Name      string    `json:"name"`
    Id        string    `json:"hash"`
    Completed bool      `json:"completed"`
    Due       time.Time `json:"due"`
}
