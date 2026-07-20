package constants

import "fmt"

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

func (s TaskStatus) String() string {
	return string(s)
}

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

func (s TaskStatus) Value() (string, error) {
	return string(s), nil
}
