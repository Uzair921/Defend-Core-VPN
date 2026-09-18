package gui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	ServerHost   string `json:"server_host"`
	ServerPort   int    `json:"server_port"`
	ServerPubKey string `json:"server_public_key"`
	UserID       string `json:"user_id"`
	DeviceID     string `json:"device_id"`
	Username     string `json:"username"`
	SavePassword bool   `json:"save_password"`
	AutoConnect  bool   `json:"auto_connect"`
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".defendcore", "client.json")
}

func LoadConfig() (*Config, error) {
	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return &Config{
			ServerHost: "192.168.174.132",
			ServerPort: 51820,
		}, nil
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	path := configPath()
	os.MkdirAll(filepath.Dir(path), 0700)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
