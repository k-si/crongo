package master

import (
	"github.com/spf13/viper"
)

var Cfg Config

type Config struct {
	HttpConfig  `mapstructure:"http"`
	EtcdConfig  `mapstructure:"etcd"`
	MongoConfig `mapstructure:"mongo"`
}

type HttpConfig struct {
	Port         int `mapstructure:"port"`
	ReadTimeOut  int `mapstructure:"read_time_out"`
	WriteTimeOut int `mapstructure:"write_time_out"`
}

type EtcdConfig struct {
	Endpoints   []string `mapstructure:"endpoints"`
	DialTimeOut int      `mapstructure:"dial_time_out"`
}

type MongoConfig struct {
	ApplyUri       string `mapstructure:"apply_uri"`
	ConnectTimeOut int    `mapstructure:"connect_time_out"`
	DBName         string `mapstructure:"db_name"`
	CollectionName string `mapstructure:"collection_name"`
}

func InitConfig(path string) (err error) {
	viper.SetConfigFile(path)
	if err = viper.ReadInConfig(); err != nil {
		return
	}
	if err = viper.Unmarshal(&Cfg); err != nil {
		return
	}
	return
}
