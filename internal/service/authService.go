package service

import (
	"GoPortfolio/internal/domain"
	"context"
)

type AuthService struct {
	tgUserMap map[string]*domain.User
	userRepo  domain.UserRepository
}

func NewAuthService(u domain.UserRepository) *AuthService {
	return &AuthService{
		tgUserMap: make(map[string]*domain.User, 16),
		userRepo:  u,
	}
}

func (a *AuthService) RegisterByTelegramIfNecessary(ctx context.Context, tgName string, firstName string) error {
	existUser, _ := a.GetExistUser(ctx, tgName)
	if existUser != nil {
		return nil
	}

	user := domain.User{
		Telegram: tgName,
		Name:     firstName,
	}

	_, err := a.RegisterUser(ctx, &user)
	return err
}

func (a *AuthService) RegisterUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	user, err := a.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	a.tgUserMap[user.Telegram] = user
	return user, nil
}

func (a *AuthService) GetExistUser(ctx context.Context, userTgName string) (*domain.User, error) {
	if a.tgUserMap[userTgName] != nil {
		return a.tgUserMap[userTgName], nil
	}
	user, err := a.userRepo.GetByTelegramName(ctx, userTgName)
	if err != nil {
		return nil, err
	}
	a.tgUserMap[userTgName] = user
	return user, nil
}
