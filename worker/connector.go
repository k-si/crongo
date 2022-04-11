package worker

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

var Connector EtcdConnector

type EtcdConnector struct {
	cfg clientv3.Config
	cli *clientv3.Client
}

func InitEtcdConnector() (err error) {
	cfg := clientv3.Config{
		Endpoints:   Cfg.Endpoints,
		DialTimeout: time.Duration(Cfg.DialTimeOut) * time.Millisecond,
	}
	cli, err := clientv3.New(cfg)
	Connector.cfg = cfg
	Connector.cli = cli
	return
}

func (etcd EtcdConnector) Close() error {
	return Connector.cli.Close()
}