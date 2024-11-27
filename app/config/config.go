package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"runtime"
)

type RedisConfig struct {
	Host     string `yaml:"host"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type UDPConfig struct {
	PkgLimitRFC1035 int `yaml:"pkg_limit_rfc1035"`
	PkgLimitEDNS0   int `yaml:"pkg_limit_edns0"`
}

type ServerConfig struct {
	Port              int    `yaml:"port"`
	Protocol          string `yaml:"protocol"`
	Workers           int    `yaml:"workers"`
	EventQueueSize    int    `yaml:"event_queue_size"`
	EventQueueTimeout int    `yaml:"event_queue_timeout_milliseconds"`
	CacheTTLSec       uint32 `yaml:"cache_ttl_seconds"`
}

type Config struct {
	Redis  RedisConfig  `yaml:"memcached"`
	UDP    UDPConfig    `yaml:"udp"`
	Server ServerConfig `yaml:"server"`
}

func Load() *Config {
	config := &Config{
		Redis: RedisConfig{
			Host:     "127.0.0.1:6379",
			Password: "",
			DB:       0,
		},
		UDP: UDPConfig{
			PkgLimitRFC1035: 512,
			PkgLimitEDNS0:   4096,
		},
		Server: ServerConfig{
			Port:              2053,
			Protocol:          "udp",
			Workers:           runtime.NumCPU() * 2,
			EventQueueSize:    1000,
			EventQueueTimeout: 500,
			CacheTTLSec:       300,
		},
	}
	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		fmt.Printf("Error reading YAML file: %v\n", err)
		return config
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		fmt.Printf("Error parsing YAML file: %v\n", err)
	}
	return config
}
