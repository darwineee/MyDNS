package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"runtime"
)

type MemcachedConfig struct {
	Servers  string `yaml:"servers"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type UDPConfig struct {
	PkgLimitRFC1035 int `yaml:"pkg_limit_rfc1035"`
	PkgLimitEDNS0   int `yaml:"pkg_limit_edns0"`
}

type ServerConfig struct {
	Port              int    `yaml:"port"`
	Protocol          string `yaml:"protocol"`
	Workers           int    `yaml:"workers"`
	ForwardHost       string `yaml:"forward_host"`
	ForwardingTimeout int    `yaml:"forwarding_timeout_milliseconds"`
	EventQueueSize    int    `yaml:"event_queue_size"`
	EventQueueTimeout int    `yaml:"event_queue_timeout_milliseconds"`
}

type Config struct {
	Memcached MemcachedConfig `yaml:"memcached"`
	UDP       UDPConfig       `yaml:"udp"`
	Sever     ServerConfig    `yaml:"server"`
}

func Load() *Config {
	config := &Config{
		Memcached: MemcachedConfig{
			Servers:  "127.0.0.1:11211",
			Username: "admin",
			Password: "admin",
		},
		UDP: UDPConfig{
			PkgLimitRFC1035: 512,
			PkgLimitEDNS0:   4096,
		},
		Sever: ServerConfig{
			Port:              53,
			Protocol:          "udp",
			Workers:           runtime.NumCPU() * 2,
			ForwardHost:       "1.1.1.1:53",
			ForwardingTimeout: 1000,
			EventQueueSize:    1000,
			EventQueueTimeout: 500,
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
