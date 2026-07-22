package database

import (
	"fmt"
	"os"
	"testing"
)

func TestConfig_Fields(t *testing.T) {
	cfg := Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "secret",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	if cfg.Host != "localhost" {
		t.Errorf("expected host 'localhost', got '%s'", cfg.Host)
	}
	if cfg.Port != 5432 {
		t.Errorf("expected port 5432, got %d", cfg.Port)
	}
	if cfg.User != "postgres" {
		t.Errorf("expected user 'postgres', got '%s'", cfg.User)
	}
	if cfg.Password != "secret" {
		t.Errorf("expected password 'secret', got '%s'", cfg.Password)
	}
	if cfg.DBName != "testdb" {
		t.Errorf("expected dbname 'testdb', got '%s'", cfg.DBName)
	}
	if cfg.SSLMode != "disable" {
		t.Errorf("expected sslmode 'disable', got '%s'", cfg.SSLMode)
	}
}

func TestConnect_InvalidDSN(t *testing.T) {
	// Suppress GORM's error logging during this expected-failure test
	null, _ := os.Open(os.DevNull)
	defer null.Close()
	origStderr := os.Stderr
	os.Stderr = null
	defer func() { os.Stderr = origStderr }()

	_, err := Connect(Config{
		Host:     "invalid-host-that-does-not-exist",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "nonexistent",
		SSLMode:  "disable",
	})
	if err == nil {
		t.Error("expected error for invalid DSN")
	}
}

func TestConnect_Integration(t *testing.T) {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	db, err := Connect(Config{
		Host:     host,
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "taskdb",
		SSLMode:  "disable",
	})
	if err != nil {
		t.Skipf("skipping integration test: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Skipf("skipping integration test: %v", err)
	}

	if sqlDB.Stats().MaxOpenConnections == 0 {
		t.Error("expected MaxOpenConnections > 0")
	}
}

func TestRunMigrations_Integration(t *testing.T) {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	dsn := fmt.Sprintf(
		"postgres://postgres:postgres@%s:5432/taskdb?sslmode=disable",
		host,
	)

	err := RunMigrations(dsn, "../../migrations")
	if err != nil {
		t.Skipf("skipping integration test: %v", err)
	}
}
