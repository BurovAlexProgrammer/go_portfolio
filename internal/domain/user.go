package domain

import (
	"context"
	"time"
)

type User struct {
	Id        int64  `gorm:"primaryKey" json:"id"`
	Name      string `gorm:"not null" json:"name" binding:"required,min=3"`
	Telegram  string `gorm:"not null;uniqueIndex" json:"telegram" binding:"required"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetByTelegramName(ctx context.Context, tgName string) (*User, error)
	//GetById(ctx context.Context, id int64) (*model.User, error)
	//Update(ctx context.Context, user *model.User) error
	//Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]*User, error)
}
