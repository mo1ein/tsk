package domains

// ListFilter defines filtering and pagination parameters for listing tasks.
type ListFilter struct {
	Status   string // Filter by task status.
	Assignee string // Filter by assignee.
	Page     int    // Page number (1-indexed).
	PageSize int    // Number of items per page.
}
