package models

import (
	"testing"
	"time"

	"github.com/mo1ein/tsk/internal/constants"
	"github.com/mo1ein/tsk/internal/domains"
)

func TestToDomain(t *testing.T) {
	now := time.Now()
	m := Task{
		ID:        42,
		Title:     "Test Task",
		Assignee:  "alice",
		Status:    constants.TaskStatusDone,
		CreatedAt: now,
		UpdatedAt: now,
	}

	d := m.ToDomain()

	if d.ID != 42 {
		t.Errorf("expected ID 42, got %d", d.ID)
	}
	if d.Title != "Test Task" {
		t.Errorf("expected title 'Test Task', got '%s'", d.Title)
	}
	if d.Assignee != "alice" {
		t.Errorf("expected assignee 'alice', got '%s'", d.Assignee)
	}
	if d.Status != constants.TaskStatusDone {
		t.Errorf("expected status done, got %s", d.Status)
	}
	if !d.CreatedAt.Equal(now) {
		t.Errorf("expected CreatedAt %v, got %v", now, d.CreatedAt)
	}
}

func TestTaskFromDomain(t *testing.T) {
	now := time.Now()
	d := domains.Task{
		ID:        7,
		Title:     "Domain Task",
		Assignee:  "bob",
		Status:    constants.TaskStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	m := TaskFromDomain(d)

	if m.ID != 7 {
		t.Errorf("expected ID 7, got %d", m.ID)
	}
	if m.Title != "Domain Task" {
		t.Errorf("expected title 'Domain Task', got '%s'", m.Title)
	}
	if m.Assignee != "bob" {
		t.Errorf("expected assignee 'bob', got '%s'", m.Assignee)
	}
	if m.Status != constants.TaskStatusPending {
		t.Errorf("expected status pending, got %s", m.Status)
	}
}

func TestRoundTrip(t *testing.T) {
	original := domains.Task{
		ID:        100,
		Title:     "Round Trip",
		Assignee:  "charlie",
		Status:    constants.TaskStatusInProgress,
		CreatedAt: time.Now().Truncate(time.Second),
		UpdatedAt: time.Now().Truncate(time.Second),
	}

	m := TaskFromDomain(original)
	result := m.ToDomain()

	if result.ID != original.ID {
		t.Errorf("ID mismatch: %d vs %d", result.ID, original.ID)
	}
	if result.Title != original.Title {
		t.Errorf("Title mismatch: %s vs %s", result.Title, original.Title)
	}
	if result.Assignee != original.Assignee {
		t.Errorf("Assignee mismatch: %s vs %s", result.Assignee, original.Assignee)
	}
	if result.Status != original.Status {
		t.Errorf("Status mismatch: %s vs %s", result.Status, original.Status)
	}
}
