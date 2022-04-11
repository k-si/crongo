package worker

import (
	"github.com/spf13/viper"
)

var Cfg Config

type Config struct {
	EtcdConfig `mapstructure:"etcd"`
}

type EtcdConfig struct {
	Endpoints   []string `mapstructure:"endpoints"`
	DialTimeOut int      `mapstructure:"dial_time_out"`
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
