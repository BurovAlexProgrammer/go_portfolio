package gorm

import (
	"GoPortfolio/internal/domain"
	"context"
	"gorm.io/gorm"
)

type UserGormRepo struct {
	db *gorm.DB
}

func NewUserGormRepo(db *gorm.DB) *UserGormRepo {
	return &UserGormRepo{db: db}
}

func (repo *UserGormRepo) List(ctx context.Context) ([]*domain.User, error) {
	var res []*domain.User
	err := repo.db.WithContext(ctx).Where(&domain.User{Name: "Test"}).Find(&res).Error
	return res, err
}

func (repo *UserGormRepo) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	err := repo.db.WithContext(ctx).Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *UserGormRepo) GetByTelegramName(ctx context.Context, tgName string) (*domain.User, error) {
	user := &domain.User{}
	err := repo.db.WithContext(ctx).First(user, domain.User{Telegram: tgName}).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}
