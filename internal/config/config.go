package config

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func Load() map[string]any {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		slog.Warn("Could not find user config dir", "err", err)
		return map[string]any{}
	}

	cfgPath := filepath.Join(cfgDir, "tome")
	os.Mkdir(cfgPath, 0755)
	viper.AddConfigPath(cfgPath)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.SetDefault("theme", "dark")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			slog.Warn("Error reading config", "err", err)
		} else {
			slog.Info("No config.yaml found, using defaults. Create one at: " + filepath.Join(cfgPath, "config.yaml"))
		}
	}

	return viper.AllSettings()
}
