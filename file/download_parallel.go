package file

import (
	"fmt"

	"github.com/Ionian-Web3-Storage/ionian-client/common/parallel"
	"github.com/Ionian-Web3-Storage/ionian-client/file/download"
	"github.com/Ionian-Web3-Storage/ionian-client/node"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const minBufSize = 8

type SegmentDownloader struct {
	clients []*node.Client
	file    *download.DownloadingFile

	segmentOffset uint32
	numChunks     uint32
	numSegments   uint32
}

func NewSegmentDownloader(clients []*node.Client, file *download.DownloadingFile) (*SegmentDownloader, error) {
	offset := file.Metadata().Offset
	if offset%DefaultSegmentSize > 0 {
		return nil, errors.Errorf("Invalid data offset in downloading file %v", offset)
	}

	fileSize := file.Metadata().Size

	return &SegmentDownloader{
		clients: clients,
		file:    file,

		segmentOffset: uint32(offset / DefaultSegmentSize),
		numChunks:     numSplits(fileSize, DefaultChunkSize),
		numSegments:   numSplits(fileSize, DefaultSegmentSize),
	}, nil
}

// Download downloads segments in parallel.
func (downloader *SegmentDownloader) Download() error {
	numTasks := downloader.numSegments - downloader.segmentOffset
	numNodes := len(downloader.clients)
	bufSize := numNodes * 2
	if bufSize < minBufSize {
		bufSize = minBufSize
	}

	return parallel.Serial(downloader, int(numTasks), numNodes, bufSize)
}

// ParallelDo implements the parallel.Interface interface.
func (downloader *SegmentDownloader) ParallelDo(routine, task int) (interface{}, error) {
	segmentIndex := downloader.segmentOffset + uint32(task)
	startIndex := segmentIndex * DefaultSegmentMaxChunks
	endIndex := startIndex + DefaultSegmentMaxChunks
	if endIndex > downloader.numChunks {
		endIndex = downloader.numChunks
	}

	// TODO download with proof and validate
	root := downloader.file.Metadata().Root
	segment, err := downloader.clients[routine].DownloadSegment(root, startIndex, endIndex)

	// remove paddings for the last chunk
	if segmentIndex == downloader.numSegments-1 && err == nil {
		fileSize := downloader.file.Metadata().Size
		if lastChunkSize := fileSize % DefaultChunkSize; lastChunkSize > 0 {
			paddings := DefaultChunkSize - lastChunkSize
			segment = segment[0 : len(segment)-int(paddings)]
		}
	}

	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"routine": routine,
			"segment": fmt.Sprintf("%v/%v", segmentIndex, downloader.numSegments),
			"chunks":  fmt.Sprintf("[%v, %v)", startIndex, endIndex),
		}).Error("Failed to download segment")
	} else if logrus.IsLevelEnabled(logrus.TraceLevel) {
		logrus.WithFields(logrus.Fields{
			"routine": routine,
			"segment": fmt.Sprintf("%v/%v", segmentIndex, downloader.numSegments),
			"chunks":  fmt.Sprintf("[%v, %v)", startIndex, endIndex),
		}).Trace("Succeeded to download segment")
	}

	return segment, err
}

// ParallelCollect implements the parallel.Interface interface.
func (downloader *SegmentDownloader) ParallelCollect(result *parallel.Result) error {
	return downloader.file.Write(result.Value.([]byte))
}
