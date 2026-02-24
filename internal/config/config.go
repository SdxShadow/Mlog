package config

import (
	"fmt"
	"os"

	"github.com/mlog/mlog/pkg/types"
	"github.com/spf13/viper"
)

var cfg *types.Config

func Load(path string) (*types.Config, error) {
	viper.SetConfigFile(path)
	viper.SetDefault("server.polling_interval", "1s")
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.file", "/var/log/mlog/mlog.log")
	viper.SetDefault("database.path", "/var/lib/mlog/mlog.db")
	viper.SetDefault("database.max_size_mb", 1000)
	viper.SetDefault("database.retention_days", 90)
	viper.SetDefault("ssh.enabled", true)
	viper.SetDefault("ssh.track_sessions", true)
	viper.SetDefault("security.enabled", true)
	viper.SetDefault("security.brute_force.threshold", 5)
	viper.SetDefault("security.brute_force.window_minutes", 5)
	viper.SetDefault("security.port_scan.threshold", 10)
	viper.SetDefault("security.port_scan.window_seconds", 5)
	viper.SetDefault("system.enabled", true)
	viper.SetDefault("system.journalctl", true)
	viper.SetDefault("application.enabled", true)
	viper.SetDefault("monitoring.realtime", true)
	viper.SetDefault("monitoring.buffer_size", 100)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg = &types.Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if cfg.Server.ID == "" {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}
		cfg.Server.ID = hostname
	}

	return cfg, nil
}

func Get() *types.Config {
	return cfg
}

func Save(path string, cfg *types.Config) error {
	viper.SetConfigFile(path)
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}
