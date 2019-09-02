// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package db

import (
	"errors"
	"mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// DbFake struct
// Implements all the high level Db methods used by attestation server
// Minimizing as much as possible reliance to mongo and testing as much
// as possible without the need for a proper mongo mock for testing
type DbFake struct {
	Attestations      []models.Attestation
	AttestationsInfo  []models.AttestationInfo
	MerkleCommitments []models.CommitmentMerkleCommitment
	MerkleProofs      []models.CommitmentMerkleProof
	latestCommitments []models.ClientCommitment
}

// Return new DbFake instance
func NewDbFake() *DbFake {
	return &DbFake{
		[]models.Attestation{},
		[]models.AttestationInfo{},
		[]models.CommitmentMerkleCommitment{},
		[]models.CommitmentMerkleProof{},
		[]models.ClientCommitment{}}
}

// Save latest attestation to Attestations
func (d *DbFake) SaveAttestation(attestation models.Attestation) error {
	for i, a := range d.Attestations {
		if a.Txid == attestation.Txid {
			d.Attestations[i] = attestation
			return nil
		}
	}
	d.Attestations = append(d.Attestations, attestation)
	return nil
}

// Save latest attestation info to AttestationsInfo
func (d *DbFake) SaveAttestationInfo(attestationInfo models.AttestationInfo) error {
	for i, a := range d.AttestationsInfo {
		if a.Txid == attestationInfo.Txid {
			d.AttestationsInfo[i] = attestationInfo
			return nil
		}
	}
	d.AttestationsInfo = append(d.AttestationsInfo, attestationInfo)
	return nil
}

// Save merkle commitments to the MerkleCommitment collection
func (d *DbFake) SaveMerkleCommitments(commitments []models.CommitmentMerkleCommitment) error {
	var newCommitments []models.CommitmentMerkleCommitment
	for _, commitment := range commitments {
		found := false
		for i, c := range d.MerkleCommitments {
			if c == commitment {
				found = true
				d.MerkleCommitments[i] = commitment
				break
			}
		}
		if !found {
			newCommitments = append(newCommitments, commitment)
		}
	}
	d.MerkleCommitments = append(d.MerkleCommitments, newCommitments...)
	return nil
}

// Save merkle proofs to the MerkleProof collection
func (d *DbFake) SaveMerkleProofs(proofs []models.CommitmentMerkleProof) error {
	var newProofs []models.CommitmentMerkleProof
	for _, proof := range proofs {
		found := false
		for i, p := range d.MerkleProofs {
			if p.Commitment == proof.Commitment &&
				p.ClientPosition == proof.ClientPosition &&
				p.MerkleRoot == proof.MerkleRoot {
				found = true
				d.MerkleProofs[i] = proof
				break
			}
		}
		if !found {
			newProofs = append(newProofs, proof)
		}
	}
	d.MerkleProofs = append(d.MerkleProofs, newProofs...)
	return nil
}

// Return attestation count with optional confirmed flag
func (d *DbFake) getAttestationCount(confirmed ...bool) (int64, error) {
	if len(confirmed) > 0 {
		count := 0
		for _, atst := range d.Attestations { // calculate count for specific confirmed/unconfirmed
			if atst.Confirmed == confirmed[0] {
				count += 1
			}
		}
		return int64(count), nil
	}
	return int64(len(d.Attestations)), nil
}

// Return latest attestation commitment hash
func (d *DbFake) GetLatestAttestationMerkleRoot(confirmed bool) (string, error) {
	count, _ := d.getAttestationCount(confirmed)
	if count == 0 {
		return "", nil
	}

	for i := len(d.Attestations) - 1; i >= 0; i-- {
		latestAttestation := d.Attestations[i]
		if latestAttestation.Confirmed == confirmed {
			return d.Attestations[i].CommitmentHash().String(), nil
		}
	}
	return "", errors.New(ErrorAttestationGet)
}

// Return Commitment from MerkleCommitment commitments for attestation with given txid hash
func (d *DbFake) getAttestationMerkleRoot(txid chainhash.Hash) (string, error) {
	// first check attestation count
	count, _ := d.getAttestationCount()
	if count == 0 {
		return "", nil
	}

	for _, attestation := range d.Attestations {
		if txid == attestation.Txid {
			return attestation.CommitmentHash().String(), nil
		}
	}
	return "", nil
}

// Return commitment for attestation with given txid
func (d *DbFake) GetAttestationMerkleCommitments(txid chainhash.Hash) ([]models.CommitmentMerkleCommitment, error) {
	// get merkle root of attestation
	merkleRoot, rootErr := d.getAttestationMerkleRoot(txid)
	if rootErr != nil {
		return []models.CommitmentMerkleCommitment{}, rootErr
	} else if merkleRoot == "" {
		return []models.CommitmentMerkleCommitment{}, nil
	}

	var MerkleCommitments []models.CommitmentMerkleCommitment
	for _, commitment := range d.MerkleCommitments {
		if commitment.MerkleRoot.String() == merkleRoot {
			MerkleCommitments = append(MerkleCommitments, commitment)
		}
	}

	return MerkleCommitments, nil
}

// Set latest commitments for testing
func (d *DbFake) SetClientCommitments(latestCommitments []models.ClientCommitment) {
	d.latestCommitments = latestCommitments
}

// Return latest commitment from fake client commitments
func (d *DbFake) GetClientCommitments() ([]models.ClientCommitment, error) {
	return d.latestCommitments, nil
}
