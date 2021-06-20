package config

import (
	"fmt"
	"gen/log"
	"gopkg.in/ini.v1"
	"os"
)

const (
	Dev  = "development"
	Prod = "production"
	Test = "test"
)

var (
	Config *Cfg
)

type Cfg struct {
	Raw      *ini.File
	Env      string
	HttpAddr string
	HttpPort string
}

func NewConfig() *Cfg {
	if Config != nil {
		return Config
	}
	Config = &Cfg{
		Raw:      ini.Empty(),
		Env:      Dev,
		HttpAddr: "127.0.0.1",
		HttpPort: "8080",
	}
	return Config
}

func (cfg *Cfg) Load(cfgFile string) error {
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		return fmt.Errorf("cfg file [%s] not existed", cfgFile)
	}
	conf, err := ini.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load file [%s] failed", cfgFile)
	}
	cfg.Raw = conf

	cfg.readAppCfg()

	log.Configure(cfg.Raw) // configure log

	return nil
}

// readAppCfg 读取APP配置
func (cfg *Cfg) readAppCfg() {
	appConfig := cfg.Raw.Section("app")

	env := appConfig.Key("env").String()
	if env != "" {
		cfg.Env = env
	}

	httpPort := appConfig.Key("http_port").String()
	if httpPort != "" {
		cfg.HttpPort = httpPort
	}

	httpAddr := appConfig.Key("http_addr").String()
	if httpAddr != "" {
		cfg.HttpAddr = httpAddr
	}
}