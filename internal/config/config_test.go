package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func setTestEnv(t *testing.T, vars map[string]string) {
	t.Helper()
	viper.Reset()
	viper.SetConfigType("env")
	// Remove all config env vars so AutomaticEnv doesn't pick up host values
	for _, k := range []string{
		"APP_ENV", "APP_DEBUG", "HTTP_HOST", "HTTP_PORT",
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"REDIS_ADDR", "REDIS_PASSWORD", "REDIS_DB", "REDIS_TTL",
	} {
		os.Unsetenv(k)
	}
	for k, v := range vars {
		os.Setenv(k, v)
	}
	t.Cleanup(func() {
		for k := range vars {
			os.Unsetenv(k)
		}
	})
}

func TestLoad(t *testing.T) {
	setTestEnv(t, map[string]string{
		"APP_ENV":        "development",
		"APP_DEBUG":      "true",
		"HTTP_HOST":      "0.0.0.0",
		"HTTP_PORT":      "8080",
		"DB_HOST":        "localhost",
		"DB_PORT":        "5432",
		"DB_USER":        "postgres",
		"DB_PASSWORD":    "postgres",
		"DB_NAME":        "taskdb",
		"DB_SSLMODE":     "disable",
		"REDIS_ADDR":     "localhost:6379",
		"REDIS_PASSWORD": "",
		"REDIS_DB":       "0",
		"REDIS_TTL":      "300",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.App.Env != "development" {
		t.Errorf("expected APP_ENV 'development', got '%s'", cfg.App.Env)
	}
	if cfg.App.Debug != true {
		t.Errorf("expected APP_DEBUG true, got %v", cfg.App.Debug)
	}
	if cfg.HTTP.Host != "0.0.0.0" {
		t.Errorf("expected HTTP_HOST '0.0.0.0', got '%s'", cfg.HTTP.Host)
	}
	if cfg.HTTP.Port != 8080 {
		t.Errorf("expected HTTP_PORT 8080, got %d", cfg.HTTP.Port)
	}
	if cfg.DB.Host != "localhost" {
		t.Errorf("expected DB_HOST 'localhost', got '%s'", cfg.DB.Host)
	}
	if cfg.DB.Port != 5432 {
		t.Errorf("expected DB_PORT 5432, got %d", cfg.DB.Port)
	}
	if cfg.DB.User != "postgres" {
		t.Errorf("expected DB_USER 'postgres', got '%s'", cfg.DB.User)
	}
	if cfg.DB.Password != "postgres" {
		t.Errorf("expected DB_PASSWORD 'postgres', got '%s'", cfg.DB.Password)
	}
	if cfg.DB.Name != "taskdb" {
		t.Errorf("expected DB_NAME 'taskdb', got '%s'", cfg.DB.Name)
	}
	if cfg.DB.SSLMode != "disable" {
		t.Errorf("expected DB_SSLMODE 'disable', got '%s'", cfg.DB.SSLMode)
	}
	if cfg.Redis.Addr != "localhost:6379" {
		t.Errorf("expected REDIS_ADDR 'localhost:6379', got '%s'", cfg.Redis.Addr)
	}
	if cfg.Redis.DB != 0 {
		t.Errorf("expected REDIS_DB 0, got %d", cfg.Redis.DB)
	}
}

func TestLoadFromEnvFile(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	os.WriteFile(dir+"/.env", []byte(
		"APP_ENV=production\n"+
			"APP_DEBUG=false\n"+
			"HTTP_HOST=127.0.0.1\n"+
			"HTTP_PORT=9090\n"+
			"DB_HOST=db.example.com\n"+
			"DB_PORT=5433\n"+
			"DB_USER=admin\n"+
			"DB_PASSWORD=secret\n"+
			"DB_NAME=mydb\n"+
			"DB_SSLMODE=require\n"+
			"REDIS_ADDR=redis.example.com:6380\n"+
			"REDIS_PASSWORD=rpass\n"+
			"REDIS_DB=2\n"+
			"REDIS_TTL=600\n",
	), 0644)

	// Unset all so AutomaticEnv doesn't override .env file values
	for _, k := range []string{
		"APP_ENV", "APP_DEBUG", "HTTP_HOST", "HTTP_PORT",
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"REDIS_ADDR", "REDIS_PASSWORD", "REDIS_DB", "REDIS_TTL",
	} {
		os.Unsetenv(k)
	}

	viper.Reset()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.App.Env != "production" {
		t.Errorf("expected APP_ENV 'production', got '%s'", cfg.App.Env)
	}
	if cfg.App.Debug != false {
		t.Errorf("expected APP_DEBUG false, got %v", cfg.App.Debug)
	}
	if cfg.HTTP.Host != "127.0.0.1" {
		t.Errorf("expected HTTP_HOST '127.0.0.1', got '%s'", cfg.HTTP.Host)
	}
	if cfg.HTTP.Port != 9090 {
		t.Errorf("expected HTTP_PORT 9090, got %d", cfg.HTTP.Port)
	}
	if cfg.DB.Host != "db.example.com" {
		t.Errorf("expected DB_HOST 'db.example.com', got '%s'", cfg.DB.Host)
	}
	if cfg.DB.User != "admin" {
		t.Errorf("expected DB_USER 'admin', got '%s'", cfg.DB.User)
	}
	if cfg.DB.Password != "secret" {
		t.Errorf("expected DB_PASSWORD 'secret', got '%s'", cfg.DB.Password)
	}
	if cfg.DB.Name != "mydb" {
		t.Errorf("expected DB_NAME 'mydb', got '%s'", cfg.DB.Name)
	}
	if cfg.DB.SSLMode != "require" {
		t.Errorf("expected DB_SSLMODE 'require', got '%s'", cfg.DB.SSLMode)
	}
	if cfg.Redis.Addr != "redis.example.com:6380" {
		t.Errorf("expected REDIS_ADDR 'redis.example.com:6380', got '%s'", cfg.Redis.Addr)
	}
	if cfg.Redis.Password != "rpass" {
		t.Errorf("expected REDIS_PASSWORD 'rpass', got '%s'", cfg.Redis.Password)
	}
	if cfg.Redis.DB != 2 {
		t.Errorf("expected REDIS_DB 2, got %d", cfg.Redis.DB)
	}
}

func TestLoadMissingEnv(t *testing.T) {
	viper.Reset()
	for _, k := range []string{
		"APP_ENV", "APP_DEBUG", "HTTP_HOST", "HTTP_PORT",
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"REDIS_ADDR", "REDIS_PASSWORD", "REDIS_DB", "REDIS_TTL",
	} {
		os.Unsetenv(k)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing env vars")
		}
	}()

	Load()
}

func TestValidate(t *testing.T) {
	viper.Set("EXISTING_VAR", "value")

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing env var")
		}
	}()

	validate("EXISTING_VAR")
	validate("NON_EXISTING_VAR")
}
