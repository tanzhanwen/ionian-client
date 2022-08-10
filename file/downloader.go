package file

import (
	"os"

	"github.com/Ionian-Web3-Storage/ionian-client/node"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type DownloadOption struct {
	UploadOption
	Root string
}

func (opt *DownloadOption) BindCommand(cmd *cobra.Command) {
	cmd.Flags().StringVar(&opt.Filename, "file", "", "File name to download")
	cmd.MarkFlagRequired("file")

	cmd.Flags().StringVar(&opt.StorageNodeURL, "node", "", "Ionian storage node URL")
	cmd.MarkFlagRequired("node")

	cmd.Flags().StringVar(&opt.Root, "root", "", "Merkle root to download file")
	cmd.MarkFlagRequired("root")
}

type Downloader struct {
	opt    DownloadOption
	client *node.Client
}

func NewDownloader(opt DownloadOption) (*Downloader, error) {
	client, err := node.NewClient(opt.StorageNodeURL)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to connect to storage node")
	}

	return &Downloader{
		opt:    opt,
		client: client,
	}, nil
}

func (downloader *Downloader) Download() error {
	// Query file info from storage node
	hash := common.HexToHash(downloader.opt.Root)
	info, err := downloader.client.GetFileInfo(hash)
	if err != nil {
		return errors.WithMessage(err, "Failed to get file size from storage node")
	}

	if info == nil {
		return errors.Errorf("File not found %v", downloader.opt.Root)
	}

	logrus.WithField("file", info).Info("File found by root hash")

	if !info.Finalized {
		return errors.New("File not finalized yet")
	}

	// Check file existence before downloading
	exists, err := downloader.checkExistence(hash)
	if err != nil {
		return errors.WithMessage(err, "Failed to check file existence")
	}

	if exists {
		logrus.Info("File already exists")
		return nil
	}

	// Download segments
	if err = downloader.downloadFile(hash, int64(info.Tx.Size)); err != nil {
		return errors.WithMessage(err, "Failed to download file")
	}

	// Validate the downloaded file
	if err = downloader.validateDownloadFile(int64(info.Tx.Size)); err != nil {
		return errors.WithMessage(err, "Failed to validate downloaded file")
	}

	return nil
}

func (downloader *Downloader) checkExistence(hash common.Hash) (bool, error) {
	file, err := Open(downloader.opt.Filename)
	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, errors.WithMessage(err, "Failed to open file")
	}

	defer file.Close()

	tree, err := file.MerkleTree()
	if err != nil {
		return false, errors.WithMessage(err, "Failed to create file merkle tree")
	}

	if tree.Root().Hex() == hash.Hex() {
		return true, nil
	}

	return true, errors.Errorf("file already exists without different hash")
}

func (downloader *Downloader) downloadFile(root common.Hash, size int64) error {
	downloadingFilename := downloader.opt.Filename + ".download"

	// TODO support to download from breakpoint
	file, err := os.Create(downloadingFilename)
	if err != nil {
		return errors.WithMessage(err, "Failed to create downloading file")
	}
	defer file.Close()

	// preserve space
	if err = file.Truncate(size); err != nil {
		return errors.WithMessage(err, "Failed to truncate file to preserve space")
	}

	logrus.Info("Begin to download file from storage node")

	numChunks := numSplits(size, DefaultChunkSize)
	numSegments := numSplits(size, DefaultChunkSize*DefaultSegmentMaxChunks)

	for i := uint32(0); i < numSegments; i++ {
		startIndex := i * DefaultSegmentMaxChunks
		endIndex := startIndex + DefaultSegmentMaxChunks
		if endIndex > numChunks {
			endIndex = numChunks
		}

		// TODO download with proof and validate
		segment, err := downloader.client.DownloadSegment(root, startIndex, endIndex)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"total":           numSegments,
				"index":           i,
				"chunkStartIndex": startIndex,
				"chunkEndIndex":   endIndex,
			}).Info("Failed to download segment")
			return errors.WithMessagef(err, "Failed to download segment")
		}

		// Handle the zero paddings for the last chunk
		if i == numSegments-1 {
			if lastChunkSize := size % DefaultChunkSize; lastChunkSize > 0 {
				paddings := DefaultChunkSize - lastChunkSize
				segment = segment[0 : len(segment)-int(paddings)]
			}
		}

		offset := int64(i) * int64(DefaultChunkSize*DefaultSegmentMaxChunks)

		if _, err := file.WriteAt(segment, offset); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"total": numSegments,
				"index": i,
			}).Info("Failed to write segment to file")
			return errors.WithMessage(err, "Failed to write segment to file")
		}

		logrus.WithFields(logrus.Fields{
			"total": numSegments,
			"index": i,
		}).Debug("Segment downloaded")
	}

	logrus.Info("Completed to download file")

	return nil
}

func (downloader *Downloader) validateDownloadFile(fileSize int64) error {
	if err := os.Rename(downloader.opt.Filename+".download", downloader.opt.Filename); err != nil {
		return errors.WithMessage(err, "Failed to rename file")
	}

	file, err := Open(downloader.opt.Filename)
	if err != nil {
		return errors.WithMessage(err, "Failed to open file")
	}
	defer file.Close()

	if file.Size() != fileSize {
		return errors.Errorf("File size mismatch: expected = %v, downloaded = %v", fileSize, file.Size())
	}

	tree, err := file.MerkleTree()
	if err != nil {
		return errors.WithMessage(err, "Failed to create merkle tree")
	}

	if rootHex := tree.Root().Hex(); rootHex != downloader.opt.Root {
		return errors.Errorf("Merkle root mismatch, downloaded = %v", rootHex)
	}

	logrus.Info("Succeeded to validate the downloaded file")

	return nil
}
