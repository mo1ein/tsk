package config

import "time"

type Config struct {
	App    App
	HTTP   HTTP
	DB     Database
	Redis  Redis
}

type App struct {
	Env   string
	Debug bool
}

type HTTP struct {
	Host string
	Port int
}

type Database struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type Redis struct {
	Addr     string
	Password string
	DB       int
	TTL      time.Duration
}
