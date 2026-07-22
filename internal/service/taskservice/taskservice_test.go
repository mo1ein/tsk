package taskservice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/graph/task-manager/internal/constants"
	"github.com/graph/task-manager/internal/domains"
)

type mockTaskRepo struct {
	tasks  map[int64]*domains.Task
	nextID int64
	err    error
}

func newMockTaskRepo() *mockTaskRepo {
	return &mockTaskRepo{tasks: make(map[int64]*domains.Task), nextID: 1}
}

func (m *mockTaskRepo) Create(_ context.Context, task *domains.Task) (*domains.Task, error) {
	if m.err != nil {
		return nil, m.err
	}
	task.ID = m.nextID
	m.nextID++
	if task.Status == "" {
		task.Status = constants.TaskStatusPending
	}
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now
	m.tasks[task.ID] = task
	return task, nil
}

func (m *mockTaskRepo) GetByID(_ context.Context, id int64) (*domains.Task, error) {
	if m.err != nil {
		return nil, m.err
	}
	t, ok := m.tasks[id]
	if !ok {
		return nil, domains.ErrTaskNotFound
	}
	return t, nil
}

func (m *mockTaskRepo) List(_ context.Context, filter domains.ListFilter) ([]domains.Task, int64, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
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

func (m *mockTaskRepo) Update(_ context.Context, task *domains.Task) (*domains.Task, error) {
	if m.err != nil {
		return nil, m.err
	}
	if _, ok := m.tasks[task.ID]; !ok {
		return nil, domains.ErrTaskNotFound
	}
	task.UpdatedAt = time.Now()
	m.tasks[task.ID] = task
	return task, nil
}

func (m *mockTaskRepo) Delete(_ context.Context, id int64) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.tasks[id]; !ok {
		return domains.ErrTaskNotFound
	}
	delete(m.tasks, id)
	return nil
}

type mockCache struct {
	tasks    map[int64]*domains.Task
	lists    map[string][]domains.Task
totals    map[string]int64
	failGet  bool
	failSet  bool
	failList bool
}

func newMockCache() *mockCache {
	return &mockCache{
		tasks: make(map[int64]*domains.Task),
		lists: make(map[string][]domains.Task),
		totals: make(map[string]int64),
	}
}

func (c *mockCache) Get(_ context.Context, id int64) (*domains.Task, error) {
	if c.failGet {
		return nil, errors.New("cache miss")
	}
	t, ok := c.tasks[id]
	if !ok {
		return nil, domains.ErrTaskNotFound
	}
	return t, nil
}

func (c *mockCache) Set(_ context.Context, task *domains.Task) error {
	if c.failSet {
		return errors.New("cache set failed")
	}
	c.tasks[task.ID] = task
	return nil
}

func (c *mockCache) Delete(_ context.Context, id int64) error {
	delete(c.tasks, id)
	return nil
}

func (c *mockCache) GetList(_ context.Context, filterKey string) ([]domains.Task, int64, error) {
	if c.failList {
		return nil, 0, errors.New("cache list miss")
	}
	tasks, ok := c.lists[filterKey]
	if !ok {
		return nil, 0, domains.ErrTaskNotFound
	}
	return tasks, c.totals[filterKey], nil
}

func (c *mockCache) SetList(_ context.Context, filterKey string, tasks []domains.Task, total int64) error {
	c.lists[filterKey] = tasks
	c.totals[filterKey] = total
	return nil
}

func (c *mockCache) InvalidateList(_ context.Context) error {
	c.lists = make(map[string][]domains.Task)
	c.totals = make(map[string]int64)
	return nil
}

func TestCreate(t *testing.T) {
	repo := newMockTaskRepo()
	cache := newMockCache()
	svc := New(repo, cache)

	task, err := svc.Create(context.Background(), &domains.Task{Title: "Test", Assignee: "alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID != 1 {
		t.Errorf("expected ID 1, got %d", task.ID)
	}
	if task.Status != constants.TaskStatusPending {
		t.Errorf("expected status pending, got %s", task.Status)
	}
}

func TestCreate_RepoError(t *testing.T) {
	repo := newMockTaskRepo()
	repo.err = errors.New("db error")
	cache := newMockCache()
	svc := New(repo, cache)

	_, err := svc.Create(context.Background(), &domains.Task{Title: "Test"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetByID_CacheHit(t *testing.T) {
	repo := newMockTaskRepo()
	cache := newMockCache()
	svc := New(repo, cache)

	cachedTask := &domains.Task{ID: 42, Title: "Cached"}
	cache.tasks[42] = cachedTask

	task, err := svc.GetByID(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title != "Cached" {
		t.Errorf("expected cached task, got %v", task)
	}
}

func TestGetByID_CacheMiss(t *testing.T) {
	repo := newMockTaskRepo()
	cache := newMockCache()
	svc := New(repo, cache)

	repo.tasks[1] = &domains.Task{ID: 1, Title: "DB Task"}

	task, err := svc.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title != "DB Task" {
		t.Errorf("expected 'DB Task', got '%s'", task.Title)
	}
	if _, ok := cache.tasks[1]; !ok {
		t.Error("expected task to be cached after DB fetch")
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo := newMockTaskRepo()
	cache := newMockCache()
	svc := New(repo, cache)

	_, err := svc.GetByID(context.Background(), 999)
	if !errors.Is(err, domains.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestList_CacheHit(t *testing.T) {
	repo := newMockTaskRepo()
	cache := newMockCache()
	svc := New(repo, cache)

	filterKey := "status=,assignee=,page=1,size=20"
	cache.lists[filterKey] = []domains.Task{{ID: 1, Title: "Cached"}}
	cache.totals[filterKey] = 1

	tasks, total, err := svc.List(context.Background(), domains.ListFilter{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}
}

func TestList_CacheMiss(t *testing.T) {
	repo := newMockTaskRepo()
	cache := newMockCache()
	svc := New(repo, cache)

	repo.tasks[1] = &domains.Task{ID: 1, Title: "Task 1"}

	tasks, total, err := svc.List(context.Background(), domains.ListFilter{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}
}

func TestList_RepoError(t *testing.T) {
	repo := newMockTaskRepo()
	repo.err = errors.New("db error")
	cache := newMockCache()
	svc := New(repo, cache)

	_, _, err := svc.List(context.Background(), domains.ListFilter{Page: 1, PageSize: 20})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdate_Success(t *testing.T) {
	repo := newMockTaskRepo()
	cache := newMockCache()
	svc := New(repo, cache)

	repo.tasks[1] = &domains.Task{ID: 1, Title: "Old"}

	updated, err := svc.Update(context.Background(), &domains.Task{ID: 1, Title: "New"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Title != "New" {
		t.Errorf("expected title 'New', got '%s'", updated.Title)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	repo := newMockTaskRepo()
	cache := newMockCache()
	svc := New(repo, cache)

	_, err := svc.Update(context.Background(), &domains.Task{ID: 999, Title: "New"})
	if !errors.Is(err, domains.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestUpdate_InvalidatesCache(t *testing.T) {
	repo := newMockTaskRepo()
	cache := newMockCache()
	svc := New(repo, cache)

	repo.tasks[1] = &domains.Task{ID: 1, Title: "Old"}
	cache.tasks[1] = &domains.Task{ID: 1, Title: "Old"}

	_, err := svc.Update(context.Background(), &domains.Task{ID: 1, Title: "New"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cache.tasks[1]; ok {
		t.Error("expected task to be removed from cache after update")
	}
}

func TestDelete_Success(t *testing.T) {
	repo := newMockTaskRepo()
	cache := newMockCache()
	svc := New(repo, cache)

	repo.tasks[1] = &domains.Task{ID: 1, Title: "To Delete"}

	err := svc.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := repo.tasks[1]; ok {
		t.Error("expected task to be deleted from repo")
	}
}

func TestDelete_NotFound(t *testing.T) {
	repo := newMockTaskRepo()
	cache := newMockCache()
	svc := New(repo, cache)

	err := svc.Delete(context.Background(), 999)
	if !errors.Is(err, domains.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestDelete_InvalidatesCache(t *testing.T) {
	repo := newMockTaskRepo()
	cache := newMockCache()
	svc := New(repo, cache)

	repo.tasks[1] = &domains.Task{ID: 1, Title: "To Delete"}
	cache.tasks[1] = &domains.Task{ID: 1, Title: "To Delete"}

	err := svc.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cache.tasks[1]; ok {
		t.Error("expected task to be removed from cache after delete")
	}
}
