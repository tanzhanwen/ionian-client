package contract

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/openweb3/web3go"
	"github.com/sirupsen/logrus"
)

type Flow struct {
	*contract
}

func MustNewFlow(contractAddr common.Address, clientWithSigner *web3go.Client) *Flow {
	return &Flow{mustNewContract(abiFlow, contractAddr, clientWithSigner)}
}

func (flow *Flow) Submit(submission Submission) (common.Hash, error) {
	logrus.WithField("submission", submission).Debug("Begin to submit flow data to blockchain")
	return flow.contract.send("submit", submission)
}

type SubmissionNode struct {
	Root   [32]byte
	Height *big.Int // sub-tree height of this node
}

type Submission struct {
	Length *big.Int // file size
	Nodes  []SubmissionNode
}

func (sub Submission) String() string {
	var heights []uint64
	for _, v := range sub.Nodes {
		heights = append(heights, v.Height.Uint64())
	}

	return fmt.Sprintf("{ Size: %v, Heights: %v }", sub.Length, heights)
}
