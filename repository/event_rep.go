package repository

import (
	"rofl-bot/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventRep struct {
	DB *gorm.DB
}

func NewEventRep(db *gorm.DB) *EventRep {
	return &EventRep{
		DB: db,
	}
}

func (r *EventRep) CreateEvent(user_id uint, event *domain.Event) (uint, error) {
	err := r.DB.Create(event).Error
	if err != nil {
		return 0, err
	}
	return event.Id, r.DB.Create(&domain.EventMember{UserId: user_id, EventId: event.Id, RoleId: 1}).Error
}

func (r *EventRep) GetEvent(id uint) (*domain.Event, error) {
	var event domain.Event
	err := r.DB.Where("id = ?", id).First(&event).Error
	return &event, err
}

func (r *EventRep) GetEventWithCount(id uint) (*domain.EventWithMembersCount, error) {
	var event domain.EventWithMembersCount
	err := r.DB.
		Table("events").
		Select(`
			events.*,
			COUNT(em_all.user_id) AS count
		`).
		Joins(`
			LEFT JOIN event_members em_all 
			ON em_all.event_id = events.id
		`).
		Where("events.id = ?", id).
		Group("events.id").
		Scan(&event).Error
	return &event, err
}

func (r *EventRep) GetEventAdmin(id uint) (*domain.User, error) {
	var eventMember domain.EventMember
	err := r.DB.Preload("User").Where("event_id = ? and role_id = ?", id, 1).First(&eventMember).Error
	return &eventMember.User, err
}

func (r *EventRep) GetUserEvents(id uint) ([]domain.EventWithMembersCount, error) {
	var events []domain.EventWithMembersCount
	err := r.DB.
		Table("events").
		Select(`
			events.*,
			COUNT(DISTINCT em_all.user_id) AS count
		`).
		Joins(`
			JOIN event_members em_user 
			ON em_user.event_id = events.id 
			AND em_user.user_id = ?
		`, id).
		Joins(`
			LEFT JOIN event_members em_all 
			ON em_all.event_id = events.id
		`).
		Group("events.id").
		Scan(&events).Error

	if err != nil {
		return nil, err
	}

	return events, nil
}

func (r *EventRep) UpdateEvent(event *domain.Event) error {
	return r.DB.Save(event).Error
}

func (r *EventRep) CheckAdmin(eventId uint, userId uint) (bool, error) {
	var eventMember domain.EventMember
	err := r.DB.Preload("Role").Where("event_id = ? and user_id = ?", eventId, userId).First(&eventMember).Error
	if err != nil {
		return false, err
	}
	return eventMember.Role.Name == "admin", nil
}

func (r *EventRep) CheckMember(eventId uint, userId uint) (bool, error) {
	var eventMember domain.EventMember
	err := r.DB.Where("event_id = ? and user_id = ?", eventId, userId).First(&eventMember).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *EventRep) DeleteEvent(id uint) error {
	return r.DB.Where("id = ?", id).Delete(&domain.Event{}).Error
}

func (r *EventRep) LeaveEvent(eventId uint, userId uint) error {
	return r.DB.Where("event_id = ? and user_id = ?", eventId, userId).Delete(&domain.EventMember{}).Error
}

func (r *EventRep) JoinEvent(eventId uint, userId uint) error {
	return r.DB.Where("event_id = ? and user_id = ?", eventId, userId).FirstOrCreate(&domain.EventMember{UserId: userId, EventId: eventId, RoleId: 2}).Error
}

func (r *EventRep) GetEventMembers(eventId uint) ([]domain.EventMembers, error) {
	var eventMembers []domain.EventMembers
	err := r.DB.Table("event_members em").Joins("JOIN roles r ON r.id = em.role_id").Joins("JOIN users u ON u.id = em.user_id").Select("u.id   AS user_id, u.username, r.id   AS role_id, r.name AS role_name").Where("event_id = ?", eventId).Find(&eventMembers).Error
	return eventMembers, err
}

func (r *EventRep) CreateInvite(invite *domain.Invite) error {
	return r.DB.Create(invite).Error
}

func (r *EventRep) GetInvite(id uuid.UUID) (*domain.Invite, error) {
	var invite domain.Invite
	err := r.DB.Where("id = ?", id).First(&invite).Error
	return &invite, err
}

func (r *EventRep) GetInviteByEvent(eventId uint) (*domain.Invite, error) {
	var invite domain.Invite
	err := r.DB.Where("event_id = ?", eventId).First(&invite).Error
	return &invite, err
}

func (r *EventRep) DeleteInvite(id uuid.UUID) error {
	return r.DB.Where("id = ?", id).Delete(&domain.Invite{}).Error
}
