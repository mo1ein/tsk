package domains

import (
	"time"

	"github.com/graph/task-manager/internal/constants"
)

type Task struct {
	ID        int64
	Title     string
	Assignee  string
	Status    constants.TaskStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}
