package repository

import (
	"errors"
	"log"
	"rofl-bot/domain"

	"gorm.io/gorm"
)

type AuthRep struct {
	DB *gorm.DB
}

func NewAuthRep(db *gorm.DB) *AuthRep {
	return &AuthRep{
		DB: db,
	}
}

func (r *AuthRep) CreateToken(token *domain.Token, username string) error {
	var user domain.User
	err := r.DB.Where("id = ?", token.UserId).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		userObj := &domain.User{
			Id:       token.UserId,
			Username: username,
		}
		e := r.DB.Create(userObj).Error
		if e != nil {
			return e
		}
	}
	if err != nil {
		return err
	}

	return r.DB.Create(token).Error
}

func (r *AuthRep) GetToken(refresh string) (*domain.Token, error) {
	var token domain.Token
	err := r.DB.Where("refresh = ?", refresh).First(&token).Error
	return &token, err
}

func (r *AuthRep) GetTokenByUserId(userId uint) (*domain.Token, error) {
	var token domain.Token
	err := r.DB.Where("user_id = ?", userId).First(&token).Error
	return &token, err
}

func (r *AuthRep) UpdateToken(userID uint, token *domain.Token) error {
	err := r.DB.Where("user_id = ?", userID).Delete(&domain.Token{}).Error
	if err != nil {
		return err
	}
	return r.DB.Create(token).Error
}

func (r *AuthRep) GetTokenWithUser(refresh string) (*domain.Token, error) {
	var token domain.Token
	err := r.DB.Preload("User").Where("refresh = ?", refresh).First(&token).Error
	return &token, err
}

func (r *AuthRep) RevokeToken(refresh string) error {
	err := r.DB.Where("refresh = ?", refresh).Update("revoked", true).Error
	return err
}

func (r *AuthRep) DeleteToken(refresh string) error {
	err := r.DB.Where("refresh = ?", refresh).Delete(&domain.Token{}).Error
	return err
}

func (r *AuthRep) DeleteTokensByUserId(userId uint) error {
	log.Println(userId)
	err := r.DB.Where("user_id = ?", userId).Delete(&domain.Token{}).Error
	return err
}
