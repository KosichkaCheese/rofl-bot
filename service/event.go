package service

import (
	"errors"
	"rofl-bot/domain"
	"rofl-bot/repository"

	"time"

	"github.com/google/uuid"
)

type EventService struct {
	rep *repository.EventRep
}

func NewEventService(rep *repository.EventRep) *EventService {
	return &EventService{rep: rep}
}

func (s *EventService) GetEvent(user_id uint, event_id uint) (*domain.Event, error) {
	isMember, err := s.rep.CheckMember(event_id, user_id)
	if err != nil && err.Error() != "record not found" {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("user is not member of this event")
	}
	return s.rep.GetEvent(event_id)
}

func (s *EventService) GetEventAdmin(event_id uint, user_id uint) (*domain.User, error) {
	isMember, err := s.rep.CheckMember(event_id, user_id)
	if err != nil && err.Error() != "record not found" {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("user is not member of this event")
	}
	return s.rep.GetEventAdmin(event_id)
}

func (s *EventService) CreateEvent(user_id uint, event *domain.CreateEvent) (uint, error) {
	new_event := &domain.Event{Name: event.Name}
	return s.rep.CreateEvent(user_id, new_event)
}

func (s *EventService) GetUserEvents(id uint) ([]domain.EventWithMembersCount, error) {
	return s.rep.GetUserEvents(id)
}

func (s *EventService) LeaveEvent(user_id, event_id uint) error {
	isMember, err := s.rep.CheckMember(event_id, user_id)
	if err != nil && err.Error() != "record not found" {
		return err
	}
	if !isMember {
		return errors.New("user is not member of this event")
	}
	return s.rep.LeaveEvent(event_id, user_id)
}

func (s *EventService) JoinEvent(user_id uint, invite_id uuid.UUID) error {

	invite, err := s.rep.GetInvite(invite_id)
	if err != nil {
		return err
	}

	isMember, err := s.rep.CheckMember(invite.EventId, user_id)
	if err != nil && err.Error() != "record not found" {
		return err
	}
	if isMember {
		return errors.New("user is already member of this event")
	}

	if time.Now().After(invite.ExpiresAt) {
		return errors.New("invite has expired")
	}

	return s.rep.JoinEvent(invite.EventId, user_id)
}

func (s *EventService) UpdateEvent(user_id uint, event *domain.UpdateEvent) error {
	isAdmin, err := s.rep.CheckAdmin(event.Id, user_id)
	if err != nil && err.Error() != "record not found" {
		return err
	}
	if !isAdmin {
		return errors.New("user is not admin of this event")
	}

	currEvent, err := s.rep.GetEvent(event.Id)
	if err != nil {
		return err
	}

	if event.Name != nil {
		currEvent.Name = *event.Name
	}
	if event.Ended != nil {
		currEvent.Ended = *event.Ended
	}

	return s.rep.UpdateEvent(currEvent)
}

func (s *EventService) DeleteEvent(user_id, event_id uint) error {
	isAdmin, err := s.rep.CheckAdmin(event_id, user_id)
	if err != nil && err.Error() != "record not found" {
		return err
	}
	if !isAdmin {
		return errors.New("user is not admin of this event")
	}

	return s.rep.DeleteEvent(event_id)
}

func (s *EventService) GetEventMembers(user_id, event_id uint) ([]domain.EventMembers, error) {
	isMember, err := s.rep.CheckMember(event_id, user_id)
	if err != nil && err.Error() != "record not found" {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("user is not member of this event")
	}

	return s.rep.GetEventMembers(event_id)
}

func (s *EventService) CreateInvite(user_id, event_id uint) (*domain.InviteLink, error) {
	isMember, err := s.rep.CheckMember(event_id, user_id)
	if err != nil && err.Error() != "record not found" {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("user is not member of this event")
	}

	currInvite, err := s.rep.GetInviteByEvent(event_id)
	if err != nil && err.Error() != "record not found" {
		return nil, err
	}
	if currInvite != nil {
		err = s.rep.DeleteInvite(currInvite.Id)
		if err != nil {
			return nil, err
		}
	}

	invite := &domain.Invite{
		Id:        uuid.New(),
		EventId:   event_id,
		ExpiresAt: time.Now().Add(time.Hour * 24),
	}

	err = s.rep.CreateInvite(invite)
	if err != nil {
		return nil, err
	}

	return &domain.InviteLink{
		Link:      "https://t.me/event_planner_vlrtnrt_bot?startapp=" + invite.Id.String(),
		ExpiresAt: invite.ExpiresAt,
	}, nil

}
