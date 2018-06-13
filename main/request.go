// Requests passed between attestation and request services

package main

type Request struct {
    Name      string    `json:"name"`
    Id        string    `json:"hash"`
    Attested  bool      `json:"attested"`
    Error     string    `json:"error"`
}
