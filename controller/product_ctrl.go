package controller

import (
	"errors"
	"rofl-bot/domain"
	"rofl-bot/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductCtrl struct {
	service *service.ProductService
}

func NewProductCtrl(service *service.ProductService) *ProductCtrl {
	return &ProductCtrl{service: service}
}

// @Summary Создание товара
// @Description Создает товар, при успехе возвращает id. Не работает, если создает не участнк события.
// @Tags product
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path uint true "Event ID"
// @Param product body domain.CreateProduct true "Product"
// @Success 200 {object} uint
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/event/{id}/product [post]
func (ctrl *ProductCtrl) CreateProduct(c *gin.Context) {
	var body domain.CreateProduct
	err := c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	var uri struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	event_id := uri.ID

	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, domain.ErrorResponse{Error: "Unauthorized or can't find user_id in context."})
		return
	}

	productId, err := ctrl.service.CreateProduct(event_id, user_id.(uint), &body)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, domain.ErrorResponse{Error: err.Error()})
			return
		}
		if err.Error() == "user is not member of this event" {
			c.JSON(400, domain.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(200, productId)
}

// @Summary Получение товара
// @Description по id ивента и товара выдает товар. !не работает, если запрашивает не участник ивента!
// @Tags product
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param event_id path uint true "Event ID"
// @Param product_id path uint true "Product ID"
// @Success 200 {object} domain.ProductResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/event/{event_id}/product/{product_id} [get]
func (ctrl *ProductCtrl) GetProduct(c *gin.Context) {
	var uri struct {
		EventID   uint `uri:"id" binding:"required"`
		ProductID uint `uri:"product_id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	event_id := uri.EventID
	product_id := uri.ProductID

	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, domain.ErrorResponse{Error: "Unauthorized or can't find user_id in context."})
		return
	}

	product, err := ctrl.service.GetProduct(event_id, user_id.(uint), product_id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, domain.ErrorResponse{Error: err.Error()})
			return
		}
		if err.Error() == "user is not member of this event" {
			c.JSON(400, domain.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(200, product)
}

// @Summary Получение всех товаров в ивенте
// @Description выдает все товары по id ивента. не работает, если пользователь не участник ивента
// @Tags product
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path uint true "Event ID"s
// @Success 200 {object} []domain.ProductList
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/event/{id}/products [get]
func (ctrl *ProductCtrl) GetProductsByEvent(c *gin.Context) {
	var uri struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	event_id := uri.ID

	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, domain.ErrorResponse{Error: "Unauthorized or can't find user_id in context."})
		return
	}

	products, err := ctrl.service.GetProductsByEvent(event_id, user_id.(uint))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, domain.ErrorResponse{Error: err.Error()})
			return
		}
		if err.Error() == "user is not member of this event" {
			c.JSON(400, domain.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(200, products)
}

// @Summary Удаление товара
// @Description по id ивента и товара удаляет товар. !не работает, если запрашивает не админ ивента!
// @Tags product
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param event_id path uint true "Event ID"
// @Param product_id path uint true "Product ID"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/event/{event_id}/product/{product_id} [delete]
func (ctrl *ProductCtrl) DeleteProduct(c *gin.Context) {
	var uri struct {
		EventID   uint `uri:"id" binding:"required"`
		ProductID uint `uri:"product_id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	event_id := uri.EventID
	product_id := uri.ProductID

	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, domain.ErrorResponse{Error: "Unauthorized or can't find user_id in context."})
		return
	}

	err := ctrl.service.DeleteProduct(event_id, user_id.(uint), product_id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, domain.ErrorResponse{Error: err.Error()})
			return
		}
		if err.Error() == "user is not admin of this event" {
			c.JSON(400, domain.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(200, domain.SuccessResponse{Message: "Product deleted."})
}

// @Summary Обновление товара
// @Description по id ивента и товара обновляет товар. !не работает, если запрашивает не участник ивента!
// @Tags product
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param event_id path uint true "Event ID"
// @Param product_id path uint true "Product ID"
// @Param product body domain.UpdateProduct true "Product"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/event/{event_id}/product/{product_id} [put]
func (ctrl *ProductCtrl) UpdateProduct(c *gin.Context) {
	var uri struct {
		EventID   uint `uri:"id" binding:"required"`
		ProductID uint `uri:"product_id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	var product domain.UpdateProduct
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	event_id := uri.EventID
	product_id := uri.ProductID

	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, domain.ErrorResponse{Error: "Unauthorized or can't find user_id in context."})
		return
	}

	err := ctrl.service.UpdateProduct(event_id, user_id.(uint), product_id, &product)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, domain.ErrorResponse{Error: err.Error()})
			return
		}
		if err.Error() == "user is not member of this event" {
			c.JSON(400, domain.ErrorResponse{Error: err.Error()})
			return
		} else if err.Error() == "sum of member prices must be equal to product price" {
			c.JSON(400, domain.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(200, domain.SuccessResponse{Message: "Product updated."})
}
