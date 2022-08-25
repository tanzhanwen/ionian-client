package file

import (
	"time"

	"github.com/Ionian-Web3-Storage/ionian-client/contract"
	"github.com/Ionian-Web3-Storage/ionian-client/file/merkle"
	"github.com/Ionian-Web3-Storage/ionian-client/node"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/openweb3/web3go"
	"github.com/openweb3/web3go/signers"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// maxDataSize is the maximum data size to upload on blockchain directly.
// const maxDataSize = int64(4 * 1024)

type UploadOption struct {
	Filename string

	FullnodeURL      string
	FullnodeContract string
	PrivateKey       string

	StorageNodeURL string
}

func (opt *UploadOption) BindCommand(cmd *cobra.Command) {
	cmd.Flags().StringVar(&opt.Filename, "file", "", "File name to upload")
	cmd.MarkFlagRequired("file")

	cmd.Flags().StringVar(&opt.FullnodeURL, "url", "", "Fullnode URL to interact with Ionian smart contract")
	cmd.MarkFlagRequired("url")
	cmd.Flags().StringVar(&opt.FullnodeContract, "contract", "", "Ionian smart contract to interact with")
	cmd.MarkFlagRequired("contract")
	cmd.Flags().StringVar(&opt.PrivateKey, "key", "", "Private key to interact with smart contract")
	cmd.MarkFlagRequired("key")

	cmd.Flags().StringVar(&opt.StorageNodeURL, "node", "", "Ionian storage node URL")
	cmd.MarkFlagRequired("node")
}

func (opt *UploadOption) NewIonian() (*contract.Flow, error) {
	sm := signers.MustNewSignerManagerByPrivateKeyStrings([]string{opt.PrivateKey})

	option := new(web3go.ClientOption).
		WithRetry(3, time.Second).
		WithTimout(5 * time.Second).
		WithSignerManager(sm)

	client, err := web3go.NewClientWithOption(opt.FullnodeURL, *option)
	if err != nil {
		return nil, err
	}

	addr := common.HexToAddress(opt.FullnodeContract)

	return contract.MustNewFlow(addr, client), nil
}

type Uploader struct {
	ionian *contract.Flow
	client *node.Client
}

func NewUploader(opt UploadOption) (*Uploader, error) {
	ionian, err := opt.NewIonian()
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to connect to full node")
	}

	client, err := node.NewClient(opt.StorageNodeURL)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to connect to storage node")
	}

	return &Uploader{
		ionian: ionian,
		client: client,
	}, nil
}

func NewUploaderLight(client *node.Client) *Uploader {
	return &Uploader{
		client: client,
	}
}

func (uploader *Uploader) Close() {
	if uploader.ionian != nil {
		uploader.ionian.Close()
	}

	uploader.client.Close()
}

func (uploader *Uploader) Upload(filename string) error {
	// Open file to upload
	file, err := Open(filename)
	if err != nil {
		return errors.WithMessage(err, "Failed to open file")
	}
	defer file.Close()

	if file.Size() == 0 {
		return errors.New("File is empty")
	}

	logrus.WithFields(logrus.Fields{
		"name":     file.Name(),
		"size":     file.Size(),
		"chunks":   file.NumChunks(),
		"segments": file.NumSegments(),
	}).Info("File prepared to upload")

	// Calculate file merkle root.
	tree, err := file.MerkleTree()
	if err != nil {
		return errors.WithMessage(err, "Failed to create file merkle tree")
	}
	logrus.WithField("root", tree.Root()).Info("File merkle root calculated")

	info, err := uploader.client.GetFileInfo(tree.Root())
	if err != nil {
		return errors.WithMessage(err, "Failed to get file info from storage node")
	}

	logrus.WithField("info", info).Debug("Log entry retrieved from storage node")

	if uploader.ionian == nil && info == nil {
		return errors.New("log entry not available on storage node")
	}

	// Upload small data on blockchain directly.
	// if file.Size() <= maxDataSize {
	// 	if info != nil {
	// 		return errors.New("File already exists on Ionian network")
	// 	}

	// 	return uploader.uploadSmallData(filename)
	// }

	if info != nil && info.Finalized {
		return errors.New("File already exists on Ionian network")
	}

	if info == nil {
		// Append log on blockchain
		if err = uploader.submitLogEntry(file, tree); err != nil {
			return errors.WithMessage(err, "Failed to submit log entry")
		}

		// Wait for storage node to retrieve log entry from blockchain
		if err = uploader.waitForLogEntry(tree.Root()); err != nil {
			return errors.WithMessage(err, "Failed to check if log entry available on storage node")
		}
	}

	// Upload file to storage node
	if err = uploader.uploadFile(file, tree); err != nil {
		return errors.WithMessage(err, "Failed to upload file")
	}

	// Wait for transaction finality
	if err = uploader.waitForFinality(tree.Root()); err != nil {
		return errors.WithMessage(err, "Failed to wait for transaction finality on storage node")
	}

	return nil
}

// func (uploader *Uploader) uploadSmallData(filename string) error {
// 	content, err := ioutil.ReadFile(filename)
// 	if err != nil {
// 		return errors.WithMessage(err, "Failed to read data from file")
// 	}

// 	hash, err := uploader.ionian.AppendLogWithData(content)
// 	if err != nil {
// 		return errors.WithMessage(err, "Failed to send transaction to append log with data")
// 	}

// 	logrus.WithField("hash", hash.Hex()).Info("Succeeded to send transaction to append log with data")

// 	return uploader.waitForSuccessfulExecution(hash)
// }

func (uploader *Uploader) waitForSuccessfulExecution(txHash common.Hash) error {
	logrus.WithField("tx", txHash).Info("Wait for transaction execution")

	receipt, err := uploader.ionian.WaitForReceipt(txHash)
	if err != nil {
		return errors.WithMessage(err, "Failed to wait for receipt")
	}

	if receipt.Status == nil {
		return errors.New("status not found in receipt")
	}

	switch *receipt.Status {
	case types.ReceiptStatusSuccessful:
		return nil
	case types.ReceiptStatusFailed:
		if receipt.TxExecErrorMsg == nil {
			return errors.New("Transaction execution failed")
		}

		return errors.Errorf("Transaction execution failed, %v", *receipt.TxExecErrorMsg)
	default:
		return errors.Errorf("Unknown receipt status %v", *receipt.Status)
	}
}

func (uploader *Uploader) submitLogEntry(file *File, tree *merkle.Tree) error {
	flow := NewFlow(file)
	submission, err := flow.CreateSubmission()
	if err != nil {
		return errors.WithMessage(err, "Failed to create flow submission")
	}

	// Submit log entry to smart contract.
	hash, err := uploader.ionian.Submit(*submission)
	if err != nil {
		return errors.WithMessage(err, "Failed to send transaction to append log entry")
	}

	logrus.WithField("hash", hash.Hex()).Info("Succeeded to send transaction to append log entry")

	return uploader.waitForSuccessfulExecution(hash)
}

// Wait for log entry ready on storage node.
func (uploader *Uploader) waitForLogEntry(root common.Hash) error {
	logrus.WithField("root", root).Info("Wait for log entry on storage node")

	for {
		time.Sleep(time.Second)

		info, err := uploader.client.GetFileInfo(root)
		if err != nil {
			return errors.WithMessage(err, "Failed to get file info from storage node")
		}

		if info != nil {
			break
		}
	}

	return nil
}

// TODO error tolerance
func (uploader *Uploader) uploadFile(file *File, tree *merkle.Tree) error {
	logrus.Info("Begin to upload file")

	iter := file.Iterate()
	var segIndex int

	for {
		ok, err := iter.Next()
		if err != nil {
			return errors.WithMessage(err, "Failed to read segment")
		}

		if !ok {
			break
		}

		segment := iter.Current()
		proof := tree.ProofAt(segIndex)

		segWithProof := node.SegmentWithProof{
			Root:  tree.Root(),
			Data:  segment,
			Index: uint32(segIndex),
			Proof: proof,
		}

		if _, err = uploader.client.UploadSegment(segWithProof); err != nil {
			return errors.WithMessage(err, "Failed to upload segment")
		}

		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			chunkIndex := segIndex * DefaultSegmentMaxChunks
			logrus.WithFields(logrus.Fields{
				"total":      file.NumSegments(),
				"index":      segIndex,
				"chunkStart": chunkIndex,
				"chunkEnd":   chunkIndex + len(segment)/DefaultChunkSize,
				"root":       segmentRoot(segment),
			}).Debug("Segment uploaded")
		}

		segIndex++
	}

	logrus.Info("Completed to upload file")

	return nil
}

func (uploader *Uploader) waitForFinality(root common.Hash) error {
	logrus.WithField("root", root).Info("Wait for transaction finalized on storage node")

	for {
		time.Sleep(time.Second)

		info, err := uploader.client.GetFileInfo(root)
		if err != nil {
			return errors.WithMessage(err, "Failed to get file info from storage node")
		}

		if info != nil && info.Finalized {
			break
		}
	}

	return nil
}
