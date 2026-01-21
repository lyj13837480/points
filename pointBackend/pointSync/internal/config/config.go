// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	logging "pointSync/internal/logger"
	"pointSync/internal/stores/gdb"
	"strings"

	"github.com/spf13/viper"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Mysql          gdb.Config
	Kv             *KvConf          `json:"kv"`
	AnkrCfg        *AnkrCfg         `json:"ankrCfg"`
	AppLog         *logging.LogConf `json:"log"`
	KqPusherConf   *KqConf          `json:"KqPusherConf"`
	KqConsumerConf kq.KqConf        `json:"KqConsumerConf"`
}

type KvConf struct {
	Redis []*Redis `json:"redis"`
}

type Redis struct {
	Pass string `json:"pass"`
	Host string `json:"host"`
	Type string `json:"type"`
}

type AnkrCfg struct {
	ApiKey   string `json:"apiKey"`
	HttpsURL string `json:"httpsURL"`
}

type KqConf struct {
	Brokers           []string `json:"Brokers"`
	Topic             string   `json:"Topic"`
	Partitions        int32    `json:"Partitions"`
	ReplicationFactor int16    `json:"ReplicationFactor"`
	RetentionMs       string   `json:"RetentionMs"`
	UserMaxCount      int64    `json:"UserMaxCount"` // 单个消息最大用户数
	UserLimit         int64    `json:"UserLimit"`    // 每个批次最大用户数
}

func UnmarshalConfig(configFilePath string) (*Config, error) {
	viper.SetConfigFile(configFilePath)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("CNFT")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var c Config
	if err := viper.Unmarshal(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

func UnmarshalCmdConfig() (*Config, error) {
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var c Config

	if err := viper.Unmarshal(&c); err != nil {
		return nil, err
	}

	return &c, nil
}
