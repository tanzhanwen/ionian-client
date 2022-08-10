package merkle

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// To prevent second preimage attack
	prefixLeaf     = byte(0)
	prefixInterior = byte(1)
)

// Binary merkle tree node
type node struct {
	parent *node
	left   *node
	right  *node
	hash   common.Hash
	isLeft bool // left hand side of parent node, which is used for merkle proof
}

func newNode(hash common.Hash) *node {
	return &node{
		hash: hash,
	}
}

func newLeafNode(content []byte) *node {
	return &node{
		hash: crypto.Keccak256Hash([]byte{prefixLeaf}, content),
	}
}

func newInteriorNode(left, right *node) *node {
	node := &node{
		left:  left,
		right: right,
		hash:  crypto.Keccak256Hash([]byte{prefixInterior}, left.hash.Bytes(), right.hash.Bytes()),
	}

	left.parent = node
	right.parent = node

	return node
}
