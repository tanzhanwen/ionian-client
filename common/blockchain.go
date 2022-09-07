package common

import (
	"time"

	"github.com/openweb3/web3go"
	"github.com/openweb3/web3go/signers"
	"github.com/sirupsen/logrus"
)

func MustNewWeb3(url, key string) *web3go.Client {
	client, err := NewWeb3(url, key)
	if err != nil {
		logrus.WithError(err).WithField("url", url).Fatal("Failed to connect to fullnode")
	}

	return client
}

func NewWeb3(url, key string) (*web3go.Client, error) {
	sm := signers.MustNewSignerManagerByPrivateKeyStrings([]string{key})

	option := new(web3go.ClientOption).
		WithRetry(3, time.Second).
		WithTimout(5 * time.Second).
		WithSignerManager(sm)

	return web3go.NewClientWithOption(url, *option)
}
