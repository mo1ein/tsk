package taskcache

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mo1ein/tsk/internal/domains"
	"github.com/redis/go-redis/v9"
)

func setupRedis(t *testing.T) *redis.Client {
	t.Helper()
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("skipping Redis test: %v", err)
	}
	rdb.FlushAll(ctx)
	return rdb
}

func TestSetAndGet(t *testing.T) {
	rdb := setupRedis(t)
	defer rdb.Close()

	repo := New(rdb, 5*time.Minute)

	task := &domains.Task{
		ID:       1,
		Title:    "Cached Task",
		Assignee: "alice",
		Status:   "pending",
	}

	if err := repo.Set(context.Background(), task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Title != "Cached Task" {
		t.Errorf("expected 'Cached Task', got '%s'", got.Title)
	}
}

func TestGet_Miss(t *testing.T) {
	rdb := setupRedis(t)
	defer rdb.Close()

	repo := New(rdb, 5*time.Minute)

	_, err := repo.Get(context.Background(), 999)
	if err == nil {
		t.Error("expected error for cache miss")
	}
}

func TestDelete(t *testing.T) {
	rdb := setupRedis(t)
	defer rdb.Close()

	repo := New(rdb, 5*time.Minute)

	task := &domains.Task{ID: 1, Title: "To Delete"}
	repo.Set(context.Background(), task)

	if err := repo.Delete(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := repo.Get(context.Background(), 1)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestSetListAndGetList(t *testing.T) {
	rdb := setupRedis(t)
	defer rdb.Close()

	repo := New(rdb, 5*time.Minute)

	tasks := []domains.Task{
		{ID: 1, Title: "Task 1"},
		{ID: 2, Title: "Task 2"},
	}

	if err := repo.SetList(context.Background(), "filter1", tasks, 2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, total, err := repo.GetList(context.Background(), "filter1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(got))
	}
}

func TestGetList_Miss(t *testing.T) {
	rdb := setupRedis(t)
	defer rdb.Close()

	repo := New(rdb, 5*time.Minute)

	_, _, err := repo.GetList(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for list cache miss")
	}
}

func TestInvalidateList(t *testing.T) {
	rdb := setupRedis(t)
	defer rdb.Close()

	repo := New(rdb, 5*time.Minute)

	repo.SetList(context.Background(), "filter1", []domains.Task{{ID: 1}}, 1)
	repo.SetList(context.Background(), "filter2", []domains.Task{{ID: 2}}, 1)

	if err := repo.InvalidateList(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, _, err := repo.GetList(context.Background(), "filter1")
	if err == nil {
		t.Error("expected error after invalidation")
	}
}

func TestSet_InvalidJSON(t *testing.T) {
	rdb := setupRedis(t)
	defer rdb.Close()

	cache := Repository{rdb: rdb, ttl: 5 * time.Minute}
	data, _ := json.Marshal("invalid")

	key := "task:1"
	rdb.Set(context.Background(), key, data, 5*time.Minute)

	got, err := cache.Get(context.Background(), 1)
	if err == nil && got == nil {
		// Redis returns the data but Unmarshal fails
		t.Log("got result, checking unmarshal behavior")
	}
}
