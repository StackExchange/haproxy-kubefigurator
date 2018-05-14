package haproxyconfigurator

import (
	"crypto/tls"
	"net/http"

	"golang.org/x/net/context"

	"github.com/coreos/etcd/client"
	"github.com/sirupsen/logrus"
)

// EtcdOptions is a struct to pass etcd connection information
type EtcdOptions struct {
	CAFile         string
	ClientCertFile string
	ClientKeyFile  string
	Hosts          []string
	Path           string
	TLSConfig      *tls.Config
}

func publish(haproxyConfig string, options EtcdOptions) {
	logger.Info("Publishing configuration")
	logger.WithFields(logrus.Fields{
		"etcd-host": options.Hosts,
		"etcd-path": options.Path,
	}).Debug("Etcd target")
	cfg := client.Config{
		Endpoints: options.Hosts,
		Transport: &http.Transport{
			TLSClientConfig: options.TLSConfig,
		},
	}
	c, err := client.New(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	kapi := client.NewKeysAPI(c)
	_, err = kapi.Set(context.Background(), options.Path, haproxyConfig, nil)
	if err != nil {
		if err == context.Canceled {
			logger.Error("Etcd request was cancelled")
		} else if err == context.DeadlineExceeded {
			logger.Error("Etcd deadline exceeded")
		} else if cerr, ok := err.(*client.ClusterError); ok {
			logger.Error("Etcd client error")
			logger.Panic(cerr.Errors)
		} else {
			logger.Error("Failed to connect to etcd endpoint")
		}
		logger.Panic(err)
	}

	logger.Info("Configuration published successfully")
}
