package service

import (
	"errors"
	"rofl-bot/domain"
	"rofl-bot/repository"
)

type ProductService struct {
	rep *repository.ProductRep
}

func NewProductService(rep *repository.ProductRep) *ProductService {
	return &ProductService{rep: rep}
}

func (s *ProductService) CreateProduct(eventId, userId uint, product *domain.CreateProduct) (uint, error) {
	isMember, err := s.rep.CheckMember(eventId, userId)
	if err != nil && err.Error() != "record not found" {
		return 0, err
	}
	if !isMember {
		return 0, errors.New("user is not member of this event")
	}

	id, err := s.rep.CreateProduct(eventId, product)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *ProductService) GetProduct(eventId, userId, productId uint) (*domain.ProductResponse, error) {
	isMember, err := s.rep.CheckMember(eventId, userId)
	if err != nil && err.Error() != "record not found" {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("user is not member of this event")
	}

	product, err := s.rep.GetProduct(productId)
	if err != nil {
		return nil, err
	}

	members := make([]domain.ProductMemberResponse, 0, len(product.Members))
	for _, m := range product.Members {
		members = append(members, domain.ProductMemberResponse{
			UserId:   m.UserId,
			Username: m.User.Username,
			Price:    m.Price})
	}

	return &domain.ProductResponse{
		Id:               product.Id,
		Name:             product.Name,
		Price:            product.Price,
		Bought:           product.Bought,
		Equal:            product.Equal,
		ConsumerId:       product.Consumer.Id,
		ConsumerUsername: product.Consumer.Username,
		Members:          members,
	}, nil
}

func (s *ProductService) GetProductsByEvent(eventId, userId uint) ([]domain.ProductList, error) {
	isMember, err := s.rep.CheckMember(eventId, userId)
	if err != nil && err.Error() != "record not found" {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("user is not member of this event")
	}

	return s.rep.GetProductsByEvent(eventId)
}

func (s *ProductService) DeleteProduct(eventId, userId, productId uint) error {
	isAdmin, err := s.rep.CheckAdmin(eventId, userId)
	if err != nil && err.Error() != "record not found" {
		return err
	}
	if !isAdmin {
		return errors.New("user is not admin of this event")
	}

	return s.rep.DeleteProduct(productId)
}

func (s *ProductService) UpdateProduct(eventId, userId, productId uint, product *domain.UpdateProduct) error {
	isMember, err := s.rep.CheckMember(eventId, userId)
	if err != nil && err.Error() != "record not found" {
		return err
	}
	if !isMember {
		return errors.New("user is not member of this event")
	}

	if !product.Equal {
		var sum float64 = 0
		for _, m := range product.Members {
			sum += *m.Price
		}

		if sum != product.Price {
			return errors.New("sum of member prices must be equal to product price")
		}
	}

	return s.rep.UpdateProduct(productId, product)
}
