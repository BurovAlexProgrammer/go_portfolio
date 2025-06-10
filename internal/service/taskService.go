package service

import (
	"GoPortfolio/internal/domain"
	"context"
)

type TaskService struct {
	taskRepo domain.TaskRepository
}

func NewTaskService(t domain.TaskRepository) *TaskService {
	return &TaskService{
		taskRepo: t,
	}
}

func (t *TaskService) Create(ctx context.Context, taskName string, userId int64) (*domain.Task, error) {
	task := &domain.Task{
		UserId: userId,
		IsDone: false,
		Name:   taskName,
	}
	err := t.taskRepo.Create(ctx, task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (t *TaskService) DoneByName(ctx context.Context, taskName string, userId int64) error {
	return t.taskRepo.DoneByName(ctx, taskName, userId)
}

func (t *TaskService) CleanTasksByUserId(ctx context.Context, userId int64) error {
	return t.taskRepo.CleanTasksByUserId(ctx, userId)
}
