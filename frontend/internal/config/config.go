package config

import (
	"fmt"
	"os"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type ServiceConfig struct {
	Name string `koanf:"name"`
}

type Config struct {
	Service ServiceConfig `koanf:"service"`
	Server  struct {
		Address string `koanf:"address"`
	} `koanf:"server"`
	API struct {
		URL string `koanf:"url"`
	} `koanf:"api"`
}

func Load(path string) (Config, error) {
	k := koanf.New(".")
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return Config{}, fmt.Errorf("load frontend configuration: %w", err)
	}
	for key, env := range map[string]string{"server.address": "WEB_ADDR", "api.url": "API_URL"} {
		if value := os.Getenv(env); value != "" {
			if err := k.Set(key, value); err != nil {
				return Config{}, fmt.Errorf("override %s: %w", env, err)
			}
		}
	}
	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return Config{}, fmt.Errorf("decode frontend configuration: %w", err)
	}
	if cfg.Service.Name == "" || cfg.Server.Address == "" {
		return Config{}, fmt.Errorf("frontend service.name and server.address are required")
	}
	if cfg.API.URL == "" {
		return Config{}, fmt.Errorf("API_URL is required")
	}
	return cfg, nil
}
