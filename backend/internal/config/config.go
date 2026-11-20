package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type ServiceConfig struct {
	Name     string `koanf:"name"`
	Version  string `koanf:"version"`
	Revision string `koanf:"revision"`
}

type Config struct {
	Service ServiceConfig `koanf:"service"`
	Server  struct {
		Address string `koanf:"address"`
	} `koanf:"server"`
	Metrics struct {
		Address string `koanf:"address"`
	} `koanf:"metrics"`
}

func Load(path string) (Config, error) {
	k := koanf.New(".")
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return Config{}, fmt.Errorf("load backend configuration: %w", err)
	}

	for key, env := range map[string]string{
		"server.address": "API_ADDR", "metrics.address": "METRICS_ADDR",
		"service.version": "APP_VERSION", "service.revision": "APP_REVISION",
	} {
		if value := os.Getenv(env); value != "" {
			if err := k.Set(key, value); err != nil {
				return Config{}, fmt.Errorf("override %s: %w", env, err)
			}
		}
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return Config{}, fmt.Errorf("decode backend configuration: %w", err)
	}

	if cfg.Service.Name == "" || cfg.Server.Address == "" || cfg.Metrics.Address == "" {
		return Config{}, errors.New("backend service.name, server.address and metrics.address are required")
	}

	return cfg, nil
}
