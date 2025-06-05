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

func (repo *UserGormRepo) Create(ctx context.Context, user *domain.User) error {
	return repo.db.WithContext(ctx).Create(user).Error
}

func (repo *UserGormRepo) List(ctx context.Context) ([]*domain.User, error) {
	var res []*domain.User
	err := repo.db.WithContext(ctx).Where(&domain.User{Name: "Test"}).Find(&res).Error
	return res, err
}
