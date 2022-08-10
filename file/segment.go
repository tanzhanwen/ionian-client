package file

import (
	"io"
	"os"

	"github.com/Ionian-Web3-Storage/ionian-client/file/merkle"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

// Segment is made up of multiple continuous chunks.
type Segment struct {
	Data       []byte // segment data
	Index      int64  // segment index that starts from 0
	ChunkIndex int64  // Chunk index of the first chunk in segment
}

func (s *Segment) Root() common.Hash {
	var builder merkle.TreeBuilder

	for offset, len := 0, len(s.Data); offset < len; offset += DefaultChunkSize {
		chunk := s.Data[offset : offset+DefaultChunkSize]
		builder.Append(chunk)
	}

	return builder.Build().Root()
}

type SegmentIterator struct {
	file   *os.File
	buf    []byte // buffer to read data from file
	size   int    // actual data size in buffer
	offset int64  // offset to read data
}

func newSegmentIterator(file *os.File) *SegmentIterator {
	buf := make([]byte, DefaultChunkSize*DefaultSegmentMaxChunks)

	return &SegmentIterator{
		file: file,
		buf:  buf,
		size: len(buf),
	}
}

func (it *SegmentIterator) Next() (bool, error) {
	if it.size < len(it.buf) {
		return false, nil
	}

	n, err := it.file.ReadAt(it.buf, it.offset)
	it.size = n
	it.offset += int64(n)

	if err == nil {
		return true, nil
	}

	if !errors.Is(err, io.EOF) {
		return false, err
	}

	if n == 0 {
		return false, nil
	}

	it.paddingLastChunk(n)

	return true, nil
}

func (it *SegmentIterator) paddingLastChunk(n int) {
	mod := n % DefaultChunkSize
	if mod == 0 {
		return
	}

	it.size += DefaultChunkSize - mod

	for i := n; i < it.size; i++ {
		it.buf[i] = 0
	}
}

func (it *SegmentIterator) Current() *Segment {
	segIndex := it.currentSegmentIndex()

	return &Segment{
		Data:       it.buf[:it.size],
		Index:      segIndex,
		ChunkIndex: segIndex * int64(DefaultSegmentMaxChunks),
	}
}

func (it *SegmentIterator) currentSegmentIndex() int64 {
	segSize := int64(DefaultChunkSize * DefaultSegmentMaxChunks)
	if it.offset%segSize == 0 {
		return it.offset/segSize - 1
	}

	return it.offset / segSize
}
