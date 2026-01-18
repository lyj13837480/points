// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"pointSync/pointSync/internal/stores/gdb"

	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Mysql   gdb.Config
	Kv      *KvConf `json:"kv"`
	AnkrCfg AnkrCfg `json:"ankr_cfg"`
}

type KvConf struct {
	Redis []*Redis `toml:"redis" json:"redis"`
}

type Redis struct {
	Pass string `json:"pass"`
	Host string `json:"host"`
	Type string `json:"type"`
}

type AnkrCfg struct {
	ApiKey   string `json:"api_key"`
	HttpsURL string `json:"https_url"`
}
