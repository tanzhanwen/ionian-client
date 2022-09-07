package cmd

import (
	"github.com/Ionian-Web3-Storage/ionian-client/file"
	"github.com/Ionian-Web3-Storage/ionian-client/node"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	downloadArgs struct {
		file string
		node string
		root string
	}

	downloadCmd = &cobra.Command{
		Use:   "download",
		Short: "Download file from Ionian network",
		Run:   download,
	}
)

func init() {
	downloadCmd.Flags().StringVar(&downloadArgs.file, "file", "", "File name to download")
	downloadCmd.MarkFlagRequired("file")
	downloadCmd.Flags().StringVar(&downloadArgs.node, "node", "", "Ionian storage node URL")
	downloadCmd.MarkFlagRequired("node")
	downloadCmd.Flags().StringVar(&downloadArgs.root, "root", "", "Merkle root to download file")
	downloadCmd.MarkFlagRequired("root")

	rootCmd.AddCommand(downloadCmd)
}

func download(*cobra.Command, []string) {
	node := node.MustNewClient(downloadArgs.node)
	defer node.Close()

	downloader := file.NewDownloader(node)

	if err := downloader.Download(downloadArgs.root, downloadArgs.file); err != nil {
		logrus.WithError(err).Fatal("Failed to download file")
	}
}
