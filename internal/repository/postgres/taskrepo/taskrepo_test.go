package taskrepo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/mo1ein/tsk/internal/constants"
	"github.com/mo1ein/tsk/internal/domains"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}
	pass := os.Getenv("DB_PASSWORD")
	if pass == "" {
		pass = "postgres"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "taskdb"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, dbname,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("skipping integration test: cannot connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Skipf("skipping integration test: database not reachable: %v", err)
	}

	db.Exec("DELETE FROM tasks")

	return db
}

func cleanupDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	db.Exec("DELETE FROM tasks")
}

func TestCreate(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupDB(t, db)

	repo := New(db)

	task, err := repo.Create(context.Background(), &domains.Task{
		Title:    "Test Task",
		Assignee: "alice",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if task.Title != "Test Task" {
		t.Errorf("expected title 'Test Task', got '%s'", task.Title)
	}
	if task.Status != constants.TaskStatusPending {
		t.Errorf("expected status pending, got %s", task.Status)
	}
	if task.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestCreate_DefaultStatus(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupDB(t, db)

	repo := New(db)

	task, err := repo.Create(context.Background(), &domains.Task{
		Title: "No Status",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Status != constants.TaskStatusPending {
		t.Errorf("expected default status pending, got %s", task.Status)
	}
}

func TestGetByID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupDB(t, db)

	repo := New(db)

	created, _ := repo.Create(context.Background(), &domains.Task{
		Title: "Find Me",
		Assignee: "bob",
	})

	found, err := repo.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, found.ID)
	}
	if found.Title != "Find Me" {
		t.Errorf("expected title 'Find Me', got '%s'", found.Title)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupDB(t, db)

	repo := New(db)

	_, err := repo.GetByID(context.Background(), 999999)
	if !errors.Is(err, domains.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestList(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupDB(t, db)

	repo := New(db)

	for i := 0; i < 5; i++ {
		repo.Create(context.Background(), &domains.Task{
			Title:    fmt.Sprintf("Task %d", i),
			Assignee: "alice",
		})
	}

	tasks, total, err := repo.List(context.Background(), domains.ListFilter{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(tasks) != 5 {
		t.Errorf("expected 5 tasks, got %d", len(tasks))
	}
}

func TestList_Pagination(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupDB(t, db)

	repo := New(db)

	for i := 0; i < 15; i++ {
		repo.Create(context.Background(), &domains.Task{
			Title:    fmt.Sprintf("Task %d", i),
			Assignee: "alice",
		})
	}

	tasks, total, err := repo.List(context.Background(), domains.ListFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 15 {
		t.Errorf("expected total 15, got %d", total)
	}
	if len(tasks) != 10 {
		t.Errorf("expected 10 tasks on page 1, got %d", len(tasks))
	}

	tasks2, _, err := repo.List(context.Background(), domains.ListFilter{Page: 2, PageSize: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks2) != 5 {
		t.Errorf("expected 5 tasks on page 2, got %d", len(tasks2))
	}
}

func TestList_FilterByStatus(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupDB(t, db)

	repo := New(db)

	repo.Create(context.Background(), &domains.Task{Title: "Pending", Status: constants.TaskStatusPending})
	done := &domains.Task{Title: "Done", Status: constants.TaskStatusDone}
	repo.Create(context.Background(), done)

	tasks, total, err := repo.List(context.Background(), domains.ListFilter{Status: "done", Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1 for done filter, got %d", total)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Title != "Done" {
		t.Errorf("expected 'Done', got '%s'", tasks[0].Title)
	}
}

func TestList_FilterByAssignee(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupDB(t, db)

	repo := New(db)

	repo.Create(context.Background(), &domains.Task{Title: "Alice Task", Assignee: "alice"})
	repo.Create(context.Background(), &domains.Task{Title: "Bob Task", Assignee: "bob"})

	tasks, total, err := repo.List(context.Background(), domains.ListFilter{Assignee: "alice", Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1 for alice filter, got %d", total)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Assignee != "alice" {
		t.Errorf("expected assignee 'alice', got '%s'", tasks[0].Assignee)
	}
}

func TestList_DefaultPagination(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupDB(t, db)

	repo := New(db)

	tasks, total, err := repo.List(context.Background(), domains.ListFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("expected total 0, got %d", total)
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestUpdate(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupDB(t, db)

	repo := New(db)

	created, _ := repo.Create(context.Background(), &domains.Task{
		Title: "Original",
		Assignee: "alice",
	})

	created.Title = "Updated"
	created.Status = constants.TaskStatusDone
	updated, err := repo.Update(context.Background(), created)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Title != "Updated" {
		t.Errorf("expected title 'Updated', got '%s'", updated.Title)
	}
	if updated.Status != constants.TaskStatusDone {
		t.Errorf("expected status done, got %s", updated.Status)
	}
	if updated.UpdatedAt.Before(time.Now().Add(-time.Second)) {
		t.Error("expected UpdatedAt to be recent")
	}
}

func TestDelete(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupDB(t, db)

	repo := New(db)

	created, _ := repo.Create(context.Background(), &domains.Task{
		Title: "To Delete",
	})

	err := repo.Delete(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = repo.GetByID(context.Background(), created.ID)
	if !errors.Is(err, domains.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound after delete, got %v", err)
	}
}

func TestDelete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupDB(t, db)

	repo := New(db)

	err := repo.Delete(context.Background(), 999999)
	if !errors.Is(err, domains.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}
