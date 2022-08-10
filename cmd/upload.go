package cmd

import (
	"github.com/Ionian-Web3-Storage/ionian-client/file"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	uploadOpt file.UploadOption

	uploadCmd = &cobra.Command{
		Use:   "upload",
		Short: "Upload file to Ionian network",
		Run:   upload,
	}
)

func init() {
	uploadOpt.BindCommand(uploadCmd)

	rootCmd.AddCommand(uploadCmd)
}

func upload(*cobra.Command, []string) {
	uploader, err := file.NewUploader(uploadOpt)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create file uploader")
	}

	if err = uploader.Upload(); err != nil {
		logrus.WithError(err).Fatal("Failed to upload file")
	}
}
