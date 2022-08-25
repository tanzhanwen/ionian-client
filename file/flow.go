package file

import (
	"math"
	"math/big"

	"github.com/Ionian-Web3-Storage/ionian-client/contract"
	"github.com/Ionian-Web3-Storage/ionian-client/file/merkle"
	"github.com/sirupsen/logrus"
)

type Flow struct {
	file *File
}

func NewFlow(file *File) *Flow {
	return &Flow{file}
}

func (flow *Flow) CreateSubmission() (*contract.Submission, error) {
	submission := contract.Submission{
		Length: big.NewInt(flow.file.Size()),
	}

	var offset int64
	for _, chunks := range flow.splitNodes() {
		node, err := flow.createNode(offset, chunks)
		if err != nil {
			return nil, err
		}
		submission.Nodes = append(submission.Nodes, *node)
		offset += chunks * DefaultChunkSize
	}

	return &submission, nil
}

// e.g. 64, 32, 1 in chunks
func (flow *Flow) splitNodes() []int64 {
	var nodes []int64

	chunks := int64(flow.file.NumChunks())

	// split from right to left
	for chunks > 0 {
		tmp := chunks & (chunks - 1) // remove the last bit 1
		nodes = append(nodes, chunks-tmp)
		chunks = tmp
	}

	// reverse
	for i, j := 0, len(nodes)-1; i < j; i, j = i+1, j-1 {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	}

	return nodes
}

func (flow *Flow) createNode(offset, chunks int64) (*contract.SubmissionNode, error) {
	batch := chunks
	if chunks > DefaultSegmentMaxChunks {
		batch = DefaultSegmentMaxChunks
	}

	return flow.createSegmentNode(offset, DefaultChunkSize*batch, DefaultChunkSize*chunks)
}

func (flow *Flow) createSegmentNode(offset, batch, size int64) (*contract.SubmissionNode, error) {
	iter := NewIterator(flow.file.underlying, offset, batch)
	var builder merkle.TreeBuilder

	for i := int64(0); i < size; {
		ok, err := iter.Next()
		if err != nil {
			return nil, err
		}

		// should always load data
		if !ok {
			logrus.WithFields(logrus.Fields{
				"offset": offset,
				"size":   size,
			}).Error("Not enough data to create submission node")
			break
		}

		segment := iter.Current()
		builder.AppendHash(segmentRoot(segment))
		i += int64(len(segment))
	}

	numChunks := size / DefaultChunkSize
	height := int64(math.Log2(float64(numChunks)))

	return &contract.SubmissionNode{
		Root:   builder.Build().Root(),
		Height: big.NewInt(height),
	}, nil
}
