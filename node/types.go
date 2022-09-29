package node

import (
	"github.com/Ionian-Web3-Storage/ionian-client/file/merkle"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Status struct {
	ConnectedPeers uint `json:"connectedPeers"`
}

type Transaction struct {
	StreamIds      []*hexutil.Big `json:"streamIds"`
	Data           []byte         `json:"data"` // in-place data
	DataMerkleRoot common.Hash    `json:"dataMerkleRoot"`
	Size           uint64         `json:"size"` // file size in bytes
	Seq            uint64         `json:"seq"`
}

type FileInfo struct {
	Tx             Transaction `json:"tx"`
	Finalized      bool        `json:"finalized"`
	IsCached       bool        `json:"isCached"`
	UploadedSegNum uint32      `json:"uploadedSegNum"`
}

type SegmentWithProof struct {
	Root     common.Hash  `json:"root"`     // file merkle root
	Data     []byte       `json:"data"`     // segment data
	Index    uint32       `json:"index"`    // segment index
	Proof    merkle.Proof `json:"proof"`    // segment merkle proof
	FileSize uint64       `json:"fileSize"` // file size
}
