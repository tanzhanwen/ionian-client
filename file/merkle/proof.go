package merkle

import (
	"errors"
	"math"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	errProofWrongFormat       = errors.New("invalid merkle proof format")
	errProofRootMismatch      = errors.New("merkle proof root mismatch")
	errProofContentMismatch   = errors.New("merkle proof content mismatch")
	errProofPositionMismatch  = errors.New("merkle proof position mismatch")
	errProofValidationFailure = errors.New("failed to validate merkle proof")
)

// Proof represents a merkle tree proof of target content, e.g. chunk or segment of file.
type Proof struct {
	// Lemma is made up of 3 parts to keep consistent with ionian-rust:
	// 1. Target content hash (leaf node).
	// 2. Hashes from bottom to top of sibing nodes.
	// 3. Root hash.
	Lemma []common.Hash `json:"lemma"`

	// Path contains flags to indicate that whether the corresponding node is on the left side.
	// All true for the left most leaf node, and all false for the right most leaf node.
	Path []bool `json:"path"`
}

func (proof *Proof) validateFormat() error {
	numSiblings := len(proof.Path)

	if numSiblings == 0 {
		if len(proof.Lemma) != 1 {
			return errProofWrongFormat
		}

		return nil
	}

	if numSiblings+2 != len(proof.Lemma) {
		return errProofWrongFormat
	}

	return nil
}

func (proof *Proof) Validate(root common.Hash, content []byte, position uint32, numLeafNodes uint32) error {
	contentHash := crypto.Keccak256Hash([]byte{prefixLeaf}, content)
	return proof.ValidateHash(root, contentHash, position, numLeafNodes)
}

func (proof *Proof) ValidateHash(root common.Hash, contentHash common.Hash, position uint32, numLeafNodes uint32) error {
	if err := proof.validateFormat(); err != nil {
		return err
	}

	// content hash mismatch
	if contentHash.Hex() != proof.Lemma[0].Hex() {
		return errProofContentMismatch
	}

	// root mismatch
	if len(proof.Lemma) > 1 && root.Hex() != proof.Lemma[len(proof.Lemma)-1].Hex() {
		return errProofRootMismatch
	}

	// validate position
	if proofPos := proof.calculateProofPosition(numLeafNodes); proofPos != position {
		return errProofPositionMismatch
	}

	// validate root by proof
	if !proof.validateRoot() {
		return errProofValidationFailure
	}

	return nil
}

func (proof *Proof) calculateProofPosition(numLeafNodes uint32) uint32 {
	var position uint32

	for i := len(proof.Path) - 1; i >= 0; i-- {
		leftSideDepth := uint32(math.Ceil(math.Log2(float64(numLeafNodes))))
		leftSideLeafNodes := uint32(math.Pow(2, float64(leftSideDepth))) / 2

		if isLeft := proof.Path[i]; isLeft {
			numLeafNodes = leftSideLeafNodes
		} else {
			position += leftSideLeafNodes
			numLeafNodes -= leftSideLeafNodes
		}
	}

	return position
}

func (proof *Proof) validateRoot() bool {
	hash := proof.Lemma[0]

	for i, isLeft := range proof.Path {
		if isLeft {
			hash = crypto.Keccak256Hash([]byte{prefixInterior}, hash.Bytes(), proof.Lemma[i+1].Bytes())
		} else {
			hash = crypto.Keccak256Hash([]byte{prefixInterior}, proof.Lemma[i+1].Bytes(), hash.Bytes())
		}
	}

	return hash.Hex() == proof.Lemma[len(proof.Lemma)-1].Hex()
}
