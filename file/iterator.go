package file

import (
	"io"
	"os"

	"github.com/pkg/errors"
)

type Iterator struct {
	file   *os.File
	buf    []byte // buffer to read data from file
	size   int    // actual data size in buffer
	offset int64  // offset to read data
}

func NewSegmentIterator(file *os.File, offset int64) *Iterator {
	return NewIterator(file, offset, DefaultChunkSize*DefaultSegmentMaxChunks)
}

func NewIterator(file *os.File, offset int64, batch int64) *Iterator {
	if batch%DefaultChunkSize > 0 {
		panic("batch size should align with chunk size")
	}

	buf := make([]byte, batch)

	return &Iterator{
		file:   file,
		buf:    buf,
		offset: offset,
	}
}

func (it *Iterator) Next() (bool, error) {
	n, err := it.file.ReadAt(it.buf, it.offset)

	it.size = n
	it.offset += int64(n)

	if err == nil {
		return true, nil
	}

	// unexpected IO error
	if !errors.Is(err, io.EOF) {
		return false, err
	}

	if n == 0 {
		return false, nil
	}

	it.paddingLastChunk()

	return true, nil
}

func (it *Iterator) paddingLastChunk() {
	if it.size == 0 || it.size == len(it.buf) {
		return
	}

	mod := it.size % DefaultChunkSize
	if mod == 0 {
		return
	}

	padded := DefaultChunkSize - mod

	for i, end := it.size, it.size+padded; i < end; i++ {
		it.buf[i] = 0
	}

	it.size += padded
}

func (it *Iterator) Current() []byte {
	return it.buf[:it.size]
}
