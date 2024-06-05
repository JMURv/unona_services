package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Port        int          `yaml:"port"`
	ServiceName string       `yaml:"serviceName"`
	DB          DBConfig     `yaml:"db"`
	Redis       RedisConfig  `yaml:"redis"`
	Kafka       KafkaConfig  `yaml:"kafka"`
	Jaeger      JaegerConfig `yaml:"jaeger"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type RedisConfig struct {
	Addr string `yaml:"addr"`
	Pass string `yaml:"pass"`
}

type KafkaConfig struct {
	Addrs             []string `yaml:"addrs"`
	NotificationTopic string   `yaml:"notificationTopic"`
}

type JaegerConfig struct {
	Sampler struct {
		Type  string `yaml:"type"`
		Param int    `yaml:"param"`
	} `yaml:"sampler"`
	Reporter struct {
		LogSpans           bool   `yaml:"LogSpans"`
		LocalAgentHostPort string `yaml:"LocalAgentHostPort"`
	} `yaml:"reporter"`
}

func LoadConfig(configName string) (*Config, error) {
	var conf Config

	data, err := os.ReadFile(fmt.Sprintf("../%v.yaml", configName))
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}
