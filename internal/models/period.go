package models

import (
	"time"
)

// Типы повторений
const (
	RecurrenceDaily    = "daily"
	RecurrenceMonthly  = "monthly"
	RecurrenceSpecific = "specific"
	RecurrenceEvenOdd  = "even_odd"
)

type CreateTaskPeriodRequest struct {
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	DueDate            CustomDate `json:"due_date"`
	Status             string     `json:"status,omitempty"`
	Tags               []string   `json:"tags,omitempty"`
	IsRecurring        bool       `json:"is_recurring"`
	RecurrenceType     string     `json:"recurrence_type,omitempty"`
	RecurrenceInterval int        `json:"recurrence_interval,omitempty"`
	RecurrenceDays     []int      `json:"recurrence_days,omitempty"`
}

type TaskOverride struct {
	TaskID       int        `json:"task_id"`
	OverrideDate time.Time  `json:"override_date"`
	Status       TaskStatus `json:"status"`
	Notes        string     `json:"notes,omitempty"`
}
