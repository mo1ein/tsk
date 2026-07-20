package domains

import "context"

type TaskRepository interface {
	Create(ctx context.Context, task *Task) (*Task, error)
	GetByID(ctx context.Context, id int64) (*Task, error)
	List(ctx context.Context, filter ListFilter) ([]Task, int64, error)
	Update(ctx context.Context, task *Task) (*Task, error)
	Delete(ctx context.Context, id int64) error
}

type ListFilter struct {
	Status   string
	Assignee string
	Page     int
	PageSize int
}
