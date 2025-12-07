package controller

import (
	"rofl-bot/domain"
	"rofl-bot/service"

	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserCtrl struct {
	service *service.UserService
}

func NewUserCtrl(service *service.UserService) *UserCtrl {
	return &UserCtrl{service: service}
}

// @Summary Получение пользователя
// @Description по id выдает пользоватля
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path uint true "User ID"
// @Success 200 {object} domain.User
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/user/{id} [get]
func (u *UserCtrl) GetUser(c *gin.Context) {
	var uri struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	id := uri.ID
	user, err := u.service.GetUser(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, domain.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(200, user)
}

// @Summary Обновление данных пользователя
// @Description принимает id, название банка и номер телефона (банк и номер не обязательны)
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user body domain.UpdateUser true "user"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/user [put]
func (u *UserCtrl) UpdateUser(c *gin.Context) {
	var body domain.UpdateUser
	err := c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, domain.ErrorResponse{Error: "Unauthorized or can't find user_id in context."})
		return
	}

	err = u.service.UpdateUser(user_id.(uint), &body)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, domain.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(200, domain.SuccessResponse{Message: "User updated."})
}

// @Summary Удаление пользователя
// @Description принимает id
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path uint true "User ID"
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/user/{id} [delete]
func (u *UserCtrl) DeleteUser(c *gin.Context) {
	id, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, domain.ErrorResponse{Error: "Unauthorized or can't find user_id in context."})
		return
	}

	err := u.service.DeleteUser(id.(uint))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, domain.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(200, domain.SuccessResponse{Message: "User deleted."})
}
