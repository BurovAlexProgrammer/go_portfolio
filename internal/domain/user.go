package domain

import "context"

type User struct {
	Id       int64  `db:"id" json:"id"`
	Name     string `db:"name" json:"name" binding:"required,min=3"`
	Telegram string `db:"telegram" json:"telegram" binding:"required"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	//GetById(ctx context.Context, id int64) (*model.User, error)
	//Update(ctx context.Context, user *model.User) error
	//Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]*User, error)
}
