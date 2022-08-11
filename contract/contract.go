package contract

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/openweb3/web3go"
	"github.com/openweb3/web3go/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var CustomGasPrice uint64

func getGasPrice() *hexutil.Big {
	if CustomGasPrice == 0 {
		return nil
	}

	return (*hexutil.Big)(new(big.Int).SetUint64(CustomGasPrice))
}

type contract struct {
	abi     abi.ABI
	address common.Address
	client  *web3go.Client // signer hooked with from address to send transactions
}

func mustNewContract(abiJSON string, address common.Address, clientWithSigner *web3go.Client) *contract {
	var abi abi.ABI
	if err := abi.UnmarshalJSON([]byte(abiJSON)); err != nil {
		logrus.WithError(err).Fatal("Failed to unmarshal ABI")
	}

	return &contract{
		abi:     abi,
		address: address,
		client:  clientWithSigner,
	}
}

func (c *contract) Close() {
	c.client.Close()
}

func (c *contract) send(method string, args ...interface{}) (common.Hash, error) {
	data, err := c.abi.Pack(method, args...)
	if err != nil {
		return common.Hash{}, errors.WithMessage(err, "Failed to pack ABI data")
	}

	txInputData := hexutil.Bytes(data)

	from, err := defaultAccount(c.client)
	if err != nil {
		return common.Hash{}, errors.WithMessage(err, "Failed to detect account")
	}

	return c.client.Eth.SendTransactionByArgs(types.TransactionArgs{
		From:     &from,
		To:       &c.address,
		Data:     &txInputData,
		GasPrice: getGasPrice(),
	})
}

func (c *contract) WaitForReceipt(txHash common.Hash) (*types.Receipt, error) {
	return waitForReceipt(c.client, txHash)
}

func defaultAccount(clientWithSigner *web3go.Client) (common.Address, error) {
	sm, err := clientWithSigner.GetSignerManager()
	if err != nil {
		return common.Address{}, errors.WithMessage(err, "Failed to get signer manager from client")
	}

	accounts := sm.List()
	if len(accounts) == 0 {
		return common.Address{}, errors.WithMessage(err, "Account not configured in signer manager")
	}

	return accounts[0].Address(), nil
}

func waitForReceipt(clientWithSigner *web3go.Client, txHash common.Hash) (*types.Receipt, error) {
	for {
		time.Sleep(time.Second)

		receipt, err := clientWithSigner.Eth.TransactionReceipt(txHash)
		if err != nil {
			return nil, err
		}

		if receipt != nil {
			return receipt, nil
		}
	}
}

func Deploy(clientWithSigner *web3go.Client, dataOrFile string) (common.Address, error) {
	from, err := defaultAccount(clientWithSigner)
	if err != nil {
		return common.Address{}, errors.WithMessage(err, "Failed to detect account")
	}

	bytecode, err := parseBytecode(dataOrFile)
	if err != nil {
		return common.Address{}, errors.WithMessage(err, "Failed to parse bytecode")
	}

	txHash, err := clientWithSigner.Eth.SendTransactionByArgs(types.TransactionArgs{
		From:     &from,
		Data:     &bytecode,
		GasPrice: getGasPrice(),
	})
	if err != nil {
		return common.Address{}, errors.WithMessage(err, "Failed to send transaction")
	}

	logrus.WithField("hash", txHash).Info("Transaction sent to blockchain")

	receipt, err := waitForReceipt(clientWithSigner, txHash)
	if err != nil {
		return common.Address{}, errors.WithMessage(err, "Failed to wait for receipt")
	}

	if receipt.ContractAddress == nil {
		return common.Address{}, errors.Errorf("Transaction execution failed %v", receipt.Status)
	}

	return *receipt.ContractAddress, nil
}

func parseBytecode(dataOrFile string) (hexutil.Bytes, error) {
	if strings.HasPrefix(dataOrFile, "0x") {
		return hexutil.Decode(dataOrFile)
	}

	content, err := ioutil.ReadFile(dataOrFile)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to read file")
	}

	var data map[string]interface{}
	if err = json.Unmarshal(content, &data); err != nil {
		return nil, errors.WithMessage(err, "Failed to unmarshal JSON")
	}

	bytecode, ok := data["bytecode"]
	if !ok {
		return nil, errors.New("bytecode field not found in JSON file")
	}

	if bytecodeStr, ok := bytecode.(string); ok {
		return hexutil.Decode(bytecodeStr)
	}

	bytecodeObj, ok := bytecode.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid type for bytecode field")
	}

	bytecode, ok = bytecodeObj["object"]
	if !ok {
		return nil, errors.New("bytecode.object field not found in JSON file")
	}

	if bytecodeStr, ok := bytecode.(string); ok {
		return hexutil.Decode(bytecodeStr)
	}

	return nil, errors.New("invalid type for bytecode field")
}
