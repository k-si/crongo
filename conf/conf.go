package conf

import (
	"github.com/spf13/viper"
)

var Cfg Config

type Config struct {
	HttpConfig `mapstructure:"http"`
	EtcdConfig `mapstructure:"etcd"`
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
