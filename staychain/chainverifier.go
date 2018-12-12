// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package staychain

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"mainstay/clients"
	"mainstay/crypto"
	"mainstay/models"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// mainstay API url consts
const (
	ApiAttestationUrl     = "/api/v1/attestation"
	ApiCommitmentUrl      = "/api/v1/commitment"
	ApiCommitmentProofUrl = "/api/v1/commitment/proof"
)

// Helper function to get response from mainstay api for url provided
func getApiResponse(url string) (map[string]interface{}, error) {
	resp, getErr := http.Get(url)
	if getErr != nil {
		return nil, &ChainVerifierError{"API request failed"}
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	var respJson map[string]interface{}
	decErr := dec.Decode(&respJson)
	if decErr != nil {
		return nil, &ChainVerifierError{"API response decoding failed"}
	}

	respMap, ok := respJson["response"]
	if !ok {
		return nil, &ChainVerifierError{fmt.Sprintf("API response decoding failed\n%v\n", respJson["error"])}
	}

	return respMap.(map[string]interface{}), nil
}

// ChainVerifierInfo struct
// Store hash and height of sidechain block attested
type ChainVerifierInfo struct {
	hash   chainhash.Hash
	height int64
}

// Hash getter
func (i *ChainVerifierInfo) Hash() chainhash.Hash {
	return i.hash
}

// Height getter
func (i *ChainVerifierInfo) Height() int64 {
	return i.height
}

// ChainVerifierError struct
type ChainVerifierError struct {
	errstr string
}

// Implement Error interface method
func (e *ChainVerifierError) Error() string {
	return e.errstr
}

// ChainVerifier struct
// Verifies that attestations are part of the staychain
// Does basic validation checks and address tweaking checks
// Verify client commitment included in attestation by proving SPV merkle proof
type ChainVerifier struct {
	sideClient   clients.SidechainClient
	apiHost      string
	cfgMain      *chaincfg.Params
	position     int
	pubkeys      []*btcec.PublicKey
	numOfSigs    int
	latestHeight int64
}

// Return new Chain Verifier instance that verifies attestations on the side chain
func NewChainVerifier(cfgMain *chaincfg.Params, side clients.SidechainClient, position int, script string, host string) ChainVerifier {

	// parse base pubkeys from multisig redeemscript of attestation service
	pubkeys, numOfSigs := crypto.ParseRedeemScript(script)

	return ChainVerifier{side, host, cfgMain, position, pubkeys, numOfSigs, 0}
}

// Basic verification for vout size and number of addresses
func verifyTxBasic(tx Tx) error {
	if len(tx.Vout) != 1 {
		return &ChainVerifierError{"Attestation TX does not have a single vout."}
	}

	if len(tx.Vout[0].ScriptPubKey.Addresses) != 1 {
		return &ChainVerifierError{"Attestation TX does not have a single address."}
	}

	return nil
}

// Verify that the transaction destination address has been generated by
// tweaking the initial multisig public keys with the correct commitment hash
// This commitment hash is provided via the mainstay API and we confirmed tweaking
func (v *ChainVerifier) verifyTxAddr(tx Tx, root string) error {
	// get target destination address from transaction
	txaddr := tx.Vout[0].ScriptPubKey.Addresses[0]
	log.Printf("txaddr: %s\n", txaddr)

	rootHash, _ := chainhash.NewHashFromStr(root)
	var tweakedPubs []*btcec.PublicKey
	commitmentBytes := rootHash.CloneBytes()

	// tweak base pubkey with commitment from api
	for _, pub := range v.pubkeys {
		tweakedPub := crypto.TweakPubKey(pub, commitmentBytes)
		tweakedPubs = append(tweakedPubs, tweakedPub)
	}
	tweakedAddr, _ := crypto.CreateMultisig(tweakedPubs, v.numOfSigs, v.cfgMain)

	// verify tweaked addr is the same as the addr in the transaction
	if tweakedAddr.String() == txaddr {
		return nil
	}

	return &ChainVerifierError{"Tweaked address does not match the transaction address"}
}

// Verify that the commitment used to generate the destination address
// includes the client commitment in the designated client position
// Proof this using an SPV merkle proof via an API call to mainstay service
func (v *ChainVerifier) verifyCommitmentProof(commitment string, root string) error {
	// get client commitment proof via api call
	respProof, respProofErr := getApiResponse(fmt.Sprintf("%s%s?position=%d&commitment=%s",
		v.apiHost, ApiCommitmentProofUrl, v.position, commitment))
	if respProofErr != nil {
		return respProofErr
	}

	log.Println()
	log.Println("Verifying merkle proof")

	// Construct CommitmentMerkleProof model from API response
	commitmentHash, _ := chainhash.NewHashFromStr(commitment)
	rootHash, _ := chainhash.NewHashFromStr(root)
	proof := models.CommitmentMerkleProof{
		MerkleRoot:     *rootHash,
		ClientPosition: int32(v.position),
		Commitment:     *commitmentHash,
	}
	var ops []models.CommitmentMerkleProofOp
	for _, op := range respProof["ops"].([]interface{}) {
		op1 := op.(map[string]interface{})
		opAppend := op1["append"].(bool)
		opCommitment, _ := chainhash.NewHashFromStr(op1["commitment"].(string))
		ops = append(ops, models.CommitmentMerkleProofOp{
			Append:     opAppend,
			Commitment: *opCommitment,
		})

	}
	proof.Ops = ops

	// Test proof of CommitmentMerkleProof received from API
	proved := models.ProveMerkleProof(proof)
	log.Println()
	if proved {
		return nil
	}
	return &ChainVerifierError{fmt.Sprintf("Could not prove client merkle commitment %s\n", commitment)}
}

// Main chainverifier method wrapping the verification process
func (v *ChainVerifier) Verify(tx Tx) (ChainVerifierInfo, error) {
	errBasic := verifyTxBasic(tx)
	if errBasic != nil {
		return ChainVerifierInfo{}, errBasic
	}

	// get attestation root commitment via api call
	respAttestation, respAttestationErr := getApiResponse(fmt.Sprintf("%s%s?txid=%s",
		v.apiHost, ApiAttestationUrl, tx.Txid))
	if respAttestationErr != nil {
		return ChainVerifierInfo{}, respAttestationErr
	}
	root := respAttestation["merkle_root"].(string)

	// first verify tx address
	errAddr := v.verifyTxAddr(tx, root)
	if errAddr != nil {
		return ChainVerifierInfo{}, errAddr
	}

	// get client commitment via api call
	respCommitment, respCommitmentErr := getApiResponse(fmt.Sprintf("%s%s?merkle_root=%s&position=%d",
		v.apiHost, ApiCommitmentUrl, root, v.position))
	if respCommitmentErr != nil { // assume no client commitment for current attestation
		return ChainVerifierInfo{}, nil
	}
	commitment := respCommitment["commitment"].(string)

	// verify commitment proof if there was a commitment
	// for this client in the current attestation transaction
	errProof := v.verifyCommitmentProof(commitment, root)
	if errProof != nil {
		return ChainVerifierInfo{}, errProof
	}

	// add commitment info in case all verification checks passed
	commitmentHash, _ := chainhash.NewHashFromStr(commitment)
	blockHeight, blockHeightErr := v.sideClient.GetBlockHeight(commitmentHash)
	if blockHeightErr != nil {
		return ChainVerifierInfo{}, blockHeightErr
	}
	info := ChainVerifierInfo{*commitmentHash, int64(blockHeight)}

	return info, nil
}
