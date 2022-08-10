package cmd

import (
	"time"

	"github.com/Ionian-Web3-Storage/ionian-client/contract"
	"github.com/openweb3/web3go"
	"github.com/openweb3/web3go/signers"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	url            string
	key            string
	bytecodeOrFile string

	deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy Ionian contract to specified blockchain",
		Run:   deploy,
	}
)

func init() {
	deployCmd.Flags().StringVar(&url, "url", "", "Fullnode URL to interact with blockchain")
	deployCmd.MarkFlagRequired("url")
	deployCmd.Flags().StringVar(&key, "key", "", "Private key to create smart contract")
	deployCmd.MarkFlagRequired("key")
	deployCmd.Flags().StringVar(&bytecodeOrFile, "bytecode", "", "Ionian smart contract bytecode")
	deployCmd.MarkFlagRequired("bytecode")

	rootCmd.AddCommand(deployCmd)
}

func deploy(*cobra.Command, []string) {
	sm := signers.MustNewSignerManagerByPrivateKeyStrings([]string{key})

	option := new(web3go.ClientOption).
		WithRetry(3, time.Second).
		WithTimout(5 * time.Second).
		WithSignerManager(sm)

	client, err := web3go.NewClientWithOption(url, *option)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to fullnode")
	}

	contract, err := contract.Deploy(client, bytecodeOrFile)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to deploy smart contract")
	}

	logrus.WithField("contract", contract).Info("Smart contract deployed")
}
