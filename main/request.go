// Requests passed between attestation and confirmation services

package main

type Request struct {
    Name      string    `json:"name"`
    Id        string    `json:"hash"`
    Attested  bool      `json:"attested"`
}
