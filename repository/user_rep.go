package repository

import (
	"rofl-bot/domain"

	"gorm.io/gorm"
)

type UserRep struct {
	DB *gorm.DB
}

func NewUserRep(db *gorm.DB) *AuthRep {
	return &AuthRep{
		DB: db,
	}
}

func (r *UserRep) CreateUser(user *domain.User) error {
	return r.DB.Create(user).Error
}
