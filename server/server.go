package server

import (
	"errors"
	"fmt"

	"mainstay/models"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// error consts
const (
	ERROR_LATEST_COMMITMENT_MISSING = "Client commitment missing for position"
)

// Server structure
// Stores information on the latest attestation and commitment
// Methods to get latest state by attestation service
type Server struct {
	dbInterface Db
}

// NewServer returns a pointer to an Server instance
func NewServer(dbInterface Db) *Server {
	return &Server{dbInterface}
}

// Handle saving Commitment underlying components to the database
func (s *Server) updateAttestationCommitment(commitment models.Commitment) error {
	// store merkle commitments
	merkleCommitments := commitment.GetMerkleCommitments()
	errSave := s.dbInterface.saveMerkleCommitments(merkleCommitments)
	if errSave != nil {
		return errSave
	}

	// store merkle proofs
	merkleProofs := commitment.GetMerkleProofs()
	errSave = s.dbInterface.saveMerkleProofs(merkleProofs)
	if errSave != nil {
		return errSave
	}

	return nil
}

// Update latest Attestation in the server
func (s *Server) UpdateLatestAttestation(attestation models.Attestation) error {

	errSave := s.dbInterface.saveAttestation(attestation)
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
	return nil
}

// Return Commitment hash of latest Attestation stored in the server
func (s *Server) GetLatestAttestationCommitmentHash() (chainhash.Hash, error) {

	// get attestation merkle root from db
	merkleRoot, rootErr := s.dbInterface.getLatestAttestationMerkleRoot()
	if rootErr != nil {
		return chainhash.Hash{}, rootErr
	} else if merkleRoot == "" { // no attestations yet
		return chainhash.Hash{}, nil
	}
	commitmentHash, errHash := chainhash.NewHashFromStr(merkleRoot)
	if errHash != nil {
		return chainhash.Hash{}, errHash
	}
	return *commitmentHash, nil
}

// Return latest commitment stored in the server
func (s *Server) GetClientCommitment() (models.Commitment, error) {

	// get latest commitments from db
	latestCommitments, errLatest := s.dbInterface.getClientCommitments()
	if errLatest != nil {
		return models.Commitment{}, errLatest
	}

	var commitmentHashes []chainhash.Hash
	// assume latest commitments ordered by position
	for pos, c := range latestCommitments {
		if int32(pos) == c.ClientPosition {
			commitmentHashes = append(commitmentHashes, c.Commitment)
		} else {
			return models.Commitment{}, errors.New(fmt.Sprintf("%s %d", ERROR_LATEST_COMMITMENT_MISSING, pos))
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
func (s *Server) GetAttestationCommitment(attestationTxid chainhash.Hash) (models.Commitment, error) {

	// get merkle commitments from db
	merkleCommitments, merkleCommitmentsErr := s.dbInterface.getAttestationMerkleCommitments(attestationTxid)

	if merkleCommitmentsErr != nil {
		return models.Commitment{}, merkleCommitmentsErr
	} else if len(merkleCommitments) == 0 {
		return models.Commitment{}, nil
	}

	// construct Commitment from MerkleCommitment commitments
	var commitmentHashes []chainhash.Hash
	for _, c := range merkleCommitments {
		commitmentHashes = append(commitmentHashes, c.Commitment)
	}

	commitment, errCommitment := models.NewCommitment(commitmentHashes)
	if errCommitment != nil {
		return models.Commitment{}, nil
	}

	return *commitment, nil
}
