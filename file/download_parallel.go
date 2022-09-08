package file

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/Ionian-Web3-Storage/ionian-client/node"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const defaultBufSize = 8

type downloadResult struct {
	segment []byte
	index   uint32
	err     error
}

type parallelDownloader struct {
	clients  []*node.Client
	root     common.Hash
	fileSize int64
	bufSize  uint32

	numSegments uint32
}

func newParallelDownader(clients []*node.Client, root common.Hash, fileSize int64) *parallelDownloader {
	return &parallelDownloader{
		clients:     clients,
		root:        root,
		fileSize:    fileSize,
		bufSize:     defaultBufSize,
		numSegments: numSplits(fileSize, DefaultChunkSize*DefaultSegmentMaxChunks),
	}
}

func (downloader *parallelDownloader) Download(file *os.File) error {
	segCh := make(chan uint32, downloader.bufSize)
	defer close(segCh)

	resultCh := make(chan downloadResult, downloader.bufSize)
	defer close(resultCh)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	for i := range downloader.clients {
		wg.Add(1)
		go downloader.downloadSegment(ctx, i, segCh, resultCh, &wg)
	}

	err := downloader.dispatchAndCollect(file, segCh, resultCh)

	// notify all threads to terminate
	cancel()

	// wait for all threads to terminate
	wg.Wait()

	return err
}

func (downloader *parallelDownloader) downloadSegment(ctx context.Context, threadIdx int, segCh <-chan uint32, resultCh chan<- downloadResult, wg *sync.WaitGroup) {
	defer wg.Done()

	numChunks := numSplits(downloader.fileSize, DefaultChunkSize)

	for {
		select {
		case <-ctx.Done():
			logrus.WithField("thread", threadIdx).Debug("Completed to download segments")
			return
		case segIndex := <-segCh:
			startIndex := segIndex * DefaultSegmentMaxChunks
			endIndex := startIndex + DefaultSegmentMaxChunks
			if endIndex > numChunks {
				endIndex = numChunks
			}

			// TODO download with proof and validate
			segment, err := downloader.clients[threadIdx].DownloadSegment(downloader.root, startIndex, endIndex)

			// remove paddings for the last chunk
			if segIndex == downloader.numSegments-1 && err == nil {
				if lastChunkSize := downloader.fileSize % DefaultChunkSize; lastChunkSize > 0 {
					paddings := DefaultChunkSize - lastChunkSize
					segment = segment[0 : len(segment)-int(paddings)]
				}
			}

			resultCh <- downloadResult{segment, segIndex, err}

			if err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"thread":  threadIdx,
					"segment": fmt.Sprintf("%v/%v", segIndex, downloader.numSegments),
					"chunks":  fmt.Sprintf("[%v, %v)", startIndex, endIndex),
				}).Error("Failed to download segment")

				return
			}

			if logrus.IsLevelEnabled(logrus.TraceLevel) {
				logrus.WithFields(logrus.Fields{
					"thread":  threadIdx,
					"segment": fmt.Sprintf("%v/%v", segIndex, downloader.numSegments),
					"chunks":  fmt.Sprintf("[%v, %v)", startIndex, endIndex),
				}).Trace("Succeeded to download segment")
			}
		}
	}
}

func (downloader *parallelDownloader) dispatchAndCollect(file *os.File, segCh chan<- uint32, resultCh <-chan downloadResult) error {
	// fill the buf at first
	for i := uint32(0); i < downloader.bufSize && i < downloader.numSegments; i++ {
		segCh <- i
	}

	// handle downloaded segments in sequence
	var nextSegmentIndex uint32
	cachedSegments := map[uint32][]byte{}

	// collect results
	for result := range resultCh {
		if result.err != nil {
			return errors.WithMessagef(result.err, "Failed to download segment %v", result.index)
		}

		cachedSegments[result.index] = result.segment

		for len(cachedSegments[nextSegmentIndex]) > 0 {
			// write segment into db in sequence
			offset := int64(nextSegmentIndex) * DefaultChunkSize * DefaultSegmentMaxChunks
			if _, err := file.WriteAt(result.segment, offset); err != nil {
				logrus.WithError(err).WithField("index", nextSegmentIndex).Info("Failed to write segment to file")
				return errors.WithMessagef(err, "Failed to write segment %v to file", nextSegmentIndex)
			}

			// dispatch new segment index to download
			if newSeg := nextSegmentIndex + downloader.bufSize; newSeg < downloader.numSegments {
				segCh <- newSeg
			}

			// clear cache and move forward
			delete(cachedSegments, nextSegmentIndex)
			nextSegmentIndex++
		}

		if nextSegmentIndex >= downloader.numSegments {
			break
		}
	}

	return nil
}
