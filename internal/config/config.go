package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type ElasticsearchTLSConfig struct {
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type ElasticsearchConfig struct {
	Host                string                 `yaml:"host"`
	MigrationsIndexName string                 `yaml:"migrations_index_name"`
	Username            string                 `yaml:"username"`
	Password            string                 `yaml:"password"`
	TLS                 ElasticsearchTLSConfig `yaml:"tls"`
}

type Config struct {
	Elasticsearch ElasticsearchConfig `yaml:"elasticsearch"`
}

func NewConfig(configPath string) (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	if cfg.Elasticsearch.MigrationsIndexName == "" {
		cfg.Elasticsearch.MigrationsIndexName = "migrations"
	}

	return &cfg, nil
}
