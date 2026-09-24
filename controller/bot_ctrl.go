package controller

import (
	"rofl-bot/domain"
	"rofl-bot/service"

	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BotCtrl struct {
	service *service.BotService
}

func NewBotCtrl(service *service.BotService) *BotCtrl {
	return &BotCtrl{service: service}
}

func botError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(404, domain.ErrorResponse{Error: err.Error()})
	case errors.Is(err, service.ErrNotAdmin):
		c.JSON(403, domain.ErrorResponse{Error: err.Error()})
	case errors.Is(err, service.ErrAlreadyMember):
		c.JSON(409, domain.ErrorResponse{Error: err.Error()})
	default:
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
	}
}

func bindEventId(c *gin.Context) (uint, bool) {
	var uri struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return 0, false
	}
	return uri.ID, true
}

// @Summary Создание события ботом
// @Description Создает событие в чате, при успехе возвращает id. user становится админом события (создается, если его нет).
// @Tags bot
// @Security BotSecret
// @Accept json
// @Produce json
// @Param event body domain.BotCreateEvent true "Event"
// @Success 200 {object} uint
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/bot/event [post]
func (ctrl *BotCtrl) CreateEvent(c *gin.Context) {
	var body domain.BotCreateEvent
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	event_id, err := ctrl.service.CreateEvent(&body)
	if err != nil {
		botError(c, err)
		return
	}

	c.JSON(200, event_id)
}

// @Summary Запись на событие через бота
// @Description Добавляет user в участники события (создается, если его нет).
// @Tags bot
// @Security BotSecret
// @Accept json
// @Param id path uint true "Event ID"
// @Param user body domain.BotJoinEvent true "User"
// @Success 204
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 409 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/bot/event/{id}/join [post]
func (ctrl *BotCtrl) JoinEvent(c *gin.Context) {
	event_id, ok := bindEventId(c)
	if !ok {
		return
	}
	var body domain.BotJoinEvent
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	if err := ctrl.service.JoinEvent(event_id, body.User); err != nil {
		botError(c, err)
		return
	}

	c.Status(204)
}

// @Summary Событие для бота
// @Description Событие с числом и списком участников, без проверки членства.
// @Tags bot
// @Security BotSecret
// @Produce json
// @Param id path uint true "Event ID"
// @Success 200 {object} domain.BotEventDetails
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/bot/event/{id} [get]
func (ctrl *BotCtrl) GetEvent(c *gin.Context) {
	event_id, ok := bindEventId(c)
	if !ok {
		return
	}

	event, err := ctrl.service.GetEvent(event_id)
	if err != nil {
		botError(c, err)
		return
	}

	c.JSON(200, event)
}

// @Summary События чата
// @Description Незавершенные события чата, которые еще не начались: сначала ближайшие, без даты — в конце. С admin_id — только события, где этот пользователь админ.
// @Tags bot
// @Security BotSecret
// @Produce json
// @Param chat_id query int true "Telegram chat ID"
// @Param admin_id query int false "Telegram user ID админа события"
// @Success 200 {object} []domain.BotEvent
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/bot/events [get]
func (ctrl *BotCtrl) GetChatEvents(c *gin.Context) {
	chatId, err := strconv.ParseInt(c.Query("chat_id"), 10, 64)
	if err != nil {
		c.JSON(400, domain.ErrorResponse{Error: "chat_id must be an integer"})
		return
	}

	var adminId *uint
	if raw := c.Query("admin_id"); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			c.JSON(400, domain.ErrorResponse{Error: "admin_id must be a positive integer"})
			return
		}
		admin := uint(id)
		adminId = &admin
	}

	events, err := ctrl.service.GetChatEvents(chatId, adminId)
	if err != nil {
		botError(c, err)
		return
	}

	c.JSON(200, events)
}

// @Summary Приглашение через бота
// @Description Создает новую пригласительную ссылку события, старая перестает работать.
// @Tags bot
// @Security BotSecret
// @Produce json
// @Param id path uint true "Event ID"
// @Success 200 {object} domain.InviteLink
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/bot/event/{id}/invite [post]
func (ctrl *BotCtrl) CreateInvite(c *gin.Context) {
	event_id, ok := bindEventId(c)
	if !ok {
		return
	}

	invite, err := ctrl.service.CreateInvite(event_id)
	if err != nil {
		botError(c, err)
		return
	}

	c.JSON(200, invite)
}

// @Summary Удаление события через бота
// @Description Удаляет событие вместе с участниками, покупками и приглашениями. Удалить может только админ события. POST, а не DELETE: тело у DELETE прокси могут отбросить.
// @Tags bot
// @Security BotSecret
// @Accept json
// @Param id path uint true "Event ID"
// @Param user body domain.BotDeleteEvent true "User"
// @Success 204
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/bot/event/{id}/delete [post]
func (ctrl *BotCtrl) DeleteEvent(c *gin.Context) {
	event_id, ok := bindEventId(c)
	if !ok {
		return
	}
	var body domain.BotDeleteEvent
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	if err := ctrl.service.DeleteEvent(event_id, body.User); err != nil {
		botError(c, err)
		return
	}

	c.Status(204)
}

// @Summary События для напоминаний
// @Description Выдает события, до начала которых больше after_hours и не больше window_hours часов, и о которых в этом окне еще не напоминали. Выданные события сразу помечаются напомненными.
// @Tags bot
// @Security BotSecret
// @Accept json
// @Produce json
// @Param window body domain.BotClaimReminders true "Window"
// @Success 200 {object} []domain.BotReminderEvent
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/bot/reminders/claim [post]
func (ctrl *BotCtrl) ClaimReminders(c *gin.Context) {
	var body domain.BotClaimReminders
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	events, err := ctrl.service.ClaimReminders(&body)
	if err != nil {
		botError(c, err)
		return
	}

	c.JSON(200, events)
}

// @Summary Миграция чата
// @Description Группа стала супергруппой — переносит ее события на новый chat_id.
// @Tags bot
// @Security BotSecret
// @Accept json
// @Param chats body domain.BotMigrateChat true "Chats"
// @Success 204
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /alena-rofl/bot/chat/migrate [post]
func (ctrl *BotCtrl) MigrateChat(c *gin.Context) {
	var body domain.BotMigrateChat
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	if err := ctrl.service.MigrateChat(&body); err != nil {
		botError(c, err)
		return
	}

	c.Status(204)
}
