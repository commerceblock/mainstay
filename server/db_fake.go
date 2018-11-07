package server

import (
	"errors"

	"mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// DbFake struct
type DbFake struct {
	attestations      []models.Attestation
	merkleCommitments []models.CommitmentMerkleCommitment
	merkleProofs      []models.CommitmentMerkleProof
	latestCommitments []models.LatestCommitment
}

// Return new DbFake instance
func NewDbFake() *DbFake {
	return &DbFake{
		[]models.Attestation{},
		[]models.CommitmentMerkleCommitment{},
		[]models.CommitmentMerkleProof{},
		[]models.LatestCommitment{}}
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

// Set latest commitments for testing
func (d *DbFake) setLatestCommitments(latestCommitments []models.LatestCommitment) {
	d.latestCommitments = latestCommitments
}

// Return latest commitment from fake client commitments
func (d *DbFake) getLatestCommitments() ([]models.LatestCommitment, error) {
	return d.latestCommitments, nil
}

// Return commitment for attestation with given txid
func (d *DbFake) getAttestationMerkleCommitments(txid chainhash.Hash) ([]models.CommitmentMerkleCommitment, error) {
	if len(d.attestations) == 0 {
		return []models.CommitmentMerkleCommitment{}, nil
	}

	var merkleCommitments []models.CommitmentMerkleCommitment
	for _, attestation := range d.attestations {
		if txid == attestation.Txid {
			for _, commitment := range d.merkleCommitments {
				if commitment.MerkleRoot == attestation.CommitmentHash() {
					merkleCommitments = append(merkleCommitments, commitment)
				}
			}
		}
	}

	// not found - error
	if len(merkleCommitments) == 0 {
		return merkleCommitments, errors.New("Merkle commitments not found")
	}
	return merkleCommitments, nil
}
