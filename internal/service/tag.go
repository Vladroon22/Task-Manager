package service

import (
	"context"

	"github.com/Vladroon22/TaskTracker/internal/models"
)

type TagServicer interface {
	CreateTag(context.Context, []string) (*models.Tag, error)
	GetTagByID(context.Context, int) (*models.Tag, error)
	ListTags(context.Context) ([]models.Tag, error)
	DeleteTag(context.Context, int) error
	AddTagToTask(context.Context, int, int, []string) error
	RemoveTagFromTask(context.Context, int, int) error
	GetTaskTags(context.Context, int) ([]models.Tag, error)
}

type TagService struct {
	repo TagServicer
}

func NewTagService(repo TagServicer) TagServicer {
	return &TagService{repo: repo}
}

func (ts *TagService) CreateTag(ctx context.Context, names []string) (*models.Tag, error) {
	return ts.repo.CreateTag(ctx, names)
}

func (ts *TagService) GetTagByID(ctx context.Context, id int) (*models.Tag, error) {
	return ts.repo.GetTagByID(ctx, id)
}

func (ts *TagService) GetTaskTags(ctx context.Context, taskID int) ([]models.Tag, error) {
	return ts.repo.GetTaskTags(ctx, taskID)
}

func (ts *TagService) ListTags(ctx context.Context) ([]models.Tag, error) {
	return ts.repo.ListTags(ctx)
}

func (ts *TagService) DeleteTag(ctx context.Context, id int) error {
	return ts.repo.DeleteTag(ctx, id)
}

func (ts *TagService) AddTagToTask(ctx context.Context, taskID, tagID int, names []string) error {
	return ts.repo.AddTagToTask(ctx, taskID, tagID, names)
}

func (ts *TagService) RemoveTagFromTask(ctx context.Context, taskID, tagID int) error {
	return ts.repo.RemoveTagFromTask(ctx, taskID, tagID)
}
