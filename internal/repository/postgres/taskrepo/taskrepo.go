package taskrepo

import (
	"context"

	"gorm.io/gorm"

	"github.com/mo1ein/tsk/internal/constants"
	"github.com/mo1ein/tsk/internal/domains"
	"github.com/mo1ein/tsk/internal/models"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, task *domains.Task) (*domains.Task, error) {
	if task.Status == "" {
		task.Status = constants.TaskStatusPending
	}

	m := models.TaskFromDomain(*task)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return nil, err
	}

	result := m.ToDomain()
	return &result, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*domains.Task, error) {
	var m models.Task
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domains.ErrTaskNotFound
		}
		return nil, err
	}

	result := m.ToDomain()
	return &result, nil
}

func (r *Repository) List(ctx context.Context, filter domains.ListFilter) ([]domains.Task, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	query := r.db.WithContext(ctx).Model(&models.Task{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Assignee != "" {
		query = query.Where("assignee = ?", filter.Assignee)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var taskModels []models.Task
	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Order("id DESC").Offset(offset).Limit(filter.PageSize).Find(&taskModels).Error; err != nil {
		return nil, 0, err
	}

	var tasks []domains.Task
	for _, m := range taskModels {
		tasks = append(tasks, m.ToDomain())
	}

	return tasks, total, nil
}

func (r *Repository) Update(ctx context.Context, task *domains.Task) (*domains.Task, error) {
	m := models.TaskFromDomain(*task)
	result := r.db.WithContext(ctx).Save(&m)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, domains.ErrTaskNotFound
	}

	updated := m.ToDomain()
	return &updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&models.Task{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domains.ErrTaskNotFound
	}
	return nil
}
