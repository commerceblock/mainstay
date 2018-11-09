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
	for _, a := range d.attestations {
		if a.Txid == attestation.Txid {
			a = attestation
			return nil
		}
	}
	d.attestations = append(d.attestations, attestation)
	return nil
}

// Save merkle commitments to the MerkleCommitment collection
func (d *DbFake) saveMerkleCommitments(commitments []models.CommitmentMerkleCommitment) error {
	var newCommitments []models.CommitmentMerkleCommitment
	for _, commitment := range commitments {
		found := false
		for _, c := range d.merkleCommitments {
			if c.MerkleRoot == commitment.MerkleRoot {
				found = true
				c = commitment
				break
			}
		}
		if !found {
			newCommitments = append(newCommitments, commitment)
		}
	}
	d.merkleCommitments = append(d.merkleCommitments, newCommitments...)
	return nil
}

// Save merkle proofs to the MerkleProof collection
func (d *DbFake) saveMerkleProofs(proofs []models.CommitmentMerkleProof) error {
	var newProofs []models.CommitmentMerkleProof
	for _, proof := range proofs {
		found := false
		for _, p := range d.merkleProofs {
			if p.MerkleRoot == proof.MerkleRoot {
				found = true
				p = proof
				break
			}
		}
		if !found {
			newProofs = append(newProofs, proof)
		}
	}
	d.merkleProofs = append(d.merkleProofs, newProofs...)
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
func (d *DbFake) SetLatestCommitments(latestCommitments []models.LatestCommitment) {
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
		return merkleCommitments, errors.New(ERROR_MERKLE_COMMITMENT_GET)
	}
	return merkleCommitments, nil
}
