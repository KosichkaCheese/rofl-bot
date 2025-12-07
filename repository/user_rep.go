package repository

import (
	"rofl-bot/domain"

	"gorm.io/gorm"
)

type UserRep struct {
	DB *gorm.DB
}

func NewUserRep(db *gorm.DB) *UserRep {
	return &UserRep{
		DB: db,
	}
}

func (r *UserRep) CreateUser(user *domain.User) error {
	return r.DB.Create(user).Error
}

func (r *UserRep) GetUser(id uint) (*domain.User, error) {
	var user domain.User
	err := r.DB.Where("id = ?", id).First(&user).Error
	return &user, err
}

func (r *UserRep) UpdateUser(user *domain.User) error {
	return r.DB.Save(user).Error
}

func (r *UserRep) DeleteUser(id uint) error {
	return r.DB.Where("id = ?", id).Delete(&domain.User{}).Error
}
