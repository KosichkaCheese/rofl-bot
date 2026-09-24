package repository

import (
	"rofl-bot/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BotRep struct {
	DB *gorm.DB
}

func NewBotRep(db *gorm.DB) *BotRep {
	return &BotRep{
		DB: db,
	}
}

// Человек мог ни разу не открывать мини-апп, поэтому пользователь создаётся здесь же.
func upsertUser(db *gorm.DB, user *domain.User) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"username", "updated_at"}),
	}).Create(user).Error
}

func (r *BotRep) UpsertUser(user *domain.User) error {
	return upsertUser(r.DB, user)
}

func (r *BotRep) CreateEvent(user *domain.User, event *domain.Event) (uint, error) {
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := upsertUser(tx, user); err != nil {
			return err
		}
		if err := tx.Create(event).Error; err != nil {
			return err
		}
		return tx.Create(&domain.EventMember{UserId: user.Id, EventId: event.Id, RoleId: 1}).Error
	})
	return event.Id, err
}

func (r *BotRep) GetEvent(id uint) (*domain.BotEvent, error) {
	var event domain.BotEvent
	err := r.DB.
		Table("events").
		Select("events.id, events.name, events.starts_at, COUNT(em.user_id) AS count").
		Joins("LEFT JOIN event_members em ON em.event_id = events.id").
		Where("events.id = ?", id).
		Group("events.id").
		Take(&event).Error
	return &event, err
}

func (r *BotRep) GetEventMembers(eventIds []uint) (map[uint][]domain.BotEventMember, error) {
	var rows []struct {
		EventId uint
		domain.BotEventMember
	}
	err := r.DB.
		Table("event_members em").
		Select("em.event_id, u.id AS user_id, u.username").
		Joins("JOIN users u ON u.id = em.user_id").
		Where("em.event_id IN ?", eventIds).
		Order("em.id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	members := make(map[uint][]domain.BotEventMember, len(eventIds))
	for _, row := range rows {
		members[row.EventId] = append(members[row.EventId], row.BotEventMember)
	}
	return members, nil
}

// Незавершённые события чата, которые ещё не начались: сначала ближайшие, без даты — в конце.
// С adminId — только события, где этот пользователь администратор.
func (r *BotRep) GetChatEvents(chatId int64, adminId *uint) ([]domain.BotEvent, error) {
	events := make([]domain.BotEvent, 0)
	query := r.DB.
		Table("events").
		Select("events.id, events.name, events.starts_at, COUNT(em.user_id) AS count").
		Joins("LEFT JOIN event_members em ON em.event_id = events.id").
		Where("events.chat_id = ? AND NOT events.ended", chatId).
		Where("events.starts_at IS NULL OR events.starts_at > now()")
	if adminId != nil {
		// Отдельный подзапрос, чтобы COUNT считал всех участников, а не только админа. role_id 1 — admin.
		query = query.Where(
			"EXISTS (SELECT 1 FROM event_members a WHERE a.event_id = events.id AND a.user_id = ? AND a.role_id = 1)",
			*adminId,
		)
	}
	err := query.
		Group("events.id").
		Order("events.starts_at ASC NULLS LAST, events.id").
		Scan(&events).Error
	return events, err
}

// Выдаёт события, до начала которых больше afterHours и не больше windowHours часов,
// и сразу помечает их напомненными для этого окна. Благодаря ON CONFLICT по первичному
// ключу event_reminders одно событие не выдаётся дважды даже при нескольких экземплярах бота.
func (r *BotRep) ClaimReminders(windowHours, afterHours int) ([]domain.BotReminderEvent, error) {
	events := make([]domain.BotReminderEvent, 0)
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		var ids []uint
		err := tx.Raw(`
			INSERT INTO event_reminders (event_id, window_hours, sent_at)
			SELECT id, ?, now() FROM events
			WHERE chat_id IS NOT NULL AND NOT ended
			  AND starts_at >  now() + make_interval(hours => ?)
			  AND starts_at <= now() + make_interval(hours => ?)
			ON CONFLICT DO NOTHING
			RETURNING event_id
		`, windowHours, afterHours, windowHours).Scan(&ids).Error
		if err != nil || len(ids) == 0 {
			return err
		}

		err = tx.
			Table("events").
			Select("id, name, starts_at, chat_id").
			Where("id IN ?", ids).
			Order("starts_at, id").
			Scan(&events).Error
		if err != nil {
			return err
		}

		members, err := (&BotRep{DB: tx}).GetEventMembers(ids)
		if err != nil {
			return err
		}
		for i := range events {
			events[i].Members = members[events[i].Id]
		}
		return nil
	})
	return events, err
}

// Группа стала супергруппой — у неё новый chat_id.
func (r *BotRep) MigrateChat(fromChatId, toChatId int64) error {
	return r.DB.Model(&domain.Event{}).Where("chat_id = ?", fromChatId).Update("chat_id", toChatId).Error
}
