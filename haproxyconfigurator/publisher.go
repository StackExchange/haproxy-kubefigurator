package haproxyconfigurator

import (
	"golang.org/x/net/context"

	"github.com/coreos/etcd/client"
	"github.com/sirupsen/logrus"
)

var etcdHost = "http://etcd.kubernetes.home.mikenewswanger.com:2379"
var etcdPublishKey = "/service-router/haproxy-config"

func publish(haproxyConfig string) {
	logger.Info("Publishing configuration")
	logger.WithFields(logrus.Fields{
		"etcd-host": etcdHost,
		"etcd-path": etcdPublishKey,
	}).Debug("Etcd target")
	cfg := client.Config{Endpoints: []string{etcdHost}}
	c, err := client.New(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	kapi := client.NewKeysAPI(c)
	_, err = kapi.Set(context.Background(), etcdPublishKey, haproxyConfig, nil)
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
