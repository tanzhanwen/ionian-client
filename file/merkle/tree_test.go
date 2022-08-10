package merkle

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func createChunkData(i int) []byte {
	return []byte(fmt.Sprintf("chunk data - %v", i))
}

func createTreeByChunks(chunks int) *Tree {
	var builder TreeBuilder

	for i := 0; i < chunks; i++ {
		chunk := createChunkData(i)
		builder.Append(chunk)
	}

	return builder.Build()
}

func TestTreeRoot(t *testing.T) {
	assert.Equal(t, "0xfd48c947d9e6ed7b4a0be6ccbe715f2a48066bcf74ddefd52e121603c0a87467", createTreeByChunks(5).Root().Hex())
	assert.Equal(t, "0xc285fe3cecd983801bf9a6bdaceb1d5d2cab01e215f66b250573f0780b78994d", createTreeByChunks(6).Root().Hex())
	assert.Equal(t, "0x6209117e41910bb511f4c21198f34703d90129d67ccf0ac22c9b08a3358045a1", createTreeByChunks(7).Root().Hex())
}

func TestTreeProof(t *testing.T) {
	for numChunks := 1; numChunks <= 32; numChunks++ {
		tree := createTreeByChunks(numChunks)

		for i := 0; i < numChunks; i++ {
			proof := tree.ProofAt(i)
			assert.NoError(t, proof.Validate(tree.Root(), createChunkData(i), uint32(i), uint32(numChunks)))
		}
	}
}

// chunksPerSegment: 2^n
func calculateRootBySegments(chunks, chunksPerSegment int) common.Hash {
	var fileBuilder TreeBuilder

	for i := 0; i < chunks; i += chunksPerSegment {
		var segBuilder TreeBuilder

		for j := 0; j < chunksPerSegment; j++ {
			index := i + j
			if index >= chunks {
				break
			}

			segBuilder.Append(createChunkData(index))
		}

		segRoot := segBuilder.Build().Root()

		fileBuilder.AppendHash(segRoot)
	}

	return fileBuilder.Build().Root()
}

// Number of chunks in segment will not impact the merkle root
func TestRootBySegment(t *testing.T) {
	for chunks := 1; chunks <= 256; chunks++ {
		root1 := createTreeByChunks(chunks).Root()   // no segment
		root2 := calculateRootBySegments(chunks, 4)  // segment with 4 chunks
		root3 := calculateRootBySegments(chunks, 16) // segment with 16 chunks

		assert.Equal(t, root1, root2)
		assert.Equal(t, root2, root3)
	}
}
