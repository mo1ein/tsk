// Package domains defines the core domain types and business errors
// used across the application layers.
package domains

import (
	"time"

	"github.com/mo1ein/tsk/internal/constants"
)

// Task represents a to-do task in the system.
type Task struct {
	ID        int64                // Unique identifier.
	Title     string               // Task title.
	Assignee  string               // Person assigned to the task.
	Status    constants.TaskStatus // Current status (pending, in_progress, done).
	CreatedAt time.Time            // Creation timestamp.
	UpdatedAt time.Time            // Last update timestamp.
}
