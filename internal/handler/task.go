// Package handler implements HTTP handlers for the task API.
package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mo1ein/tsk/internal/constants"
	"github.com/mo1ein/tsk/internal/domains"
)

// TaskService defines the interface for task business operations.
type TaskService interface {
	Create(ctx context.Context, task *domains.Task) (*domains.Task, error)
	GetByID(ctx context.Context, id int64) (*domains.Task, error)
	List(ctx context.Context, filter domains.ListFilter) ([]domains.Task, int64, error)
	Update(ctx context.Context, task *domains.Task) (*domains.Task, error)
	Delete(ctx context.Context, id int64) error
}

// TaskHandler handles HTTP requests for task operations.
type TaskHandler struct {
	svc TaskService
}

// NewTaskHandler creates a new task HTTP handler.
func NewTaskHandler(svc TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

// CreateTaskRequest is the request body for creating a task.
type CreateTaskRequest struct {
	Title    string `json:"title" binding:"required" example:"Buy groceries"`
	Assignee string `json:"assignee" example:"alice"`
}

// UpdateTaskRequest is the request body for updating a task.
type UpdateTaskRequest struct {
	Title    *string `json:"title,omitempty" example:"Buy groceries"`
	Assignee *string `json:"assignee,omitempty" example:"alice"`
	Status   *string `json:"status,omitempty" example:"done"`
}

// TaskResponse is the JSON response for a task.
type TaskResponse struct {
	ID        int64                `json:"id" example:"1"`
	Title     string               `json:"title" example:"Buy groceries"`
	Assignee  string               `json:"assignee" example:"alice"`
	Status    constants.TaskStatus `json:"status" example:"pending"`
	CreatedAt string               `json:"created_at" example:"2025-07-20T12:00:00Z"`
	UpdatedAt string               `json:"updated_at" example:"2025-07-20T12:00:00Z"`
}

func toResponse(t *domains.Task) TaskResponse {
	return TaskResponse{
		ID:        t.ID,
		Title:     t.Title,
		Assignee:  t.Assignee,
		Status:    t.Status,
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// CreateTask godoc
// @Summary      Create a new task
// @Description  Create a new task with title and optional assignee
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        request body CreateTaskRequest true "Task to create"
// @Success      201  {object}  TaskResponse
// @Failure      400  {object}  map[string]string
// @Router       /tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := &domains.Task{
		Title:    req.Title,
		Assignee: req.Assignee,
	}

	created, err := h.svc.Create(c.Request.Context(), task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, toResponse(created))
}

// GetTask godoc
// @Summary      Get a task by ID
// @Description  Get a task by its ID
// @Tags         tasks
// @Produce      json
// @Param        id path int true "Task ID"
// @Success      200  {object}  TaskResponse
// @Failure      404  {object}  map[string]string
// @Router       /tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	task, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domains.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get task"})
		return
	}

	c.JSON(http.StatusOK, toResponse(task))
}

// ListTasks godoc
// @Summary      List tasks
// @Description  List tasks with optional filtering and pagination
// @Tags         tasks
// @Produce      json
// @Param        status query string false "Filter by status"
// @Param        assignee query string false "Filter by assignee"
// @Param        page query int false "Page number" default(1)
// @Param        page_size query int false "Page size" default(20)
// @Success      200  {object}  map[string]interface{}
// @Router       /tasks [get]
func (h *TaskHandler) ListTasks(c *gin.Context) {
	filter := domains.ListFilter{
		Status:   c.Query("status"),
		Assignee: c.Query("assignee"),
	}

	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			filter.Page = p
		}
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil {
			filter.PageSize = ps
		}
	}

	tasks, total, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tasks"})
		return
	}

	var response []TaskResponse
	for _, t := range tasks {
		task := t
		response = append(response, toResponse(&task))
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": response,
		"total": total,
	})
}

// UpdateTask godoc
// @Summary      Update a task
// @Description  Update a task by ID
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id path int true "Task ID"
// @Param        request body UpdateTaskRequest true "Task update"
// @Success      200  {object}  TaskResponse
// @Failure      404  {object}  map[string]string
// @Router       /tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domains.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get task"})
		return
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Assignee != nil {
		task.Assignee = *req.Assignee
	}
	if req.Status != nil {
		task.Status = constants.TaskStatus(*req.Status)
	}

	updated, err := h.svc.Update(c.Request.Context(), task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update task"})
		return
	}

	c.JSON(http.StatusOK, toResponse(updated))
}

// DeleteTask godoc
// @Summary      Delete a task
// @Description  Delete a task by ID
// @Tags         tasks
// @Param        id path int true "Task ID"
// @Success      204
// @Failure      404  {object}  map[string]string
// @Router       /tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, domains.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete task"})
		return
	}

	c.Status(http.StatusNoContent)
}
