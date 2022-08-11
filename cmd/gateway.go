package cmd

import (
	"strings"

	"github.com/Ionian-Web3-Storage/ionian-client/gateway"
	"github.com/spf13/cobra"
)

var (
	nodes string

	gatewayCmd = &cobra.Command{
		Use:   "gateway",
		Short: "Start gateway service",
		Run:   startGateway,
	}
)

func init() {
	gatewayCmd.Flags().StringVar(&nodes, "nodes", "http://127.0.0.1:5678,http://127.0.0.1:5679,http://127.0.0.1:5680", "Storage node list separated by comma")

	rootCmd.AddCommand(gatewayCmd)
}

func startGateway(*cobra.Command, []string) {
	gateway.MustServeLocal(strings.Split(nodes, ","))
}
