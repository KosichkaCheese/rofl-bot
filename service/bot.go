package service

import (
	"crypto/subtle"
	"errors"
	"rofl-bot/config"
	"rofl-bot/domain"
	"rofl-bot/repository"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Бот сравнивает текст ошибки дословно, чтобы показать пользователю понятное сообщение.
var ErrAlreadyMember = errors.New("user is already member of this event")

type BotService struct {
	rep    *repository.BotRep
	events *EventService
}

func NewBotService(rep *repository.BotRep, events *EventService) *BotService {
	return &BotService{rep: rep, events: events}
}

// Без @username показываем имя, как это делает бот.
func toUser(user domain.BotUser) *domain.User {
	username := user.Username
	if username == "" {
		username = user.FirstName
	}
	if username == "" {
		username = "Guest"
	}
	return &domain.User{Id: user.Id, Username: username}
}

// Бот ждёт время начала в UTC.
func utc(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	u := t.UTC()
	return &u
}

func (s *BotService) CreateEvent(req *domain.BotCreateEvent) (uint, error) {
	chatId := req.ChatId
	event := &domain.Event{Name: req.Name, StartsAt: utc(req.StartsAt), ChatId: &chatId}
	return s.rep.CreateEvent(toUser(req.User), event)
}

func (s *BotService) JoinEvent(event_id uint, user domain.BotUser) error {
	if _, err := s.events.rep.GetEvent(event_id); err != nil {
		return err
	}
	if err := s.rep.UpsertUser(toUser(user)); err != nil {
		return err
	}

	isMember, err := s.events.rep.CheckMember(event_id, user.Id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if isMember {
		return ErrAlreadyMember
	}

	return s.events.rep.JoinEvent(event_id, user.Id)
}

func (s *BotService) GetEvent(event_id uint) (*domain.BotEventDetails, error) {
	event, err := s.rep.GetEvent(event_id)
	if err != nil {
		return nil, err
	}
	members, err := s.rep.GetEventMembers([]uint{event_id})
	if err != nil {
		return nil, err
	}

	event.StartsAt = utc(event.StartsAt)
	details := &domain.BotEventDetails{BotEvent: *event, Members: members[event_id]}
	if details.Members == nil {
		details.Members = make([]domain.BotEventMember, 0)
	}
	return details, nil
}

func (s *BotService) GetChatEvents(chatId int64) ([]domain.BotEvent, error) {
	events, err := s.rep.GetChatEvents(chatId)
	if err != nil {
		return nil, err
	}
	for i := range events {
		events[i].StartsAt = utc(events[i].StartsAt)
	}
	return events, nil
}

func (s *BotService) CreateInvite(event_id uint) (*domain.InviteLink, error) {
	if _, err := s.events.rep.GetEvent(event_id); err != nil {
		return nil, err
	}
	return s.events.IssueInvite(event_id)
}

func (s *BotService) ClaimReminders(req *domain.BotClaimReminders) ([]domain.BotReminderEvent, error) {
	events, err := s.rep.ClaimReminders(req.WindowHours, req.AfterHours)
	if err != nil {
		return nil, err
	}
	for i := range events {
		events[i].StartsAt = events[i].StartsAt.UTC()
		if events[i].Members == nil {
			events[i].Members = make([]domain.BotEventMember, 0)
		}
	}
	return events, nil
}

func (s *BotService) MigrateChat(req *domain.BotMigrateChat) error {
	return s.rep.MigrateChat(req.FromChatId, req.ToChatId)
}

// Пускает только бота: заголовок X-Bot-Secret должен совпасть с BOT_API_SECRET.
func BotAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := c.GetHeader("X-Bot-Secret")
		if cfg.BOT_API_SECRET == "" ||
			subtle.ConstantTimeCompare([]byte(secret), []byte(cfg.BOT_API_SECRET)) != 1 {
			c.JSON(401, domain.ErrorResponse{Error: "Unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}
