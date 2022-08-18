package cmd

import (
	"io/ioutil"
	"math/rand"
	"os"
	"time"

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

	if exists {
		if !overwrite {
			logrus.Warn("File already exists")
			return
		}

		logrus.Info("Overrite file")
	}

	rand.Seed(time.Now().UnixNano())

	data := make([]byte, size)
	if n, err := rand.Read(data); err != nil {
		logrus.WithError(err).Fatal("Failed to generate random data")
	} else if n != len(data) {
		logrus.WithField("n", n).Fatal("Invalid data len")
	}

	if err = ioutil.WriteFile(filename, data, os.ModePerm); err != nil {
		logrus.WithError(err).Fatal("Failed to write file")
	}

	file, err := file.Open(filename)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to open file")
	}

	tree, err := file.MerkleTree()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to generate merkle tree")
	}

	logrus.WithField("root", tree.Root()).Info("Succeeded to write file")
}
