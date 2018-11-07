package server

import (
	"mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// DbFake struct
type DbFake struct {
	height            int32
	attestations      []models.Attestation
	merkleCommitments []models.CommitmentMerkleCommitment
	merkleProofs      []models.CommitmentMerkleProof
}

// Return new DbFake instance
func NewDbFake() *DbFake {
	return &DbFake{0, []models.Attestation{}, []models.CommitmentMerkleCommitment{}, []models.CommitmentMerkleProof{}}
}

// Save latest attestation to attestations
func (d *DbFake) saveAttestation(attestation models.Attestation) error {
	d.attestations = append(d.attestations, attestation)
	return nil
}

// Save merkle commitments to the MerkleCommitment collection
func (d *DbFake) saveMerkleCommitments(commitments []models.CommitmentMerkleCommitment) error {
	d.merkleCommitments = append(d.merkleCommitments, commitments...)
	return nil
}

// Save merkle proofs to the MerkleProof collection
func (d *DbFake) saveMerkleProofs(proofs []models.CommitmentMerkleProof) error {
	d.merkleProofs = append(d.merkleProofs, proofs...)
	return nil
}

// Return latest attestation commitment hash
func (d *DbFake) getLatestAttestationMerkleRoot() (string, error) {
	if len(d.attestations) == 0 {
		return "", nil
	}
	latestAttestation := d.attestations[len(d.attestations)-1]
	return latestAttestation.CommitmentHash().String(), nil
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
func (d *DbFake) getLatestCommitments() ([]models.LatestCommitment, error) {
	var latestCommitments []models.LatestCommitment

	commitmentHash, errHash := chainhash.NewHashFromStr(fakeLatestCommitment[d.height])
	if errHash != nil {
		return []models.LatestCommitment{}, errHash
	}
	latestCommitments = append(latestCommitments, models.LatestCommitment{*commitmentHash, 0})

	return latestCommitments, nil
}

// Return commitment for attestation with given txid
func (d *DbFake) getAttestationMerkleCommitments(attestationTxid chainhash.Hash) ([]models.CommitmentMerkleCommitment, error) {
	return []models.CommitmentMerkleCommitment{}, nil
}
