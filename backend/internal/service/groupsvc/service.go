package groupsvc

import (
	"errors"
	"log"
	"strings"

	"sn-backend/internal/model"
	"sn-backend/internal/repository"
	ws "sn-backend/internal/websocket"
)

const (
	maxTitleLen       = 100
	maxDescriptionLen = 1000
)

var (
	ErrInvalidTitle       = errors.New("group: title is required (max 100 chars)")
	ErrInvalidDescription = errors.New("group: description is too long (max 1000 chars)")
	ErrNotFound           = errors.New("group: not found")
	ErrAlreadyMember      = errors.New("group: user is already a member")
	ErrInvitationExists   = errors.New("group: invitation already pending")
	ErrRequestExists      = errors.New("group: join request already pending")
	ErrNotGroupMember     = errors.New("group: only group members can do that")
	ErrNotGroupCreator    = errors.New("group: only the group creator can do that")
	ErrNotRecipient       = errors.New("group: user is not the invitation recipient")
	ErrSelfInvite         = errors.New("group: cannot invite yourself")
	ErrSelfRequest        = errors.New("group: cannot request to join your own group")
)

// Repository is the subset of repository.Repository the group service needs.
type Repository interface {
	CreateGroup(*model.Group) error
	GetGroup(int64) (*model.Group, error)
	ListGroups() ([]*model.Group, error)
	GetGroupMembers(int64) ([]*model.GroupMember, error)
	IsGroupMember(int64, int64) (bool, error)
	AddGroupMember(int64, int64) error
	CountGroupMembers(int64) (int, error)
	GetUserByID(int64) (*model.User, error)
	GetGroupCreator(int64) (*model.GroupCreator, error)

	CreateGroupInvitation(*model.GroupInvitation) (*model.GroupInvitation, error)
	GetGroupInvitationByID(int64) (*model.GroupInvitation, error)
	PendingGroupInvitation(int64, int64) (*model.GroupInvitation, error)
	GetPendingInvitationsForUser(int64) ([]*model.GroupInvitation, error)
	GetPendingInvitationsForGroup(int64) ([]*model.GroupInvitation, error)
	UpdateGroupInvitationStatus(int64, string) error

	CreateGroupJoinRequest(*model.GroupJoinRequest) (*model.GroupJoinRequest, error)
	GetGroupJoinRequestByID(int64) (*model.GroupJoinRequest, error)
	PendingGroupJoinRequest(int64, int64) (*model.GroupJoinRequest, error)
	GetPendingJoinRequestsForGroup(int64) ([]*model.GroupJoinRequest, error)
	UpdateGroupJoinRequestStatus(int64, string) error

	GroupDetailPayload(int64, int64) (*model.GroupDetail, error)
	GroupListPayload(int64) ([]*model.GroupListItem, error)

	AcceptGroupInvitationTx(int64, int64) error
	AcceptGroupJoinRequestTx(int64, int64) error
	RefuseGroupInvitationTx(int64, int64) error
	RefuseGroupJoinRequestTx(int64, int64) error

	CreateNotification(*model.Notification) error
}

type Service struct {
	repo Repository
	hub  *ws.Hub
}

// New wires the service to a repository and the websocket hub for
// notification fan-out; nil hub means notifications are only persisted.
func New(repo Repository, hub *ws.Hub) *Service {
	return &Service{repo: repo, hub: hub}
}

func (s *Service) Create(creatorID int64, title, description string) (*model.Group, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if title == "" || len(title) > maxTitleLen {
		return nil, ErrInvalidTitle
	}
	if len(description) > maxDescriptionLen {
		return nil, ErrInvalidDescription
	}

	group := &model.Group{CreatorID: creatorID, Title: title, Description: description}
	if err := s.repo.CreateGroup(group); err != nil {
		return nil, err
	}
	// The creator is a member from the start.
	if err := s.repo.AddGroupMember(group.ID, creatorID); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *Service) List(viewerID int64) ([]*model.GroupListItem, error) {
	return s.repo.GroupListPayload(viewerID)
}

// Detail returns one group for a viewer. Membership of the group gates the
// response: users who are neither members nor creators only see groups they
// could still join or were invited to, everything else is 404 to the client.
func (s *Service) Detail(viewerID, groupID int64) (*model.GroupDetail, error) {
	detail, err := s.repo.GroupDetailPayload(groupID, viewerID)
	if err != nil {
		return nil, err
	}
	if !detail.IsMember && !detail.IsCreator {
		pendingInvitation, err := s.repo.PendingGroupInvitation(groupID, viewerID)
		if err == nil {
			detail.PendingInvite = pendingInvitation.Status == model.GroupInvitationPending
		} else if !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}
		if !detail.PendingInvite && !detail.PendingJoin {
			return nil, ErrNotFound
		}
	}
	return detail, nil
}

// Members are visible to group members only.
func (s *Service) Members(viewerID, groupID int64) ([]*model.GroupMember, error) {
	member, err := s.repo.IsGroupMember(groupID, viewerID)
	if err != nil {
		return nil, err
	}
	if !member {
		return nil, ErrNotGroupMember
	}
	return s.repo.GetGroupMembers(groupID)
}

// Invite lets a current member invite an existing user. It rejects self
// invites, unknown users, existing members, duplicate pending invitations
// and stale/processed ones being reused.
func (s *Service) Invite(memberID, groupID, toUserID int64) (*model.GroupInvitation, error) {
	if memberID == toUserID {
		return nil, ErrSelfInvite
	}
	isMember, err := s.repo.IsGroupMember(groupID, memberID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotGroupMember
	}
	if _, err := s.repo.GetUserByID(toUserID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound // target user does not exist
		}
		return nil, err
	}
	targetIsMember, err := s.repo.IsGroupMember(groupID, toUserID)
	if err != nil {
		return nil, err
	}
	if targetIsMember {
		return nil, ErrAlreadyMember
	}

	existing, err := s.repo.PendingGroupInvitation(groupID, toUserID)
	switch {
	case err == nil:
		if existing.Status == model.GroupInvitationPending {
			return nil, ErrInvitationExists
		}
		// A processed invitation (accepted/declined) stays as history and a
		// fresh one is only possible if we could delete it; the schema keeps
		// one row per (group, user), so re-inviting after decline is not
		// possible. Surface it as a duplicate.
		return nil, ErrInvitationExists
	case errors.Is(err, repository.ErrNotFound):
		// no previous invitation for this (group, user) pair
	default:
		return nil, err
	}

	invitation := &model.GroupInvitation{
		GroupID:    groupID,
		FromUserID: memberID,
		ToUserID:   toUserID,
		Status:     model.GroupInvitationPending,
	}
	created, err := s.repo.CreateGroupInvitation(invitation)
	if err != nil {
		if isUnique(err) {
			return nil, ErrInvitationExists
		}
		return nil, err
	}

	group, err := s.repo.GetGroup(groupID)
	if err != nil {
		return nil, err
	}
	inviter, err := s.repo.GetUserByID(memberID)
	if err != nil {
		return nil, err
	}
	notification := &model.Notification{
		UserID:  toUserID,
		Type:    model.NotificationGroupInvite,
		ActorID: memberID,
		Content: inviter.Nickname + " invited you to join \"" + group.Title + "\"",
		GroupID: &group.ID,
	}
	s.notify(notification)
	return created, nil
}

// RespondInvitation accepts or refuses an invitation. Only the recipient of
// the pending invitation may respond to it.
func (s *Service) RespondInvitation(userID, invitationID int64, accept bool) error {
	if accept {
		if err := s.repo.AcceptGroupInvitationTx(invitationID, userID); err != nil {
			return err
		}
		invitation, err := s.repo.GetGroupInvitationByID(invitationID)
		if err != nil {
			return err
		}
		notification := &model.Notification{
			UserID:  invitation.FromUserID,
			Type:    model.NotificationGroupInviteResp,
			ActorID: userID,
			Content: "accepted your invitation to \"" + invitation.GroupTitle + "\"",
			GroupID: &invitation.GroupID,
		}
		s.notify(notification)
		return nil
	}
	return s.repo.RefuseGroupInvitationTx(invitationID, userID)
}

// RequestJoin lets a non-member ask to join a group. Creators do not request
// to join their own group and duplicates are rejected.
func (s *Service) RequestJoin(userID, groupID int64) (*model.GroupJoinRequest, error) {
	group, err := s.repo.GetGroup(groupID)
	if err != nil {
		return nil, err
	}
	if group.CreatorID == userID {
		return nil, ErrSelfRequest
	}
	isMember, err := s.repo.IsGroupMember(groupID, userID)
	if err != nil {
		return nil, err
	}
	if isMember {
		return nil, ErrAlreadyMember
	}

	existing, err := s.repo.PendingGroupJoinRequest(groupID, userID)
	switch {
	case err == nil:
		if existing.Status == model.GroupJoinPending {
			return nil, ErrRequestExists
		}
		// Same as invitations: one row per (group, user) means a processed
		// request blocks a new one.
		return nil, ErrRequestExists
	case errors.Is(err, repository.ErrNotFound):
		// no previous request for this (group, user) pair
	default:
		return nil, err
	}

	request := &model.GroupJoinRequest{
		GroupID: groupID,
		UserID:  userID,
		Status:  model.GroupJoinPending,
	}
	created, err := s.repo.CreateGroupJoinRequest(request)
	if err != nil {
		if isUnique(err) {
			return nil, ErrRequestExists
		}
		return nil, err
	}

	notification := &model.Notification{
		UserID:  group.CreatorID,
		Type:    model.NotificationGroupJoinReq,
		ActorID: userID,
		Content: requesterName(created) + " requested to join \"" + group.Title + "\"",
		GroupID: &group.ID,
	}
	s.notify(notification)
	return created, nil
}

// RespondJoinRequest accepts or refuses a join request. Only the group
// creator may respond to requests for their group.
func (s *Service) RespondJoinRequest(creatorID, requestID int64, accept bool) error {
	request, err := s.repo.GetGroupJoinRequestByID(requestID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	group, err := s.repo.GetGroup(request.GroupID)
	if err != nil {
		return err
	}
	if group.CreatorID != creatorID {
		return ErrNotGroupCreator
	}

	if !accept {
		return s.repo.RefuseGroupJoinRequestTx(requestID, request.GroupID)
	}

	if err := s.repo.AcceptGroupJoinRequestTx(requestID, request.GroupID); err != nil {
		return err
	}
	notification := &model.Notification{
		UserID:  request.UserID,
		Type:    model.NotificationGroupJoinResp,
		ActorID: creatorID,
		Content: "your request to join \"" + group.Title + "\" was accepted",
		GroupID: &group.ID,
	}
	s.notify(notification)
	return nil
}

func (s *Service) PendingInvitations(userID int64) ([]*model.GroupInvitation, error) {
	return s.repo.GetPendingInvitationsForUser(userID)
}

func (s *Service) PendingJoinRequests(viewerID, groupID int64) ([]*model.GroupJoinRequest, error) {
	group, err := s.repo.GetGroup(groupID)
	if err != nil {
		return nil, err
	}
	if group.CreatorID != viewerID {
		return nil, ErrNotGroupCreator
	}
	return s.repo.GetPendingJoinRequestsForGroup(groupID)
}

func requesterName(request *model.GroupJoinRequest) string {
	name := strings.TrimSpace(request.FirstName + " " + request.LastName)
	if name == "" {
		name = request.Nickname
	}
	return name
}

// notify persists a notification and best-effort pushes it to the user's
// websockets via the existing hub. Persistence failures are logged but do
// not fail the group operation.
func (s *Service) notify(n *model.Notification) {
	if err := s.repo.CreateNotification(n); err != nil {
		log.Printf("groupsvc: could not create notification: %v", err)
		return
	}
	if s.hub != nil {
		s.hub.PublishNotification(n)
	}
}

// isUnique reports SQLite UNIQUE-constraint violations from the driver so
// races between the pre-check and the insert stay safe.
func isUnique(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
