package repository

import (
	"GoPortfolio/internal/model"
	"context"
)

type Repository interface {
	Create(ctx context.Context, user *model.User) error
	//GetById(ctx context.Context, id int64) (*model.User, error)
	//Update(ctx context.Context, user *model.User) error
	//Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]*model.User, error)
}
