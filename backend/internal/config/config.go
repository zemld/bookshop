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
	Name string `koanf:"name"`
}

type Config struct {
	Service ServiceConfig `koanf:"service"`
	Server  struct {
		Address string `koanf:"address"`
	} `koanf:"server"`
}

func Load(path string) (Config, error) {
	k := koanf.New(".")
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return Config{}, fmt.Errorf("load backend configuration: %w", err)
	}

	if value := os.Getenv("API_ADDR"); value != "" {
		if err := k.Set("server.address", value); err != nil {
			return Config{}, fmt.Errorf("override API_ADDR: %w", err)
		}
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return Config{}, fmt.Errorf("decode backend configuration: %w", err)
	}

	if cfg.Service.Name == "" || cfg.Server.Address == "" {
		return Config{}, errors.New("backend service.name and server.address are required")
	}

	return cfg, nil
}
