package domains

import (
	"errors"
	"testing"
)

func TestErrTaskNotFound(t *testing.T) {
	if ErrTaskNotFound == nil {
		t.Fatal("ErrTaskNotFound should not be nil")
	}
	if ErrTaskNotFound.Error() != "task not found" {
		t.Errorf("expected 'task not found', got '%s'", ErrTaskNotFound.Error())
	}
}

func TestErrTaskNotFound_Is(t *testing.T) {
	err := ErrTaskNotFound
	if !errors.Is(err, ErrTaskNotFound) {
		t.Error("expected errors.Is to return true")
	}
}

func TestTask_Fields(t *testing.T) {
	task := Task{
		ID:       1,
		Title:    "Test",
		Assignee: "alice",
	}

	if task.ID != 1 {
		t.Errorf("expected ID 1, got %d", task.ID)
	}
	if task.Title != "Test" {
		t.Errorf("expected title 'Test', got '%s'", task.Title)
	}
	if task.Assignee != "alice" {
		t.Errorf("expected assignee 'alice', got '%s'", task.Assignee)
	}
}

func TestListFilter_Fields(t *testing.T) {
	f := ListFilter{
		Status:   "done",
		Assignee: "bob",
		Page:     2,
		PageSize: 10,
	}

	if f.Status != "done" {
		t.Errorf("expected status 'done', got '%s'", f.Status)
	}
	if f.Assignee != "bob" {
		t.Errorf("expected assignee 'bob', got '%s'", f.Assignee)
	}
	if f.Page != 2 {
		t.Errorf("expected page 2, got %d", f.Page)
	}
	if f.PageSize != 10 {
		t.Errorf("expected page_size 10, got %d", f.PageSize)
	}
}
