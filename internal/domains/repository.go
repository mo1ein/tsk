package domains

type ListFilter struct {
	Status   string
	Assignee string
	Page     int
	PageSize int
}
