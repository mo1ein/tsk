package handler

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/graph/task-manager/internal/metrics"
	"github.com/graph/task-manager/internal/middleware"
)

func SetupRouter(handler *TaskHandler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.MetricsMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/metrics", metrics.PrometheusHandler())
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	tasks := r.Group("/tasks")
	{
		tasks.POST("", handler.CreateTask)
		tasks.GET("", handler.ListTasks)
		tasks.GET("/:id", handler.GetTask)
		tasks.PUT("/:id", handler.UpdateTask)
		tasks.DELETE("/:id", handler.DeleteTask)
	}

	return r
}
