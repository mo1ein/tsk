package mock

import (
	"context"
	"testing"

	"github.com/graph/task-manager/internal/constants"
	"github.com/graph/task-manager/internal/domains"
)

func TestCreate(t *testing.T) {
	repo := New()
	ctx := context.Background()

	task := &domains.Task{
		Title:    "Test Task",
		Assignee: "alice",
	}

	created, err := repo.Create(ctx, task)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created.ID != 1 {
		t.Errorf("expected ID 1, got %d", created.ID)
	}
	if created.Title != "Test Task" {
		t.Errorf("expected title 'Test Task', got '%s'", created.Title)
	}
	if created.Status != constants.TaskStatusPending {
		t.Errorf("expected status 'pending', got '%s'", created.Status)
	}
}

func TestGetByID(t *testing.T) {
	repo := New()
	ctx := context.Background()

	task := &domains.Task{Title: "Test Task", Assignee: "alice"}
	created, _ := repo.Create(ctx, task)

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Title != created.Title {
		t.Errorf("expected title '%s', got '%s'", created.Title, got.Title)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo := New()
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 999)
	if err != domains.ErrTaskNotFound {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestUpdate(t *testing.T) {
	repo := New()
	ctx := context.Background()

	task := &domains.Task{Title: "Original", Assignee: "alice"}
	created, _ := repo.Create(ctx, task)

	created.Title = "Updated"
	updated, err := repo.Update(ctx, created)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Title != "Updated" {
		t.Errorf("expected title 'Updated', got '%s'", updated.Title)
	}
}

func TestDelete(t *testing.T) {
	repo := New()
	ctx := context.Background()

	task := &domains.Task{Title: "To Delete", Assignee: "alice"}
	created, _ := repo.Create(ctx, task)

	err := repo.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = repo.GetByID(ctx, created.ID)
	if err != domains.ErrTaskNotFound {
		t.Errorf("expected ErrTaskNotFound after delete, got %v", err)
	}
}

func TestList(t *testing.T) {
	repo := New()
	ctx := context.Background()

	repo.Create(ctx, &domains.Task{Title: "Task 1", Assignee: "alice", Status: constants.TaskStatusPending})
	repo.Create(ctx, &domains.Task{Title: "Task 2", Assignee: "bob", Status: constants.TaskStatusDone})

	tasks, total, err := repo.List(ctx, domains.ListFilter{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestList_FilterByStatus(t *testing.T) {
	repo := New()
	ctx := context.Background()

	repo.Create(ctx, &domains.Task{Title: "Task 1", Assignee: "alice", Status: constants.TaskStatusPending})
	repo.Create(ctx, &domains.Task{Title: "Task 2", Assignee: "bob", Status: constants.TaskStatusDone})

	tasks, total, err := repo.List(ctx, domains.ListFilter{Status: "done"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if tasks[0].Title != "Task 2" {
		t.Errorf("expected 'Task 2', got '%s'", tasks[0].Title)
	}
}
