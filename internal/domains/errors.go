package domains

import "errors"

// Domain errors returned by repository and service layers.
var (
	// ErrTaskNotFound is returned when a requested task does not exist.
	ErrTaskNotFound = errors.New("task not found")
)
