package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"

	"github.com/mo1ein/tsk/docs"
	"github.com/mo1ein/tsk/internal/handler"
	"github.com/mo1ein/tsk/internal/repository/postgres/taskrepo"
	"github.com/mo1ein/tsk/internal/repository/redis/taskcache"
	"github.com/mo1ein/tsk/internal/service/taskservice"
	"github.com/mo1ein/tsk/pkg/database"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API server",
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	dbCfg := database.Config{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		DBName:   cfg.DB.Name,
		SSLMode:  cfg.DB.SSLMode,
	}

	db, err := database.Connect(dbCfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	defer sqlDB.Close()

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name, cfg.DB.SSLMode,
	)
	if err := database.RunMigrations(dsn, "migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer rdb.Close()

	taskRepo := taskrepo.New(db)
	taskCache := taskcache.New(rdb, cfg.Redis.TTL)
	taskSvc := taskservice.New(taskRepo, taskCache)
	taskHandler := handler.NewTaskHandler(taskSvc)
	router := handler.SetupRouter(taskHandler)

	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)

	docs.SwaggerInfo.Host = addr

	log.Printf("Server starting on %s", addr)
	return http.ListenAndServe(addr, router)
}
