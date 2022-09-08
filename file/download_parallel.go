package file

import (
	"fmt"
	"os"

	"github.com/Ionian-Web3-Storage/ionian-client/common/parallel"
	"github.com/Ionian-Web3-Storage/ionian-client/node"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const minBufSize = 8

type parallelDownloader struct {
	clients  []*node.Client
	root     common.Hash
	file     *os.File
	fileSize int64

	numChunks   uint32
	numSegments uint32
}

func newParallelDownader(clients []*node.Client, root common.Hash, file *os.File, fileSize int64) *parallelDownloader {
	return &parallelDownloader{
		clients:     clients,
		root:        root,
		file:        file,
		fileSize:    fileSize,
		numChunks:   numSplits(fileSize, DefaultChunkSize),
		numSegments: numSplits(fileSize, DefaultChunkSize*DefaultSegmentMaxChunks),
	}
}

// Download downloads segments in parallel.
func (downloader *parallelDownloader) Download() error {
	numNodes := len(downloader.clients)
	bufSize := numNodes * 2
	if bufSize < minBufSize {
		bufSize = minBufSize
	}

	return parallel.Serial(downloader, int(downloader.numSegments), numNodes, bufSize)
}

// ParallelDo implements the parallel.Interface interface.
func (downloader *parallelDownloader) ParallelDo(routine, task int) (interface{}, error) {
	startIndex := uint32(task) * DefaultSegmentMaxChunks
	endIndex := startIndex + DefaultSegmentMaxChunks
	if endIndex > downloader.numChunks {
		endIndex = downloader.numChunks
	}

	// TODO download with proof and validate
	segment, err := downloader.clients[routine].DownloadSegment(downloader.root, startIndex, endIndex)

	// remove paddings for the last chunk
	if uint32(task) == downloader.numSegments-1 && err == nil {
		if lastChunkSize := downloader.fileSize % DefaultChunkSize; lastChunkSize > 0 {
			paddings := DefaultChunkSize - lastChunkSize
			segment = segment[0 : len(segment)-int(paddings)]
		}
	}

	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"routine": routine,
			"segment": fmt.Sprintf("%v/%v", task, downloader.numSegments),
			"chunks":  fmt.Sprintf("[%v, %v)", startIndex, endIndex),
		}).Error("Failed to download segment")
	} else if logrus.IsLevelEnabled(logrus.TraceLevel) {
		logrus.WithFields(logrus.Fields{
			"routine": routine,
			"segment": fmt.Sprintf("%v/%v", task, downloader.numSegments),
			"chunks":  fmt.Sprintf("[%v, %v)", startIndex, endIndex),
		}).Trace("Succeeded to download segment")
	}

	return segment, err
}

// ParallelCollect implements the parallel.Interface interface.
func (downloader *parallelDownloader) ParallelCollect(result *parallel.Result) error {
	segment := result.Value.([]byte)
	offset := int64(result.Task) * DefaultChunkSize * DefaultSegmentMaxChunks

	n, err := downloader.file.WriteAt(segment, offset)

	if err != nil {
		return errors.WithMessagef(err, "Failed to write segment %v to file", result.Task)
	}

	if n != len(segment) {
		return errors.Errorf("Failed to write segment %v to file due to length mismatch, expected = %v, actual = %v", len(segment), n)
	}

	return nil
}
