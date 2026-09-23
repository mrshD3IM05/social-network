package eventsvc

import (
	"errors"
	"log"
	"strings"
	"time"

	"sn-backend/internal/model"
	"sn-backend/internal/repository"
	ws "sn-backend/internal/websocket"
)

const (
	maxTitleLen       = 100
	maxDescriptionLen = 1000
)

var (
	ErrInvalidTitle       = errors.New("event: title is required (max 100 chars)")
	ErrInvalidDescription = errors.New("event: description is too long (max 1000 chars)")
	ErrInvalidDateTime    = errors.New("event: a valid future date and time is required")
	ErrNotFound           = errors.New("event: not found")
	ErrInvalidChoice      = errors.New("event: choice must be going or not_going")
	ErrNotGroupMember     = errors.New("event: only group members can do that")
)

// Repository is the subset of repository.Repository the event service needs.
type Repository interface {
	CreateEvent(*model.GroupEvent) error
	ListGroupEvents(groupID, viewerID int64) ([]*model.EventListItem, error)
	SetEventResponse(int64, int64, string) error
	GetEventResponse(int64, int64) (*model.EventResponse, error)
	EventResponseCounts(int64) (int, int, error)

	GetGroup(int64) (*model.Group, error)
	GetGroupIDForEvent(int64) (int64, error)
	IsGroupMember(int64, int64) (bool, error)
	GroupMemberIDs(int64) ([]int64, error)
	GetUserByID(int64) (*model.User, error)

	CreateNotification(*model.Notification) error
}

// Service reuses the group notification plumbing (persist + hub fan-out).
type Service struct {
	repo Repository
	hub  *ws.Hub
}

func New(repo Repository, hub *ws.Hub) *Service {
	return &Service{repo: repo, hub: hub}
}

// Create adds an event to a group. Every member may create events: the
// subject gives no special event role, and this matches how posting in the
// group works.
func (s *Service) Create(creatorID, groupID int64, title, description string, dateTime time.Time) (*model.GroupEvent, error) {
	if _, err := s.repo.GetGroup(groupID); err != nil {
		return nil, err // repository.ErrNotFound → 404
	}
	member, err := s.repo.IsGroupMember(groupID, creatorID)
	if err != nil {
		return nil, err
	}
	if !member {
		return nil, ErrNotGroupMember
	}

	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if title == "" || len(title) > maxTitleLen {
		return nil, ErrInvalidTitle
	}
	if len(description) > maxDescriptionLen {
		return nil, ErrInvalidDescription
	}
	if dateTime.IsZero() || dateTime.Before(time.Now().Add(-time.Minute)) {
		return nil, ErrInvalidDateTime
	}

	event := &model.GroupEvent{
		GroupID:     groupID,
		CreatorID:   creatorID,
		Title:       title,
		Description: description,
		DateTime:    dateTime,
	}
	if err := s.repo.CreateEvent(event); err != nil {
		return nil, err
	}

	s.notifyMembers(event)
	return event, nil
}

// List returns the events of a group with counts and the viewer's choice.
// Members only.
func (s *Service) List(viewerID, groupID int64) ([]*model.EventListItem, error) {
	member, err := s.repo.IsGroupMember(groupID, viewerID)
	if err != nil {
		return nil, err
	}
	if !member {
		return nil, ErrNotGroupMember
	}
	return s.repo.ListGroupEvents(groupID, viewerID)
}

// Respond sets or changes the viewer's going / not-going answer. One row per
// (event, user) is guaranteed by the event_responses UNIQUE constraint and
// the upsert, so changing the answer replaces the old one.
func (s *Service) Respond(viewerID, eventID int64, choice string) (going, notGoing int, err error) {
	if choice != model.EventChoiceGoing && choice != model.EventChoiceNotGoing {
		return 0, 0, ErrInvalidChoice
	}
	groupID, err := s.repo.GetGroupIDForEvent(eventID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return 0, 0, ErrNotFound
		}
		return 0, 0, err
	}
	member, err := s.repo.IsGroupMember(groupID, viewerID)
	if err != nil {
		return 0, 0, err
	}
	if !member {
		return 0, 0, ErrNotGroupMember
	}

	if err := s.repo.SetEventResponse(eventID, viewerID, choice); err != nil {
		return 0, 0, err
	}
	going, notGoing, err = s.repo.EventResponseCounts(eventID)
	return going, notGoing, err
}

// MyResponse returns the viewer's current answer for an event.
func (s *Service) MyResponse(viewerID, eventID int64) (*model.EventResponse, error) {
	groupID, err := s.repo.GetGroupIDForEvent(eventID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	member, err := s.repo.IsGroupMember(groupID, viewerID)
	if err != nil {
		return nil, err
	}
	if !member {
		return nil, ErrNotGroupMember
	}
	response, err := s.repo.GetEventResponse(eventID, viewerID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, nil // member who has not answered yet
	}
	return response, err
}

// notifyMembers tells every member except the creator that a new event was
// scheduled (subject requirement). Non-members never receive it because the
// recipients come straight from group_members.
func (s *Service) notifyMembers(event *model.GroupEvent) {
	memberIDs, err := s.repo.GroupMemberIDs(event.GroupID)
	if err != nil {
		log.Printf("eventsvc: could not list members for notification: %v", err)
		return
	}
	group, err := s.repo.GetGroup(event.GroupID)
	if err != nil {
		return
	}
	creator, err := s.repo.GetUserByID(event.CreatorID)
	if err != nil {
		return
	}
	name := strings.TrimSpace(creator.FirstName + " " + creator.LastName)
	if name == "" {
		name = creator.Nickname
	}
	for _, memberID := range memberIDs {
		if memberID == event.CreatorID {
			continue
		}
		s.notify(&model.Notification{
			UserID:  memberID,
			Type:    model.NotificationEventCreated,
			ActorID: event.CreatorID,
			Content: name + " created the event \"" + event.Title + "\" in \"" + group.Title + "\"",
			GroupID: &event.GroupID,
		})
	}
}

// notify persists a notification and pushes it to the user's websockets via
// the existing hub (same contract as groupsvc.notify).
func (s *Service) notify(n *model.Notification) {
	if err := s.repo.CreateNotification(n); err != nil {
		log.Printf("eventsvc: could not create notification: %v", err)
		return
	}
	if s.hub != nil {
		s.hub.PublishNotification(n)
	}
}
