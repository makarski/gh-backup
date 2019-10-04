package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AccessToken  string   `envconfig:"github_token" required:"true"`
	Organisation string   `envconfig:"organisation" required:"true"`
	Languages    []string `envconfig:"languages" required:"true"`
	OnlyPrivate  bool     `envconfig:"only_private"`
	BackupDir    string   `envconfig:"backup_dir" required:"true"`
}

// NewConfig reads the environment configs into the Config struct
func NewConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("ghbackup", &cfg); err != nil {
		return nil, fmt.Errorf("NewConfig error: %s", err)
	}

	return &cfg, nil
}
