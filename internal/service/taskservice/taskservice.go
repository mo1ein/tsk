package taskservice

import (
	"context"
	"fmt"

	"github.com/graph/task-manager/internal/domains"
	"github.com/graph/task-manager/internal/repository/redis/taskcache"
)

type taskRepository interface {
	Create(ctx context.Context, task *domains.Task) (*domains.Task, error)
	GetByID(ctx context.Context, id int64) (*domains.Task, error)
	List(ctx context.Context, filter domains.ListFilter) ([]domains.Task, int64, error)
	Update(ctx context.Context, task *domains.Task) (*domains.Task, error)
	Delete(ctx context.Context, id int64) error
}

type Service struct {
	taskRepo taskRepository
	cache    taskcache.Repository
}

func New(taskRepo taskRepository, cache taskcache.Repository) *Service {
	return &Service{
		taskRepo: taskRepo,
		cache:    cache,
	}
}

func (s *Service) Create(ctx context.Context, task *domains.Task) (*domains.Task, error) {
	created, err := s.taskRepo.Create(ctx, task)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, created)
	_ = s.cache.InvalidateList(ctx)

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*domains.Task, error) {
	cached, err := s.cache.Get(ctx, id)
	if err == nil {
		return cached, nil
	}

	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, task)

	return task, nil
}

func (s *Service) List(ctx context.Context, filter domains.ListFilter) ([]domains.Task, int64, error) {
	filterKey := fmt.Sprintf("status=%s,assignee=%s,page=%d,size=%d", filter.Status, filter.Assignee, filter.Page, filter.PageSize)

	cached, total, err := s.cache.GetList(ctx, filterKey)
	if err == nil {
		return cached, total, nil
	}

	tasks, total, err := s.taskRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	_ = s.cache.SetList(ctx, filterKey, tasks, total)

	return tasks, total, nil
}

func (s *Service) Update(ctx context.Context, task *domains.Task) (*domains.Task, error) {
	updated, err := s.taskRepo.Update(ctx, task)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Delete(ctx, updated.ID)
	_ = s.cache.InvalidateList(ctx)

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	err := s.taskRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	_ = s.cache.Delete(ctx, id)
	_ = s.cache.InvalidateList(ctx)

	return nil
}
