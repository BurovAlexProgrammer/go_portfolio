package usecase

import (
	"GoPortfolio/internal/domain"
	"context"
)

type UserUsecase struct {
	repo domain.UserRepository
}

func NewUserUsecase(repo domain.UserRepository) *UserUsecase {
	return &UserUsecase{
		repo: repo,
	}
}

func (u UserUsecase) CreateUser(ctx context.Context, user *domain.User) error {
	const op = "user_usecase.CreateUser"

	err := u.repo.Create(ctx, user)
	if err != nil {
		return err
	}

	return nil
}
