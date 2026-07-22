// Package taskcache provides Redis-based caching for tasks.
package taskcache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/mo1ein/tsk/internal/domains"
)

const taskKey = "task:%d"
const listPrefix = "tasks:"

// Repository implements task caching using Redis.
type Repository struct {
	rdb *redis.Client
	ttl time.Duration
}

// New creates a new Redis task cache with the given TTL.
func New(rdb *redis.Client, ttl time.Duration) Repository {
	return Repository{
		rdb: rdb,
		ttl: ttl,
	}
}

// Get retrieves a cached task by ID. Returns ErrTaskNotFound on cache miss.
func (r Repository) Get(ctx context.Context, id int64) (*domains.Task, error) {
	key := fmt.Sprintf(taskKey, id)

	data, err := r.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, domains.ErrTaskNotFound
		}
		return nil, err
	}

	var task domains.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

// Set stores a task in the cache with the configured TTL.
func (r Repository) Set(ctx context.Context, task *domains.Task) error {
	key := fmt.Sprintf(taskKey, task.ID)
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, key, data, r.ttl).Err()
}

// Delete removes a task from the cache.
func (r Repository) Delete(ctx context.Context, id int64) error {
	key := fmt.Sprintf(taskKey, id)
	return r.rdb.Del(ctx, key).Err()
}

// GetList retrieves a cached list result by filter key. Returns ErrTaskNotFound on cache miss.
func (r Repository) GetList(ctx context.Context, filterKey string) ([]domains.Task, int64, error) {
	key := listPrefix + filterKey

	data, err := r.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, 0, domains.ErrTaskNotFound
		}
		return nil, 0, err
	}

	var result struct {
		Tasks []domains.Task `json:"tasks"`
		Total int64          `json:"total"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, 0, err
	}

	return result.Tasks, result.Total, nil
}

// SetList stores a list query result in the cache.
func (r Repository) SetList(ctx context.Context, filterKey string, tasks []domains.Task, total int64) error {
	key := listPrefix + filterKey

	result := struct {
		Tasks []domains.Task `json:"tasks"`
		Total int64          `json:"total"`
	}{
		Tasks: tasks,
		Total: total,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, key, data, r.ttl).Err()
}

// InvalidateList removes all cached list results.
func (r Repository) InvalidateList(ctx context.Context) error {
	iter := r.rdb.Scan(ctx, 0, listPrefix+"*", 100).Iterator()
	for iter.Next(ctx) {
		if err := r.rdb.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}
