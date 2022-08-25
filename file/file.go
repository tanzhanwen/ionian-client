package file

import (
	"errors"
	"os"

	"github.com/Ionian-Web3-Storage/ionian-client/file/merkle"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// DefaultChunkSize represents the default chunk size in bytes.
	DefaultChunkSize = 256

	// DefaultSegmentMaxChunks represents the default maximum number of chunks within a segment.
	DefaultSegmentMaxChunks = 1024
)

// ErrFileRequired is returned when manipulate on a folder.
var ErrFileRequired = errors.New("file required")

type File struct {
	os.FileInfo
	underlying *os.File
}

func Exists(name string) (bool, error) {
	file, err := os.Open(name)
	if os.IsNotExist(err) {
		return false, nil
	}

	defer file.Close()

	return true, err
}

func Open(name string) (*File, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, ErrFileRequired
	}

	return &File{
		FileInfo:   info,
		underlying: file,
	}, nil
}

func (file *File) Close() error {
	return file.underlying.Close()
}

func (file *File) NumChunks() uint32 {
	return numSplits(file.Size(), DefaultChunkSize)
}

func (file *File) NumSegments() uint32 {
	return numSplits(file.Size(), DefaultChunkSize*DefaultSegmentMaxChunks)
}

func (file *File) Iterate() *Iterator {
	return NewSegmentIterator(file.underlying, 0)
}

func (file *File) MerkleTree() (*merkle.Tree, error) {
	iter := file.Iterate()
	var builder merkle.TreeBuilder

	for {
		ok, err := iter.Next()
		if err != nil {
			return nil, err
		}

		if !ok {
			break
		}

		segRoot := segmentRoot(iter.Current())

		builder.AppendHash(segRoot)
	}

	return builder.Build(), nil
}

func numSplits(total int64, unit int) uint32 {
	if total%int64(unit) == 0 {
		return uint32(total / int64(unit))
	}

	return uint32(total/int64(unit)) + 1
}

func segmentRoot(chunks []byte) common.Hash {
	dataLen := len(chunks)
	if dataLen == 0 {
		return common.Hash{}
	}

	var builder merkle.TreeBuilder

	for offset := 0; offset < dataLen; offset += DefaultChunkSize {
		chunk := chunks[offset : offset+DefaultChunkSize]
		builder.Append(chunk)
	}

	return builder.Build().Root()
}
