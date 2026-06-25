package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Vladroon22/TaskTracker/internal/models"
)

type TaskServicer interface {
	Create(ctx context.Context, task *models.CreateTaskRequest) (*models.Task, error)
	GetByID(ctx context.Context, id int) (*models.Task, error)
	List(ctx context.Context, filter *models.TaskFilter) ([]models.Task, error)
	Update(ctx context.Context, id int, update *models.UpdateTaskRequest) (*models.Task, error)
	Delete(ctx context.Context, id int) error
	CreateTaskWithPeriod(ctx context.Context, req *models.CreateTaskPeriodRequest) (*models.Task, error)
	GetTasksForPeriod(ctx context.Context, from, to time.Time) ([]models.Task, error)
	SetTaskStatus(ctx context.Context, taskID int, date time.Time, status models.TaskStatus) error
	DeleteTaskOverride(ctx context.Context, taskID int, date time.Time) error
}

type Service struct {
	repo TaskServicer
}

func NewTaskService(repo TaskServicer) TaskServicer {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, task *models.CreateTaskRequest) (*models.Task, error) {
	if task.Title == "" {
		return nil, fmt.Errorf("%v", "task cannot be empty")
	}

	if task.DueDate == "" {
		return nil, fmt.Errorf("%v", "Due date is required")
	}

	if task.Status != "" && !task.Status.IsValid() {
		return nil, fmt.Errorf("%v", "Invalid status")
	}

	return s.repo.Create(ctx, task)
}

func (s *Service) GetByID(ctx context.Context, id int) (*models.Task, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, filter *models.TaskFilter) ([]models.Task, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) Update(ctx context.Context, id int, update *models.UpdateTaskRequest) (*models.Task, error) {
	if m, err := s.repo.Update(ctx, id, update); err == nil {
		return m, nil
	} else if err.Error() == "task not found" {
		return nil, fmt.Errorf("%v", "Task not found")
	} else if err.Error() == fmt.Sprintf("invalid status: %s", *update.Status) {
		return nil, fmt.Errorf("%v", err.Error())
	} else {
		return nil, fmt.Errorf("%v", "Failed to update task")
	}
}

func (s *Service) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err == nil {
		return nil
	} else if err.Error() == "task not found" {
		return fmt.Errorf("%v", "Task not found")
	} else {
		return fmt.Errorf("%v", "Failed to delete task")
	}
}

func (s *Service) CreateTaskWithPeriod(ctx context.Context, req *models.CreateTaskPeriodRequest) (*models.Task, error) {
	return s.repo.CreateTaskWithPeriod(ctx, req)
}
func (s *Service) GetTasksForPeriod(ctx context.Context, from, to time.Time) ([]models.Task, error) {
	return s.repo.GetTasksForPeriod(ctx, from, to)
}

func (s *Service) SetTaskStatus(ctx context.Context, taskID int, date time.Time, status models.TaskStatus) error {
	return s.repo.SetTaskStatus(ctx, taskID, date, status)
}

func (s *Service) DeleteTaskOverride(ctx context.Context, taskID int, date time.Time) error {
	return s.DeleteTaskOverride(ctx, taskID, date)
}
