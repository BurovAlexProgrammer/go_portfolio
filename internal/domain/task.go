package domain

import (
	"context"
	"time"
)

type Task struct {
	Id        int64  `gorm:"primaryKey" json:"id"`
	Name      string `gorm:"not null;uniqueIndex:idx_task_user" json:"name" binding:"required,min=3"`
	UserId    int64  `gorm:"not null;uniqueIndex:idx_task_user" json:"userId" binding:"required"`
	IsDone    bool   `json:"isDone" binding:"required"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	DoneByName(ctx context.Context, taskName string, userId int64) error
	CleanTasksByUserId(ctx context.Context, userId int64) error
}
