// Package models defines database models and domain-to-model conversion.
package models

import (
	"time"

	"github.com/mo1ein/tsk/internal/constants"
	"github.com/mo1ein/tsk/internal/domains"
	"gorm.io/gorm"
)

// Task is the GORM model for the tasks table.
type Task struct {
	ID        int64                `gorm:"column:id"`
	Title     string               `gorm:"column:title"`
	Assignee  string               `gorm:"column:assignee"`
	Status    constants.TaskStatus `gorm:"column:status"`
	CreatedAt time.Time            `gorm:"column:created_at"`
	UpdatedAt time.Time            `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt       `gorm:"column:deleted_at;index"`
}

// ToDomain converts the database model to a domain Task.
func (t Task) ToDomain() domains.Task {
	return domains.Task{
		ID:        t.ID,
		Title:     t.Title,
		Assignee:  t.Assignee,
		Status:    t.Status,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

// TaskFromDomain converts a domain Task to the database model.
func TaskFromDomain(d domains.Task) Task {
	return Task{
		ID:        d.ID,
		Title:     d.Title,
		Assignee:  d.Assignee,
		Status:    d.Status,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}
