// Package config provides application configuration loaded from environment
// variables and .env files using viper.
package config

import "time"

// Config holds all application configuration.
type Config struct {
	App   App
	HTTP  HTTP
	DB    Database
	Redis Redis
}

// App holds application-level settings.
type App struct {
	Env   string // Environment name (e.g. "development", "production").
	Debug bool   // Enable debug mode.
}

// HTTP holds the HTTP server configuration.
type HTTP struct {
	Host string // Bind address.
	Port int    // Listen port.
}

// Database holds PostgreSQL connection configuration.
type Database struct {
	Host     string // Database host.
	Port     int    // Database port.
	User     string // Database user.
	Password string // Database password.
	Name     string // Database name.
	SSLMode  string // SSL mode (disable, require, etc.).
}

// Redis holds Redis connection configuration.
type Redis struct {
	Addr     string        // Redis address (host:port).
	Password string        // Redis password.
	DB       int           // Redis database number.
	TTL      time.Duration // Default cache TTL.
}
