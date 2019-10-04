package config

import (
	"errors"
	"flag"
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AccessToken  string   `envconfig:"github_token"`
	Organisation string   `envconfig:"organisation" required:"true"`
	Languages    []string `envconfig:"languages" required:"true"`
	OnlyPrivate  bool     `envconfig:"only_private"`
	BackupDir    string   `envconfig:"backup_dir" required:"true"`
}

// NewConfig reads the environment configs into the Config struct
func NewConfig() (*Config, error) {
	errfmt := "NewConfig error: %s"

	var cfg Config
	if err := envconfig.Process("ghbackup", &cfg); err != nil {
		return nil, fmt.Errorf(errfmt, err)
	}

	if err := parseFlags(&cfg); err != nil {
		return nil, fmt.Errorf(errfmt, err)
	}

	return &cfg, nil
}

func parseFlags(cfg *Config) error {
	flag.StringVar(&cfg.AccessToken, "token", cfg.AccessToken, "-token=GITHUB_TOKEN")
	flag.Parse()

	if cfg.AccessToken == "" {
		return errors.New("GITHUB_TOKEN is not defined")
	}

	return nil
}
