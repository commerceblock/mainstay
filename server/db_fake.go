package server

import (
	"errors"
	"mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// DbFake struct
type DbFake struct {
	attestations      []models.Attestation
	attestationsInfo  []models.AttestationInfo
	merkleCommitments []models.CommitmentMerkleCommitment
	merkleProofs      []models.CommitmentMerkleProof
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

// Save latest attestation to attestations
func (d *DbFake) saveAttestation(attestation models.Attestation) error {
	for i, a := range d.attestations {
		if a.Txid == attestation.Txid {
			d.attestations[i] = attestation
			return nil
		}
	}
	d.attestations = append(d.attestations, attestation)
	return nil
}

// Save latest attestation info to attestationsInfo
func (d *DbFake) saveAttestationInfo(attestationInfo models.AttestationInfo) error {
	for i, a := range d.attestationsInfo {
		if a.Txid == attestationInfo.Txid {
			d.attestationsInfo[i] = attestationInfo
			return nil
		}
	}
	d.attestationsInfo = append(d.attestationsInfo, attestationInfo)
	return nil
}

// Save merkle commitments to the MerkleCommitment collection
func (d *DbFake) saveMerkleCommitments(commitments []models.CommitmentMerkleCommitment) error {
	var newCommitments []models.CommitmentMerkleCommitment
	for _, commitment := range commitments {
		found := false
		for i, c := range d.merkleCommitments {
			if c == commitment {
				found = true
				d.merkleCommitments[i] = commitment
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
		for i, p := range d.merkleProofs {
			if p.Commitment == proof.Commitment &&
				p.ClientPosition == proof.ClientPosition &&
				p.MerkleRoot == proof.MerkleRoot {
				found = true
				d.merkleProofs[i] = proof
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

// Return attestation count with optional confirmed flag
func (d *DbFake) getAttestationCount(confirmed ...bool) (int64, error) {
	if len(confirmed) > 0 {
		count := 0
		for _, atst := range d.attestations { // calculate count for specific confirmed/unconfirmed
			if atst.Confirmed == confirmed[0] {
				count += 1
			}
		}
		return int64(count), nil
	}
	return int64(len(d.attestations)), nil
}

// Return latest attestation commitment hash
func (d *DbFake) getLatestAttestationMerkleRoot(confirmed bool) (string, error) {
	count, _ := d.getAttestationCount(confirmed)
	if count == 0 {
		return "", nil
	}

	for i := len(d.attestations) - 1; i >= 0; i-- {
		latestAttestation := d.attestations[i]
		if latestAttestation.Confirmed == confirmed {
			return d.attestations[i].CommitmentHash().String(), nil
		}
	}
	return "", errors.New(ERROR_ATTESTATION_GET)
}

// Return Commitment from MerkleCommitment commitments for attestation with given txid hash
func (d *DbFake) getAttestationMerkleRoot(txid chainhash.Hash) (string, error) {
	// first check attestation count
	count, _ := d.getAttestationCount()
	if count == 0 {
		return "", nil
	}

	for _, attestation := range d.attestations {
		if txid == attestation.Txid {
			return attestation.CommitmentHash().String(), nil
		}
	}
	return "", nil
}

// Return commitment for attestation with given txid
func (d *DbFake) getAttestationMerkleCommitments(txid chainhash.Hash) ([]models.CommitmentMerkleCommitment, error) {
	// get merkle root of attestation
	merkleRoot, rootErr := d.getAttestationMerkleRoot(txid)
	if rootErr != nil {
		return []models.CommitmentMerkleCommitment{}, rootErr
	} else if merkleRoot == "" {
		return []models.CommitmentMerkleCommitment{}, nil
	}

	var merkleCommitments []models.CommitmentMerkleCommitment
	for _, commitment := range d.merkleCommitments {
		if commitment.MerkleRoot.String() == merkleRoot {
			merkleCommitments = append(merkleCommitments, commitment)
		}
	}

	return merkleCommitments, nil
}

// Set latest commitments for testing
func (d *DbFake) SetClientCommitments(latestCommitments []models.ClientCommitment) {
	d.latestCommitments = latestCommitments
}

// Return latest commitment from fake client commitments
func (d *DbFake) getClientCommitments() ([]models.ClientCommitment, error) {
	return d.latestCommitments, nil
}
