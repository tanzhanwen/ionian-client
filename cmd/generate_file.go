package cmd

import (
	"io/ioutil"
	"math/rand"
	"os"

	"github.com/Ionian-Web3-Storage/ionian-client/file"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	size      uint64
	filename  string
	overwrite bool

	genFileCmd = &cobra.Command{
		Use:   "gen",
		Short: "Generate a temp file for test purpose",
		Run:   generateTempFile,
	}
)

func init() {
	genFileCmd.Flags().Uint64Var(&size, "size", 4*1024*1024, "File size in bytes")
	genFileCmd.Flags().StringVar(&filename, "file", "tmp123456", "File name to generate")
	genFileCmd.Flags().BoolVar(&overwrite, "overwrite", true, "Whether to overwrite existing file")

	rootCmd.AddCommand(genFileCmd)
}

func generateTempFile(*cobra.Command, []string) {
	exists, err := file.Exists(filename)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to check file existence")
	}

	if exists && !overwrite {
		logrus.Warn("File already exists")
		return
	}

	data := make([]byte, size)
	if _, err = rand.Read(data); err != nil {
		logrus.WithError(err).Fatal("Failed to generate random data")
	}

	if err = ioutil.WriteFile(filename, data, os.ModePerm); err != nil {
		logrus.WithError(err).Fatal("Failed to write file")
	}

	logrus.Info("Succeeded to write file")
}
