package repository

import (
	"rofl-bot/domain"

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

func (r *EventRep) GetEventAdmin(id uint) (*domain.User, error) {
	var eventMember domain.EventMember
	err := r.DB.Preload("User").Where("event_id = ? and role_id = ?", id, 1).First(&eventMember).Error
	return &eventMember.User, err
}

func (r *EventRep) GetUserEvents(id uint) ([]domain.Event, error) {
	var eventsMember []domain.EventMember
	err := r.DB.Preload("Event").Where("user_id = ?", id).Find(&eventsMember).Error
	if err != nil {
		return nil, err
	}

	events := make([]domain.Event, 0, len(eventsMember))
	for _, eventMember := range eventsMember {
		if eventMember.Event.Id != 0 {
			events = append(events, eventMember.Event)
		}
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
