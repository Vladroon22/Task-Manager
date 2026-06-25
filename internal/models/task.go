package models

import (
	"errors"
	"time"
)

type TaskStatus string

const (
	StatusNew        TaskStatus = "new"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
	StatusCancelled  TaskStatus = "cancelled"
)

var (
	ErrTaskNotFound = errors.New("Task not found")
	ErrTagNotFound  = errors.New("Tag not found")
	ErrTagIsSystem  = errors.New("Tag is system")
	ErrTagNotOnTask = errors.New("Tag wasn't found on this task")
)

type Task struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	DueDate     CustomDate `json:"due_date"`
	Status      TaskStatus `json:"status"`
	Tags        []Tag      `json:"tags,omitempty"`

	IsRecurring        bool    `json:"is_recurring"`
	RecurrenceType     *string `json:"recurrence_type,omitempty"`
	RecurrenceInterval *int    `json:"recurrence_interval,omitempty"`
	RecurrenceDays     []int   `json:"recurrence_days,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateTaskRequest struct {
	Title       string     `json:"title" validate:"required"`
	Description string     `json:"description,omitempty"`
	DueDate     time.Time  `json:"due_date" validate:"required"`
	Status      TaskStatus `json:"status,omitempty"`
	Tags        []string   `json:"tags,omitempty"`

	IsRecurring        bool   `json:"is_recurring"`
	RecurrenceType     string `json:"recurrence_type,omitempty"`
	RecurrenceInterval int    `json:"recurrence_interval,omitempty"`
	RecurrenceDays     *[]int `json:"recurrence_days,omitempty"`
}

type UpdateTaskRequest struct {
	Title       *string     `json:"title,omitempty"`
	Description *string     `json:"description,omitempty"`
	DueDate     *time.Time  `json:"due_date,omitempty"`
	Status      *TaskStatus `json:"status,omitempty"`
}

type TaskFilter struct {
	Status   *TaskStatus `json:"status,omitempty"`
	DateFrom *time.Time  `json:"date_from,omitempty"`
	DateTo   *time.Time  `json:"date_to,omitempty"`
	Tags     []string    `json:"tags,omitempty"`
	Limit    int         `json:"limit,omitempty"`
	Offset   int         `json:"offset,omitempty"`
}

func (s TaskStatus) IsValid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone, StatusCancelled:
		return true
	}
	return false
}
