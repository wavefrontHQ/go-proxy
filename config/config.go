package config

import (
	"log"

	"github.com/spf13/viper"
)

const (
	DefaultFlushThreads      = 4
	DefaultFlushInterval     = 1000
	DefaultFlushMaxPoints    = 40000
	DefaultMemoryBufferLimit = 640000
)

type ProxyConfig struct {
	Server                string
	Hostname              string
	Token                 string
	PushListenerPorts     string
	OpenTSDBPorts         string
	FlushThreads          int
	PushFlushInterval     int
	PushFlushMaxPoints    int
	PushMemoryBufferLimit int
	IdFile                string
	LogFile               string
	PprofAddr             string
}

func LoadConfig(filename string) (*ProxyConfig, error) {
	log.Println("Loading configuration from", filename)

	viper.SetConfigType("properties")
	viper.SetConfigFile(filename)

	err := viper.ReadInConfig()
	if err != nil {
		return &ProxyConfig{}, err
	}

	proxyConfig := &ProxyConfig{}
	err = viper.Unmarshal(&proxyConfig)
	if err != nil {
		return proxyConfig, err
	}
	setDefaults(proxyConfig)
	return proxyConfig, nil
}

func setDefaults(cfg *ProxyConfig) {
	if cfg.FlushThreads == 0 {
		cfg.FlushThreads = DefaultFlushThreads
	}

	if cfg.PushFlushInterval == 0 {
		cfg.PushFlushInterval = DefaultFlushInterval
	}

	if cfg.PushFlushMaxPoints == 0 {
		cfg.PushFlushMaxPoints = DefaultFlushMaxPoints
	}

	if cfg.PushMemoryBufferLimit == 0 {
		cfg.PushMemoryBufferLimit = DefaultMemoryBufferLimit
	}
}
