package constants

import (
	"testing"
)

func TestTaskStatus_Values(t *testing.T) {
	if TaskStatusPending != "pending" {
		t.Errorf("expected 'pending', got '%s'", TaskStatusPending)
	}
	if TaskStatusDone != "done" {
		t.Errorf("expected 'done', got '%s'", TaskStatusDone)
	}
}

func TestTaskStatus_Scan(t *testing.T) {
	var s TaskStatus

	if err := s.Scan("done"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != TaskStatusDone {
		t.Errorf("expected done, got %s", s)
	}

	if err := s.Scan(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != TaskStatusPending {
		t.Errorf("expected pending for nil, got %s", s)
	}

	if err := s.Scan(123); err == nil {
		t.Error("expected error for non-string scan")
	}
}

func TestTaskStatus_Value(t *testing.T) {
	s := TaskStatusDone
	val, err := s.Value()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "done" {
		t.Errorf("expected 'done', got '%s'", val)
	}
}
