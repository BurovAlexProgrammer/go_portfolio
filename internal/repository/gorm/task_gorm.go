package gorm

import (
	"GoPortfolio/internal/domain"
	"context"
	"gorm.io/gorm"
)

type TaskGormRepo struct {
	db *gorm.DB
}

func NewTaskGormRepo(db *gorm.DB) *TaskGormRepo {
	return &TaskGormRepo{
		db: db,
	}
}

func (t TaskGormRepo) Create(ctx context.Context, task *domain.Task) error {
	return t.db.WithContext(ctx).Create(task).Error
}

func (t TaskGormRepo) CleanTasksByUserId(ctx context.Context, userId int64) error {
	return t.db.WithContext(ctx).Where(&domain.Task{UserId: userId}).Delete(&domain.Task{}).Error
}

func (t TaskGormRepo) DoneByName(ctx context.Context, taskName string, userId int64) error {
	task := &domain.Task{}
	result := t.db.WithContext(ctx).First(task, &domain.Task{UserId: userId, Name: taskName})

	if result.Error != nil {
		return result.Error
	}

	task.IsDone = true
	return t.db.WithContext(ctx).Updates(task).Error
}

func (t TaskGormRepo) ListByUser(ctx context.Context, userId int64) ([]domain.Task, error) {
	tasks := make([]domain.Task, 2)
	result := t.db.WithContext(ctx).Where(tasks, &domain.Task{UserId: userId})

	if result.Error != nil {
		return nil, result.Error
	}

	return tasks, nil
}
