package followsvc

import (
	"fmt"

	"sn-backend/internal/model"
)

const (
	NotifyFollowRequest  = "follow_request"
	NotifyFollowAccepted = "follow_accepted"
)

// Notifier persists a notification and pushes it to the recipient's open
// sockets. The websocket hub plus the repository satisfy it together.
type Notifier interface {
	CreateNotification(*model.Notification) error
	PublishNotification(*model.Notification)
}

// WithNotifier returns the same service wired to emit follow notifications.
func (s *Service) WithNotifier(notifier Notifier) *Service {
	s.notifier = notifier
	return s
}

func (s *Service) notify(userID, actorID int64, kind, content string) {
	if s.notifier == nil {
		return
	}
	notification := &model.Notification{UserID: userID, Type: kind, ActorID: actorID, Content: content}
	if err := s.notifier.CreateNotification(notification); err != nil {
		return
	}
	s.notifier.PublishNotification(notification)
}

// notifyFollowRequest tells the target that someone asked to follow them.
func (s *Service) notifyFollowRequest(follow *model.FollowRequest) {
	actor, err := s.repo.GetUserByID(follow.FromUserID)
	if err != nil {
		return
	}
	s.notify(follow.ToUserID, follow.FromUserID, NotifyFollowRequest,
		fmt.Sprintf("%s wants to follow you", displayName(actor)))
}

// notifyFollowAccepted tells the requester that their request went through.
func (s *Service) notifyFollowAccepted(follow *model.FollowRequest) {
	actor, err := s.repo.GetUserByID(follow.ToUserID)
	if err != nil {
		return
	}
	s.notify(follow.FromUserID, follow.ToUserID, NotifyFollowAccepted,
		fmt.Sprintf("%s accepted your follow request", displayName(actor)))
}

func displayName(user *model.User) string {
	name := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	if name == " " {
		return user.Nickname
	}
	return name
}
