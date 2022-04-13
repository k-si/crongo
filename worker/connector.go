package worker

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var (
	EtcdConn  EtcdConnector
	MongoConn MongoConnector
)

// etcd
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
	EtcdConn.cfg = cfg
	EtcdConn.cli = cli
	return
}

func (conn EtcdConnector) Close() error {
	return conn.cli.Close()
}

// mongodb
type MongoConnector struct {
	cli        *mongo.Client
	db         *mongo.Database
	collection *mongo.Collection
}

func InitMongoConnector() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(Cfg.ConnectTimeOut)*time.Millisecond)
	defer cancel()
	MongoConn.cli, err = mongo.Connect(ctx, options.Client().ApplyURI(Cfg.ApplyUri))
	MongoConn.db = MongoConn.cli.Database(Cfg.DBName)
	MongoConn.collection = MongoConn.db.Collection(Cfg.CollectionName)
	return
}

func (conn MongoConnector) Close() (err error) {
	return conn.cli.Disconnect(context.Background())
}
