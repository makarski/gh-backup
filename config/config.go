package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	AccessToken  string   `json:"access_token"`
	Organisation string   `json:"organisation"`
	Languages    []string `json:"languages"`
	OnlyPrivate  bool     `json:"only_private"`
	BackupDir    string   `json:"backup_dir"`
}

func NewConfig(name string) (*Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	fileKey := filepath.Join(wd, name)
	b, err := ioutil.ReadFile(fileKey)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
