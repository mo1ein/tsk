package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/graph/task-manager/internal/constants"
	"github.com/graph/task-manager/internal/domains"
)

type mockTaskService struct {
	tasks  map[int64]*domains.Task
	nextID int64
}

func newMockTaskService() *mockTaskService {
	return &mockTaskService{
		tasks:  make(map[int64]*domains.Task),
		nextID: 1,
	}
}

func (m *mockTaskService) Create(ctx context.Context, task *domains.Task) (*domains.Task, error) {
	task.ID = m.nextID
	m.nextID++
	if task.Status == "" {
		task.Status = constants.TaskStatusPending
	}
	m.tasks[task.ID] = task
	return task, nil
}

func (m *mockTaskService) GetByID(ctx context.Context, id int64) (*domains.Task, error) {
	task, ok := m.tasks[id]
	if !ok {
		return nil, domains.ErrTaskNotFound
	}
	return task, nil
}

func (m *mockTaskService) List(ctx context.Context, filter domains.ListFilter) ([]domains.Task, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	var filtered []domains.Task
	for _, t := range m.tasks {
		if filter.Status != "" && string(t.Status) != filter.Status {
			continue
		}
		if filter.Assignee != "" && t.Assignee != filter.Assignee {
			continue
		}
		filtered = append(filtered, *t)
	}

	total := int64(len(filtered))
	start := (filter.Page - 1) * filter.PageSize
	if start >= int(total) {
		return []domains.Task{}, total, nil
	}
	end := start + filter.PageSize
	if end > int(total) {
		end = int(total)
	}

	return filtered[start:end], total, nil
}

func (m *mockTaskService) Update(ctx context.Context, task *domains.Task) (*domains.Task, error) {
	if _, ok := m.tasks[task.ID]; !ok {
		return nil, domains.ErrTaskNotFound
	}
	m.tasks[task.ID] = task
	return task, nil
}

func (m *mockTaskService) Delete(ctx context.Context, id int64) error {
	if _, ok := m.tasks[id]; !ok {
		return domains.ErrTaskNotFound
	}
	delete(m.tasks, id)
	return nil
}

func setupTestRouter() (*gin.Engine, *mockTaskService) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockTaskService()
	handler := NewTaskHandler(mockSvc)

	r := gin.New()
	r.Use(gin.Recovery())

	tasks := r.Group("/tasks")
	{
		tasks.POST("", handler.CreateTask)
		tasks.GET("", handler.ListTasks)
		tasks.GET("/:id", handler.GetTask)
		tasks.PUT("/:id", handler.UpdateTask)
		tasks.DELETE("/:id", handler.DeleteTask)
	}

	return r, mockSvc
}

func TestCreateTask(t *testing.T) {
	r, _ := setupTestRouter()

	body := `{"title":"Test Task","assignee":"alice"}`
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var resp TaskResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Title != "Test Task" {
		t.Errorf("expected title 'Test Task', got '%s'", resp.Title)
	}
}

func TestCreateTask_InvalidJSON(t *testing.T) {
	r, _ := setupTestRouter()

	body := `invalid json`
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCreateTask_MissingTitle(t *testing.T) {
	r, _ := setupTestRouter()

	body := `{"assignee":"alice"}`
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetTask(t *testing.T) {
	r, mockSvc := setupTestRouter()

	mockSvc.Create(context.Background(), &domains.Task{Title: "Test Task", Assignee: "alice"})

	req := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp TaskResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.ID != 1 {
		t.Errorf("expected ID 1, got %d", resp.ID)
	}
}

func TestGetTask_InvalidID(t *testing.T) {
	r, _ := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/tasks/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetTask_NotFound(t *testing.T) {
	r, _ := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/tasks/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestListTasks(t *testing.T) {
	r, mockSvc := setupTestRouter()

	mockSvc.Create(context.Background(), &domains.Task{Title: "Task 1", Assignee: "alice"})
	mockSvc.Create(context.Background(), &domains.Task{Title: "Task 2", Assignee: "bob"})

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if total := result["total"].(float64); total != 2 {
		t.Errorf("expected total 2, got %v", total)
	}
}

func TestListTasks_WithPagination(t *testing.T) {
	r, mockSvc := setupTestRouter()

	for i := 0; i < 25; i++ {
		mockSvc.Create(context.Background(), &domains.Task{Title: "Task", Assignee: "alice"})
	}

	req := httptest.NewRequest(http.MethodGet, "/tasks?page=2&page_size=10", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	tasks := result["tasks"].([]interface{})
	if len(tasks) != 10 {
		t.Errorf("expected 10 tasks on page 2, got %d", len(tasks))
	}
}

func TestListTasks_WithStatusFilter(t *testing.T) {
	r, mockSvc := setupTestRouter()

	mockSvc.Create(context.Background(), &domains.Task{Title: "Task 1", Assignee: "alice", Status: constants.TaskStatusPending})
	task2, _ := mockSvc.Create(context.Background(), &domains.Task{Title: "Task 2", Assignee: "bob", Status: constants.TaskStatusPending})
	mockSvc.Update(context.Background(), &domains.Task{ID: task2.ID, Title: "Task 2", Assignee: "bob", Status: constants.TaskStatusDone})

	req := httptest.NewRequest(http.MethodGet, "/tasks?status=done", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if total := result["total"].(float64); total != 1 {
		t.Errorf("expected total 1 for done filter, got %v", total)
	}
}

func TestListTasks_WithAssigneeFilter(t *testing.T) {
	r, mockSvc := setupTestRouter()

	mockSvc.Create(context.Background(), &domains.Task{Title: "Task 1", Assignee: "alice"})
	mockSvc.Create(context.Background(), &domains.Task{Title: "Task 2", Assignee: "bob"})

	req := httptest.NewRequest(http.MethodGet, "/tasks?assignee=alice", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if total := result["total"].(float64); total != 1 {
		t.Errorf("expected total 1 for alice filter, got %v", total)
	}
}

func TestUpdateTask(t *testing.T) {
	r, mockSvc := setupTestRouter()

	mockSvc.Create(context.Background(), &domains.Task{Title: "Original", Assignee: "alice"})

	body := `{"title":"Updated Task"}`
	req := httptest.NewRequest(http.MethodPut, "/tasks/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp TaskResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Title != "Updated Task" {
		t.Errorf("expected title 'Updated Task', got '%s'", resp.Title)
	}
}

func TestUpdateTask_InvalidID(t *testing.T) {
	r, _ := setupTestRouter()

	body := `{"title":"Updated Task"}`
	req := httptest.NewRequest(http.MethodPut, "/tasks/abc", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestUpdateTask_NotFound(t *testing.T) {
	r, _ := setupTestRouter()

	body := `{"title":"Updated Task"}`
	req := httptest.NewRequest(http.MethodPut, "/tasks/999", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestDeleteTask(t *testing.T) {
	r, mockSvc := setupTestRouter()

	mockSvc.Create(context.Background(), &domains.Task{Title: "To Delete", Assignee: "alice"})

	req := httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
}

func TestDeleteTask_InvalidID(t *testing.T) {
	r, _ := setupTestRouter()

	req := httptest.NewRequest(http.MethodDelete, "/tasks/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestDeleteTask_NotFound(t *testing.T) {
	r, _ := setupTestRouter()

	req := httptest.NewRequest(http.MethodDelete, "/tasks/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}
