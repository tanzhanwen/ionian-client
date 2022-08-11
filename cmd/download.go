package cmd

import (
	"github.com/Ionian-Web3-Storage/ionian-client/file"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	downloadOpt file.DownloadOption

	downloadCmd = &cobra.Command{
		Use:   "download",
		Short: "Download file from Ionian network",
		Run:   download,
	}
)

func init() {
	downloadOpt.BindCommand(downloadCmd)

	rootCmd.AddCommand(downloadCmd)
}

func download(*cobra.Command, []string) {
	downloader, err := file.NewDownloader(downloadOpt.StorageNodeURL)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create file downloader")
	}
	defer downloader.Close()

	if err = downloader.Download(downloadOpt.Root, downloadOpt.Filename); err != nil {
		logrus.WithError(err).Fatal("Failed to download file")
	}
}
