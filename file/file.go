package file

import (
	"errors"
	"os"

	"github.com/Ionian-Web3-Storage/ionian-client/file/merkle"
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

func (file *File) Iterate() *SegmentIterator {
	return newSegmentIterator(file.underlying)
}

func (file *File) MerkleTree() (*merkle.Tree, error) {
	iter := newSegmentIterator(file.underlying)
	var builder merkle.TreeBuilder

	for {
		ok, err := iter.Next()
		if err != nil {
			return nil, err
		}

		if !ok {
			break
		}

		segRoot := iter.Current().Root()

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
