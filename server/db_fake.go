package server

import (
	"mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// DbFake struct
type DbFake struct {
	height       int32
	attestations []models.Attestation
	commitments  []models.Commitment
}

// Return new DbFake instance
func NewDbFake() (*DbFake, error) {
	return &DbFake{0, []models.Attestation{}, []models.Commitment{}}, nil
}

// Save latest attestation to attestations
func (d *DbFake) saveAttestation(attestation models.Attestation) error {
	d.attestations = append(d.attestations, attestation)
	return nil
}

// Save latest commitment to commitments
func (d *DbFake) saveCommitment(commitment models.Commitment) error {
	d.commitments = append(d.commitments, commitment)
	return nil
}

// Return latest attestation commitment hash
func (d *DbFake) getLatestAttestedCommitmentHash() (chainhash.Hash, error) {
	if len(d.attestations) == 0 {
		return chainhash.Hash{}, nil
	}
	latestAttestation := d.attestations[len(d.attestations)-1]
	return latestAttestation.CommitmentHash(), nil
}

// Fake client commitment hashes
var fakeLatestCommitment = []string{
	"1a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"2a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"3a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"4a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"5a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"6a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"7a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"8a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"9a39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"aa39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"ba39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"ca39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"da39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"ea39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
	"fa39e34e881d9a1e6cdc3418b54aa57747106bc75e9e84426661f27f98ada3b7",
}

// Return latest commitment from fake client commitments
func (d *DbFake) getLatestCommitment() (models.Commitment, error) {
	var commitmentHashes []chainhash.Hash

	commitmentHash, errHash := chainhash.NewHashFromStr(fakeLatestCommitment[d.height])
	if errHash != nil {
		return models.Commitment{}, errHash
	}
	commitmentHashes = append(commitmentHashes, *commitmentHash)

	commitment, errCommitment := models.NewCommitment(commitmentHashes)
	if errCommitment != nil {
		return models.Commitment{}, errCommitment
	}

	return *commitment, nil
}

// Return commitment for attestation with given txid
func (d *DbFake) getAttestationCommitment(attestationTxid chainhash.Hash) (models.Commitment, error) {
	return models.Commitment{}, nil
}
