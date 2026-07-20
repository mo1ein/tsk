package mock

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph/task-manager/internal/constants"
	"github.com/graph/task-manager/internal/domains"
)

type TaskRepository struct {
	tasks  map[int64]*domains.Task
	nextID int64
	mu     sync.RWMutex
}

func New() *TaskRepository {
	return &TaskRepository{
		tasks:  make(map[int64]*domains.Task),
		nextID: 1,
	}
}

func (m *TaskRepository) Create(ctx context.Context, task *domains.Task) (*domains.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task.ID = m.nextID
	m.nextID++
	if task.Status == "" {
		task.Status = constants.TaskStatusPending
	}
	m.tasks[task.ID] = task
	return task, nil
}

func (m *TaskRepository) GetByID(ctx context.Context, id int64) (*domains.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, ok := m.tasks[id]
	if !ok {
		return nil, domains.ErrTaskNotFound
	}
	return task, nil
}

func (m *TaskRepository) List(ctx context.Context, filter domains.ListFilter) ([]domains.Task, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

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

func (m *TaskRepository) Update(ctx context.Context, task *domains.Task) (*domains.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.tasks[task.ID]; !ok {
		return nil, domains.ErrTaskNotFound
	}
	m.tasks[task.ID] = task
	return task, nil
}

func (m *TaskRepository) Delete(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.tasks[id]; !ok {
		return fmt.Errorf("task not found: %w", domains.ErrTaskNotFound)
	}
	delete(m.tasks, id)
	return nil
}
