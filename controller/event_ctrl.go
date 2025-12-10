package controller

import (
	"rofl-bot/domain"
	"rofl-bot/service"

	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventCtrl struct {
	service *service.EventService
}

func NewEventCtrl(service *service.EventService) *EventCtrl {
	return &EventCtrl{service: service}
}

// @Summary Получение события
// @Description по id выдает событие. !не работает, если запрашивает не участник ивента!
// @Tags event
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path uint true "Event ID"
// @Success 200 {object} domain.Event
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/event/{id} [get]
func (ctrl *EventCtrl) GetEvent(c *gin.Context) {
	var uri struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	id := uri.ID
	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, domain.ErrorResponse{Error: "Unauthorized or can't find user_id in context."})
		return
	}

	event, err := ctrl.service.GetEvent(user_id.(uint), uint(id))
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

	c.JSON(200, event)
}

// @Summary Получение админа события
// @Description по id события выдает админа. !не работает, если запрашивает не участник ивента!
// @Tags event
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path uint true "Event ID"
// @Success 200 {object} domain.User
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/event/{id}/admin [get]
func (ctrl *EventCtrl) GetEventAdmin(c *gin.Context) {
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

	admin, err := ctrl.service.GetEventAdmin(event_id, user_id.(uint))
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

	c.JSON(200, admin)
}

// @Summary Получение всех ивентов текущего юзера
// @Description выдает все ивенты авторизованного пользователя
// @Tags event
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} []domain.EventWithMembersCount
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/events [get]
func (ctrl *EventCtrl) GetUserEvents(c *gin.Context) {
	id, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, domain.ErrorResponse{Error: "Unauthorized or can't find user_id in context."})
		return
	}

	events, err := ctrl.service.GetUserEvents(id.(uint))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, domain.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(200, events)
}

// @Summary Создание события
// @Description Создает событие, при успехе возвращает id. авторизованный пользователь становится админом события.
// @Tags event
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param event body domain.CreateEvent true "Event"
// @Success 200 {object} uint
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/event [post]
func (ctrl *EventCtrl) CreateEvent(c *gin.Context) {
	var body domain.CreateEvent
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

	event_id, err := ctrl.service.CreateEvent(user_id.(uint), &body)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, domain.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(200, event_id)
}

// @Summary Обновление события
// @Description Можно изменить название или завершенность. оба параметра необязательны. !Изменить событие может только его админ!
// @Tags event
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param event body domain.UpdateEvent true "Event"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/event [put]
func (ctrl *EventCtrl) UpdateEvent(c *gin.Context) {
	var body domain.UpdateEvent
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

	err = ctrl.service.UpdateEvent(user_id.(uint), &body)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, domain.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(200, domain.SuccessResponse{Message: "Event updated."})
}

// @Summary Удаление события
// @Description по id удаляет событие. !Удалить событие может только его админ!
// @Tags event
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path uint true "Event ID"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/event/{id} [delete]
func (ctrl *EventCtrl) DeleteEvent(c *gin.Context) {
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

	err := ctrl.service.DeleteEvent(user_id.(uint), event_id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, domain.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(200, domain.SuccessResponse{Message: "Event deleted."})
}

// @Summary Присоединиться к ивенту.
// @Description по id invite-ссылки присоединяет авторизованного пользователя к ивенту. не работает для пользователей, которые уже состоят в этом ивенте.
// @Tags event
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Invite ID (uuid)"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/invite/{id} [post]
func (ctrl *EventCtrl) JoinEvent(c *gin.Context) {
	var uri struct {
		ID string `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	invite_id, err := uuid.Parse(uri.ID)
	if err != nil {
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, domain.ErrorResponse{Error: "Unauthorized or can't find user_id in context."})
		return
	}

	err = ctrl.service.JoinEvent(user_id.(uint), invite_id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, domain.ErrorResponse{Error: "Invite is invalid"})
			return
		}
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(200, domain.SuccessResponse{Message: "Event joined."})
}

// @Summary Покинуть ивент
// @Description по id ивента удаляет из него авторизованного пользователя. !не работает, если пользователь не участник ивента!
// @Tags event
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path uint true "Event ID"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/event/{id}/leave [put]
func (ctrl *EventCtrl) LeaveEvent(c *gin.Context) {
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

	err := ctrl.service.LeaveEvent(user_id.(uint), event_id)
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

	c.JSON(200, domain.SuccessResponse{Message: "Event left."})
}

// @Summary Получение участников ивента
// @Description по id ивента получает всех участников !не работает, если пользователь не участник ивента!
// @Tags event
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path uint true "Event ID"
// @Success 200 {object} []domain.EventMembers
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/event/{id}/members [get]
func (ctrl *EventCtrl) GetEventMembers(c *gin.Context) {
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

	members, err := ctrl.service.GetEventMembers(user_id.(uint), event_id)
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

	c.JSON(200, members)
}

// @Summary Создание ссылки для приглашения
// @Description Создает ссылку для приглашения в ивент с указанным id. не работает, если пользователь не участник ивента. Ссылка живет 24 часа, при повторном использовании старая ссылка заменяется на новую.
// @Tags event
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path uint true "Event ID"
// @Success 200 {object} domain.InviteLink
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/event/{id}/invite [post]
func (ctrl *EventCtrl) CreateInvite(c *gin.Context) {
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

	link, err := ctrl.service.CreateInvite(user_id.(uint), event_id)
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

	c.JSON(200, link)
}
