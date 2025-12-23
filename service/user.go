package service

import (
	"rofl-bot/domain"
	"rofl-bot/repository"
)

type UserService struct {
	rep *repository.UserRep
}

func NewUserService(rep *repository.UserRep) *UserService {
	return &UserService{rep: rep}
}

func (s *UserService) GetUser(id uint) (*domain.User, error) {
	return s.rep.GetUser(id)
}

func (s *UserService) CreateUser(user *domain.User) error {
	if user.Username == "" {
		user.Username = "Guest"
	}
	return s.rep.CreateUser(user)
}

func (s *UserService) UpdateUser(id uint, user *domain.UpdateUser) error {
	var currUser *domain.User
	var err error
	currUser, err = s.rep.GetUser(id)
	if err != nil {
		return err
	}

	if user.Bank != nil {
		currUser.Bank = *user.Bank
	}
	if user.PhoneNumber != nil {
		currUser.PhoneNumber = *user.PhoneNumber
	}
	return s.rep.UpdateUser(currUser)
}

func (s *UserService) DeleteUser(id uint) error {
	return s.rep.DeleteUser(id)
}
