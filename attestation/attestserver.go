// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package attestation

import (
	"mainstay/db"
	"mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// AttestServer structure
// Stores information on the latest attestation and commitment
// Methods to get latest state by attestation service
type AttestServer struct {
	// underlying database interface
	dbInterface db.Db
}

// NewAttestServer returns a pointer to an AttestServer instance
func NewAttestServer(dbInterface db.Db) *AttestServer {
	return &AttestServer{dbInterface}
}

// Handle saving Commitment underlying components to the database
func (s *AttestServer) updateAttestationCommitment(commitment models.Commitment) error {
	// store merkle commitments
	merkleCommitments := commitment.GetMerkleCommitments()
	errSave := s.dbInterface.SaveMerkleCommitments(merkleCommitments)
	if errSave != nil {
		return errSave
	}

	// store merkle proofs
	merkleProofs := commitment.GetMerkleProofs()
	errSave = s.dbInterface.SaveMerkleProofs(merkleProofs)
	if errSave != nil {
		return errSave
	}

	return nil
}

// Update latest Attestation in the server
func (s *AttestServer) UpdateLatestAttestation(attestation models.Attestation) error {
	errSave := s.dbInterface.SaveAttestation(attestation)
	if errSave != nil {
		return errSave
	}
	commitment, errCommitment := attestation.Commitment()
	if errCommitment != nil {
		return errCommitment
	}
	errSave = s.updateAttestationCommitment(*commitment)
	if errSave != nil {
		return errSave
	}

	if attestation.Confirmed {
		errSave = s.dbInterface.SaveAttestationInfo(attestation.Info)
		if errSave != nil {
			return errSave
		}
	}

	return nil
}

// Return Commitment hash of latest Attestation stored in the server
func (s *AttestServer) GetLatestAttestationCommitmentHash(confirmed ...bool) (chainhash.Hash, error) {
	// optional param to set confirmed flag - looks for confirmed only by default
	confirmedParam := true
	if len(confirmed) > 0 {
		confirmedParam = confirmed[0]
	}

	// get attestation merkle root from db
	_, rootErr := s.dbInterface.GetLatestAttestationMerkleRoot(confirmedParam)
//	if rootErr != nil {
		return chainhash.Hash{}, rootErr
//	} else if merkleRoot == "" { // no attestations yet
//		return chainhash.Hash{}, nil
//	}
//	commitmentHash, errHash := chainhash.NewHashFromStr(merkleRoot)
//	if errHash != nil {
//		return chainhash.Hash{}, errHash
//	}
//	return *commitmentHash, nil
}

// Return latest commitment stored in the server
func (s *AttestServer) GetClientCommitment() (models.Commitment, error) {

	// get latest commitments from db
	latestCommitments, errLatest := s.dbInterface.GetClientCommitments()
	if errLatest != nil {
		return models.Commitment{}, errLatest
	}

	var commitmentHashes []chainhash.Hash
	if len(latestCommitments) > 0 {
		// initialise hash slice with the maximum position returned from the commitment results
		// asume latestCommitments ordered (ASC) by client position
		commitmentHashes = make([]chainhash.Hash, latestCommitments[len(latestCommitments)-1].ClientPosition+1)
		// set commitments in ordered position for resulting slice
		// missing positions have been initialized to zero hash
		for _, c := range latestCommitments {
			commitmentHashes[c.ClientPosition] = c.Commitment
		}
	}

	// construct Commitment from MerkleCommitment commitments
	commitment, errCommitment := models.NewCommitment(commitmentHashes)
	if errCommitment != nil {
		return models.Commitment{}, errCommitment
	}

	// db interface
	return *commitment, nil
}

// Return Commitment for a particular Attestation transaction id
func (s *AttestServer) GetAttestationCommitment(attestationTxid chainhash.Hash, confirmed ...bool) (models.Commitment, error) {
	// optional param to set confirmed flag - looks for confirmed only by default
	confirmedParam := true
	if len(confirmed) > 0 {
		confirmedParam = confirmed[0]
	}

	// get merkle commitments from db
	merkleCommitments, merkleCommitmentsErr := s.dbInterface.GetAttestationMerkleCommitments(attestationTxid)
	if merkleCommitmentsErr != nil {
		return models.Commitment{}, merkleCommitmentsErr
	} else if len(merkleCommitments) == 0 {
		if confirmedParam { // assume first attestation
			return models.Commitment{}, nil
		}
	}

	// construct Commitment from MerkleCommitment commitments
	var commitmentHashes []chainhash.Hash
	for _, c := range merkleCommitments {
		commitmentHashes = append(commitmentHashes, c.Commitment)
	}

	commitment, errCommitment := models.NewCommitment(commitmentHashes)
	if errCommitment != nil {
		return models.Commitment{}, errCommitment
	}

	return *commitment, nil
}
