// Package constants defines application-wide constants and enums.
package constants

import "fmt"

// TaskStatus represents the status of a task.
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"     // Task is created but not started.
	TaskStatusInProgress TaskStatus = "in_progress" // Task is being worked on.
	TaskStatusDone       TaskStatus = "done"        // Task is completed.
)

// String returns the string representation of the status.
func (s TaskStatus) String() string {
	return string(s)
}

// Scan implements the sql.Scanner interface for database deserialization.
func (s *TaskStatus) Scan(value interface{}) error {
	if value == nil {
		*s = TaskStatusPending
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("cannot scan %T into TaskStatus", value)
	}
	*s = TaskStatus(str)
	return nil
}

// Value implements the driver.Valuer interface for database serialization.
func (s TaskStatus) Value() (string, error) {
	return string(s), nil
}
