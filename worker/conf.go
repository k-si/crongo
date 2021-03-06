package worker

import (
	"github.com/spf13/viper"
)

var Cfg Config

type Config struct {
	AppConfig   `mapstructure:"app"`
	EtcdConfig  `mapstructure:"etcd"`
	MongoConfig `mapstructure:"mongo"`
}

type AppConfig struct {
	BalanceOptimization bool `mapstructure:"balance_optimization"`
	BalanceSleepTime    int  `mapstructure:"balance_sleep_time"`
	JobEventChanSize    int  `mapstructure:"job_event_chan_size"`
	JobPlanChanSize     int  `mapstructure:"job_plan_chan_size"`
	JobLogChanSize      int  `mapstructure:"job_log_chan_size"`
	JobLogBundleSize    int  `mapstructure:"job_log_bundle_size"`
	JobLogSendInterval  int  `mapstructure:"job_log_send_interval"`
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
