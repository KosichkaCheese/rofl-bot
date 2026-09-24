package domain

import "time"

// Сервисный API Telegram-бота: /alena-rofl/bot/*, авторизация по X-Bot-Secret.

// Пользователь Telegram, от имени которого действует бот.
type BotUser struct {
	Id        uint   `json:"id" binding:"required"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
}

type BotCreateEvent struct {
	Name     string     `json:"name" binding:"required"`
	StartsAt *time.Time `json:"starts_at"`
	ChatId   int64      `json:"chat_id" binding:"required"`
	User     BotUser    `json:"user"`
}

type BotJoinEvent struct {
	User BotUser `json:"user"`
}

type BotDeleteEvent struct {
	User BotUser `json:"user"`
}

type BotClaimReminders struct {
	WindowHours int `json:"window_hours" binding:"required,gt=0"`
	AfterHours  int `json:"after_hours" binding:"gte=0,ltfield=WindowHours"`
}

type BotMigrateChat struct {
	FromChatId int64 `json:"from_chat_id" binding:"required"`
	ToChatId   int64 `json:"to_chat_id" binding:"required"`
}

type BotEventMember struct {
	UserId   uint   `json:"user_id"`
	Username string `json:"username"`
}

type BotEvent struct {
	Id       uint       `json:"id"`
	Name     string     `json:"name"`
	StartsAt *time.Time `json:"starts_at"`
	Count    int        `json:"count"`
}

type BotEventDetails struct {
	BotEvent
	Members []BotEventMember `json:"members"`
}

type BotReminderEvent struct {
	Id       uint             `json:"id"`
	Name     string           `json:"name"`
	StartsAt time.Time        `json:"starts_at"`
	ChatId   int64            `json:"chat_id"`
	Members  []BotEventMember `json:"members" gorm:"-"`
}
