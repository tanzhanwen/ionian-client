package contract

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/openweb3/web3go"
)

type Ionian struct {
	*contract
}

func MustNewIonian(contractAddr common.Address, clientWithSigner *web3go.Client) *Ionian {
	return &Ionian{mustNewContract(abiIonian, contractAddr, clientWithSigner)}
}

func (ionian *Ionian) AppendLog(dataRoot [32]byte, sizeBytes *big.Int) (common.Hash, error) {
	return ionian.contract.send("appendLog", dataRoot, sizeBytes)
}

func (ionian *Ionian) AppendLogWithData(data []byte) (common.Hash, error) {
	return ionian.contract.send("appendLogWithData", data)
}
