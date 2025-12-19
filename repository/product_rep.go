package repository

import (
	"rofl-bot/domain"

	"gorm.io/gorm"
)

type ProductRep struct {
	DB *gorm.DB
}

func NewProductRep(db *gorm.DB) *ProductRep {
	return &ProductRep{
		DB: db,
	}
}

func (r *ProductRep) CreateProduct(eventId uint, product *domain.CreateProduct) (uint, error) {
	var newProduct domain.Product
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		newProduct = domain.Product{
			EventId:    eventId,
			Name:       product.Name,
			Price:      product.Price,
			ConsumerId: product.ConsumerId,
			Bought:     product.Bought,
			Equal:      product.Equal,
		}

		if err := tx.Create(&newProduct).Error; err != nil {
			return err
		}

		if len(product.Members) > 0 && !product.Equal {
			members := make([]domain.ProductMember, 0, len(product.Members))
			for _, member := range product.Members {
				members = append(members, domain.ProductMember{
					ProductId: newProduct.Id,
					UserId:    member.UserId,
					Price:     *member.Price,
				})
			}

			err := tx.Create(&members).Error
			if err != nil {
				return err
			}
		} else if len(product.Members) > 0 && product.Equal {
			ppm := newProduct.Price / float64(len(product.Members))
			members := make([]domain.ProductMember, 0, len(product.Members))
			for _, member := range product.Members {
				members = append(members, domain.ProductMember{
					ProductId: newProduct.Id,
					UserId:    member.UserId,
					Price:     ppm,
				})
			}

			err := tx.Create(&members).Error
			if err != nil {
				return err
			}
		}

		return nil
	})

	return newProduct.Id, err
}

func (r *ProductRep) GetProduct(id uint) (*domain.Product, error) {
	var product domain.Product
	err := r.DB.
		Preload("Consumer").
		Preload("Members.User").
		Where("id = ?", id).
		First(&product).Error

	if err != nil {
		return nil, err
	}
	return &product, err
}

func (r *ProductRep) GetProductsByEvent(eventId uint) ([]domain.ProductList, error) {
	var products []domain.ProductList
	err := r.DB.
		Table("products").
		Select(`products.id, products.name, products.price, COUNT(DISTINCT pm.user_id) AS count`).
		Joins("LEFT JOIN product_members pm ON pm.product_id = products.id").
		Where("products.event_id = ?", eventId).
		Group("products.id").
		Scan(&products).Error
	return products, err
}

func (r *ProductRep) CheckMember(eventId, userId uint) (bool, error) {
	var eventMember domain.EventMember
	err := r.DB.Where("event_id = ? and user_id = ?", eventId, userId).First(&eventMember).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *ProductRep) CheckAdmin(eventId, userId uint) (bool, error) {
	var eventMember domain.EventMember
	err := r.DB.Preload("Role").Where("event_id = ? and user_id = ?", eventId, userId).First(&eventMember).Error
	if err != nil {
		return false, err
	}
	return eventMember.Role.Name == "admin", nil
}

func (r *ProductRep) DeleteProduct(id uint) error {
	return r.DB.Where("id = ?", id).Delete(&domain.Product{}).Error
}

func (r *ProductRep) UpdateProduct(id uint, newProduct *domain.UpdateProduct) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		var product domain.Product
		if err := tx.Preload("Members").First(&product, id).Error; err != nil {
			return err
		}

		product.Name = newProduct.Name
		product.Price = newProduct.Price
		product.Bought = newProduct.Bought
		product.Equal = newProduct.Equal

		if err := tx.Save(&product).Error; err != nil {
			return err
		}

		curMembers := make(map[uint]domain.ProductMember)
		for _, member := range product.Members {
			curMembers[member.UserId] = member
		}

		newMembers := make(map[uint]domain.UpdateProductMember)
		for _, member := range newProduct.Members {
			newMembers[member.UserId] = member
		}

		var toDelete []uint
		for userId := range curMembers {
			if _, ok := newMembers[userId]; !ok {
				toDelete = append(toDelete, userId)
			}
		}
		if len(toDelete) > 0 {
			if err := tx.Where("product_id = ? and user_id in (?)", id, toDelete).Delete(&domain.ProductMember{}).Error; err != nil {
				return err
			}
		}

		num := len(newProduct.Members)
		for userId, newMember := range newMembers {
			var price float64
			if newProduct.Equal {
				price = product.Price / float64(num)
			} else {
				if newMember.Price != nil {
					price = *newMember.Price
				} else {
					price = 0
				}
			}

			if oldMember, ok := curMembers[userId]; ok {
				if oldMember.Price != price {
					if err := tx.Model(&domain.ProductMember{}).Where("product_id = ? and user_id = ?", id, userId).Update("price", price).Error; err != nil {
						return err
					}
				}
			} else {
				m := domain.ProductMember{
					UserId:    userId,
					ProductId: id,
					Price:     price,
				}

				if err := tx.Create(&m).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}
