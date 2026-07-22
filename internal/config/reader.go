package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Load reads configuration from environment variables and .env files.
// It panics if required environment variables are missing.
func Load() (*Config, error) {
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/task-manager")
	viper.SetConfigName(".env")
	viper.AllowEmptyEnv(true)

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	return &Config{
		App: App{
			Env:   loadString("APP_ENV"),
			Debug: loadBool("APP_DEBUG"),
		},
		HTTP: HTTP{
			Host: loadString("HTTP_HOST"),
			Port: loadInt("HTTP_PORT"),
		},
		DB: Database{
			Host:     loadString("DB_HOST"),
			Port:     loadInt("DB_PORT"),
			User:     loadString("DB_USER"),
			Password: loadString("DB_PASSWORD"),
			Name:     loadString("DB_NAME"),
			SSLMode:  loadString("DB_SSLMODE"),
		},
		Redis: Redis{
			Addr:     loadString("REDIS_ADDR"),
			Password: loadString("REDIS_PASSWORD"),
			DB:       loadInt("REDIS_DB"),
			TTL:      loadDuration("REDIS_TTL") * time.Second,
		},
	}, nil
}
