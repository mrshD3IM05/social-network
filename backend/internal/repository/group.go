package repository

import (
	"database/sql"
	"errors"
	"strings"

	"sn-backend/internal/model"
)

const groupColumns = `g.id, g.creator_id, g.title, g.description, g.created_at`

func scanGroup(s scanner) (*model.Group, error) {
	group := new(model.Group)
	if err := s.Scan(
		&group.ID,
		&group.CreatorID,
		&group.Title,
		&group.Description,
		&group.CreatedAt,
	); err != nil {
		return nil, err
	}
	return group, nil
}

func (r *Repository) CreateGroup(group *model.Group) error {
	if group == nil {
		return errors.New("group is nil")
	}

	result, err := r.db.Exec(
		`INSERT INTO groups (creator_id, title, description) VALUES (?, ?, ?)`,
		group.CreatorID,
		group.Title,
		group.Description,
	)
	if err != nil {
		return err
	}

	group.ID, err = result.LastInsertId()
	return err
}

func (r *Repository) GetGroup(id int64) (*model.Group, error) {
	group, err := scanGroup(r.QueryRow(
		`SELECT `+groupColumns+` FROM groups g WHERE g.id = ?`,
		id,
	))
	if err != nil {
		return nil, notFound(err)
	}
	return group, nil
}

func (r *Repository) ListGroups() ([]*model.Group, error) {
	rows, err := r.db.Query(`SELECT ` + groupColumns + ` FROM groups g ORDER BY g.created_at DESC, g.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]*model.Group, 0)
	for rows.Next() {
		group, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

func (r *Repository) UpdateGroup(group *model.Group) error {
	if group == nil {
		return errors.New("group is nil")
	}

	_, err := r.db.Exec(
		`UPDATE groups SET creator_id = ?, title = ?, description = ? WHERE id = ?`,
		group.CreatorID,
		group.Title,
		group.Description,
		group.ID,
	)
	return err
}

func (r *Repository) DeleteGroup(id int64) error {
	_, err := r.db.Exec("DELETE FROM groups WHERE id = ?", id)
	return err
}

// ---------------------------------------------------------------- members

func (r *Repository) AddGroupMember(groupID, userID int64) error {
	_, err := r.db.Exec(`INSERT INTO group_members (group_id, user_id) VALUES (?, ?)`, groupID, userID)
	return err
}

func (r *Repository) IsGroupMember(groupID, userID int64) (bool, error) {
	var exists int
	err := r.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = ? AND user_id = ?)`,
		groupID, userID,
	).Scan(&exists)
	return exists == 1, err
}

const groupMemberColumns = `gm.group_id, gm.user_id, u.first_name, u.last_name, u.nickname, u.avatar, gm.created_at`

func scanGroupMember(s scanner) (*model.GroupMember, error) {
	member := new(model.GroupMember)
	if err := s.Scan(
		&member.GroupID,
		&member.UserID,
		&member.FirstName,
		&member.LastName,
		&member.Nickname,
		&member.Avatar,
		&member.JoinedAt,
	); err != nil {
		return nil, err
	}
	return member, nil
}

func (r *Repository) GetGroupMembers(groupID int64) ([]*model.GroupMember, error) {
	rows, err := r.db.Query(
		`SELECT `+groupMemberColumns+`
		 FROM group_members gm
		 JOIN users u ON u.id = gm.user_id
		 WHERE gm.group_id = ?
		 ORDER BY gm.created_at, gm.user_id`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]*model.GroupMember, 0)
	for rows.Next() {
		member, err := scanGroupMember(rows)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, rows.Err()
}

// ----------------------------------------------------------- invitations

const groupInvitationColumns = `gi.id, gi.group_id, g.title, gi.from_user_id, gi.to_user_id, gi.status, gi.created_at`

func scanGroupInvitation(s scanner) (*model.GroupInvitation, error) {
	invitation := new(model.GroupInvitation)
	if err := s.Scan(
		&invitation.ID,
		&invitation.GroupID,
		&invitation.GroupTitle,
		&invitation.FromUserID,
		&invitation.ToUserID,
		&invitation.Status,
		&invitation.CreatedAt,
	); err != nil {
		return nil, err
	}
	return invitation, nil
}

func (r *Repository) CreateGroupInvitation(invitation *model.GroupInvitation) (*model.GroupInvitation, error) {
	result, err := r.db.Exec(
		`INSERT INTO group_invitations (group_id, from_user_id, to_user_id, status) VALUES (?, ?, ?, ?)`,
		invitation.GroupID, invitation.FromUserID, invitation.ToUserID, invitation.Status,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetGroupInvitationByID(id)
}

func (r *Repository) GetGroupInvitationByID(id int64) (*model.GroupInvitation, error) {
	invitation, err := scanGroupInvitation(r.QueryRow(
		`SELECT `+groupInvitationColumns+`
		 FROM group_invitations gi
		 JOIN groups g ON g.id = gi.group_id
		 WHERE gi.id = ?`,
		id,
	))
	if err != nil {
		return nil, notFound(err)
	}
	return invitation, nil
}

// PendingGroupInvitation returns the current invitation for (group, toUser)
// regardless of status, so the service can decide between reuse and reject.
func (r *Repository) PendingGroupInvitation(groupID, toUserID int64) (*model.GroupInvitation, error) {
	invitation, err := scanGroupInvitation(r.QueryRow(
		`SELECT `+groupInvitationColumns+`
		 FROM group_invitations gi
		 JOIN groups g ON g.id = gi.group_id
		 WHERE gi.group_id = ? AND gi.to_user_id = ?
		 ORDER BY gi.id DESC LIMIT 1`,
		groupID, toUserID,
	))
	if err != nil {
		return nil, notFound(err)
	}
	return invitation, nil
}

func (r *Repository) GetPendingInvitationsForUser(userID int64) ([]*model.GroupInvitation, error) {
	rows, err := r.db.Query(
		`SELECT `+groupInvitationColumns+`
		 FROM group_invitations gi
		 JOIN groups g ON g.id = gi.group_id
		 WHERE gi.to_user_id = ? AND gi.status = ?
		 ORDER BY gi.created_at DESC, gi.id DESC`,
		userID, model.GroupInvitationPending,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invitations := make([]*model.GroupInvitation, 0)
	for rows.Next() {
		invitation, err := scanGroupInvitation(rows)
		if err != nil {
			return nil, err
		}
		invitations = append(invitations, invitation)
	}
	return invitations, rows.Err()
}

func (r *Repository) GetPendingInvitationsForGroup(groupID int64) ([]*model.GroupInvitation, error) {
	rows, err := r.db.Query(
		`SELECT `+groupInvitationColumns+`
		 FROM group_invitations gi
		 JOIN groups g ON g.id = gi.group_id
		 WHERE gi.group_id = ? AND gi.status = ?
		 ORDER BY gi.created_at DESC, gi.id DESC`,
		groupID, model.GroupInvitationPending,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invitations := make([]*model.GroupInvitation, 0)
	for rows.Next() {
		invitation, err := scanGroupInvitation(rows)
		if err != nil {
			return nil, err
		}
		invitations = append(invitations, invitation)
	}
	return invitations, rows.Err()
}

func (r *Repository) UpdateGroupInvitationStatus(id int64, status string) error {
	_, err := r.db.Exec(`UPDATE group_invitations SET status = ? WHERE id = ?`, status, id)
	return err
}

// --------------------------------------------------------- join requests

const groupJoinRequestColumns = `gj.id, gj.group_id, g.title, gj.user_id, u.first_name, u.last_name, u.nickname, gj.status, gj.created_at`

func scanGroupJoinRequest(s scanner) (*model.GroupJoinRequest, error) {
	request := new(model.GroupJoinRequest)
	if err := s.Scan(
		&request.ID,
		&request.GroupID,
		&request.GroupTitle,
		&request.UserID,
		&request.FirstName,
		&request.LastName,
		&request.Nickname,
		&request.Status,
		&request.CreatedAt,
	); err != nil {
		return nil, err
	}
	return request, nil
}

func (r *Repository) CreateGroupJoinRequest(request *model.GroupJoinRequest) (*model.GroupJoinRequest, error) {
	result, err := r.db.Exec(
		`INSERT INTO group_join_requests (group_id, user_id, status) VALUES (?, ?, ?)`,
		request.GroupID, request.UserID, request.Status,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetGroupJoinRequestByID(id)
}

func (r *Repository) GetGroupJoinRequestByID(id int64) (*model.GroupJoinRequest, error) {
	request, err := scanGroupJoinRequest(r.QueryRow(
		`SELECT `+groupJoinRequestColumns+`
		 FROM group_join_requests gj
		 JOIN groups g ON g.id = gj.group_id
		 JOIN users u ON u.id = gj.user_id
		 WHERE gj.id = ?`,
		id,
	))
	if err != nil {
		return nil, notFound(err)
	}
	return request, nil
}

// PendingGroupJoinRequest returns the current request for (group, user)
// regardless of status, so the service can decide between reuse and reject.
func (r *Repository) PendingGroupJoinRequest(groupID, userID int64) (*model.GroupJoinRequest, error) {
	request, err := scanGroupJoinRequest(r.QueryRow(
		`SELECT `+groupJoinRequestColumns+`
		 FROM group_join_requests gj
		 JOIN groups g ON g.id = gj.group_id
		 JOIN users u ON u.id = gj.user_id
		 WHERE gj.group_id = ? AND gj.user_id = ?
		 ORDER BY gj.id DESC LIMIT 1`,
		groupID, userID,
	))
	if err != nil {
		return nil, notFound(err)
	}
	return request, nil
}

func (r *Repository) GetPendingJoinRequestsForGroup(groupID int64) ([]*model.GroupJoinRequest, error) {
	rows, err := r.db.Query(
		`SELECT `+groupJoinRequestColumns+`
		 FROM group_join_requests gj
		 JOIN groups g ON g.id = gj.group_id
		 JOIN users u ON u.id = gj.user_id
		 WHERE gj.group_id = ? AND gj.status = ?
		 ORDER BY gj.created_at DESC, gj.id DESC`,
		groupID, model.GroupJoinPending,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := make([]*model.GroupJoinRequest, 0)
	for rows.Next() {
		request, err := scanGroupJoinRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, rows.Err()
}

func (r *Repository) UpdateGroupJoinRequestStatus(id int64, status string) error {
	_, err := r.db.Exec(`UPDATE group_join_requests SET status = ? WHERE id = ?`, status, id)
	return err
}

// GetGroupCreator loads a user and maps it to the public creator shape so
// sensitive fields (password hash, email, date of birth) never reach clients.
func (r *Repository) GetGroupCreator(id int64) (*model.GroupCreator, error) {
	user, err := r.GetUserByID(id)
	if err != nil {
		return nil, err
	}
	return &model.GroupCreator{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Avatar:    user.Avatar,
		Nickname:  user.Nickname,
		AboutMe:   user.AboutMe,
		Private:   user.Private,
	}, nil
}

// ------------------------------------------------------------ aggregates

// ListGroupPostIDs returns the IDs of the posts that belong to a group.
func (r *Repository) ListGroupPostIDs(groupID int64) ([]int64, error) {
	rows, err := r.db.Query(`SELECT id FROM posts WHERE group_id = ?`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// IsGroupPost reports whether postID belongs to groupID.
func (r *Repository) IsGroupPost(groupID, postID int64) (bool, error) {
	var exists int
	err := r.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM posts WHERE id = ? AND group_id = ?)`,
		postID, groupID,
	).Scan(&exists)
	return exists == 1, err
}

// GetGroupIDForPost returns the group a post belongs to. ErrNotFound when the
// post does not exist or is not a group post.
func (r *Repository) GetGroupIDForPost(postID int64) (int64, error) {
	var groupID *int64
	if err := r.QueryRow(`SELECT group_id FROM posts WHERE id = ?`, postID).Scan(&groupID); err != nil {
		return 0, notFound(err)
	}
	if groupID == nil {
		return 0, ErrNotFound
	}
	return *groupID, nil
}

// IsGroupEvent reports whether eventID belongs to groupID.
func (r *Repository) IsGroupEvent(groupID, eventID int64) (bool, error) {
	var exists int
	err := r.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM group_events WHERE id = ? AND group_id = ?)`,
		eventID, groupID,
	).Scan(&exists)
	return exists == 1, err
}

// GetGroupIDForEvent returns the group an event belongs to.
func (r *Repository) GetGroupIDForEvent(eventID int64) (int64, error) {
	var groupID int64
	if err := r.QueryRow(`SELECT group_id FROM group_events WHERE id = ?`, eventID).Scan(&groupID); err != nil {
		return 0, notFound(err)
	}
	return groupID, nil
}

func (r *Repository) CountGroupMembers(groupID int64) (int, error) {
	var count int
	err := r.QueryRow(`SELECT COUNT(*) FROM group_members WHERE group_id = ?`, groupID).Scan(&count)
	return count, err
}

// GroupMemberships returns the group IDs userID belongs to (used to flag
// list/detail responses without N+1 queries).
func (r *Repository) GroupMemberships(userID int64) (map[int64]bool, error) {
	rows, err := r.db.Query(`SELECT group_id FROM group_members WHERE user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	memberships := make(map[int64]bool)
	for rows.Next() {
		var groupID int64
		if err := rows.Scan(&groupID); err != nil {
			return nil, err
		}
		memberships[groupID] = true
	}
	return memberships, rows.Err()
}

// PendingJoinRequestGroups returns the group IDs where userID has a pending
// join request.
func (r *Repository) PendingJoinRequestGroups(userID int64) (map[int64]bool, error) {
	rows, err := r.db.Query(
		`SELECT group_id FROM group_join_requests WHERE user_id = ? AND status = ?`,
		userID, model.GroupJoinPending,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pending := make(map[int64]bool)
	for rows.Next() {
		var groupID int64
		if err := rows.Scan(&groupID); err != nil {
			return nil, err
		}
		pending[groupID] = true
	}
	return pending, rows.Err()
}

// --------------------------------------------------- detail and browsing

func (r *Repository) GroupDetailPayload(groupID, viewerID int64) (*model.GroupDetail, error) {
	group, err := r.GetGroup(groupID)
	if err != nil {
		return nil, err
	}

	members, err := r.GetGroupMembers(groupID)
	if err != nil {
		return nil, err
	}

	detail := &model.GroupDetail{
		Group:   *group,
		Members: make([]model.GroupMember, 0, len(members)),
	}
	for _, member := range members {
		detail.Members = append(detail.Members, *member)
	}
	detail.MemberCount = len(detail.Members)

	detail.Creator, err = r.GetGroupCreator(group.CreatorID)
	if err != nil {
		return nil, err
	}

	detail.IsCreator = viewerID == group.CreatorID
	if !detail.IsCreator {
		detail.IsMember, err = r.IsGroupMember(groupID, viewerID)
		if err != nil {
			return nil, err
		}
		pending, err := r.PendingGroupJoinRequest(groupID, viewerID)
		if err == nil {
			detail.PendingJoin = pending.Status == model.GroupJoinPending
		} else if !errors.Is(err, ErrNotFound) {
			return nil, err
		}
	}

	return detail, nil
}

func (r *Repository) GroupListPayload(viewerID int64) ([]*model.GroupListItem, error) {
	groups, err := r.ListGroups()
	if err != nil {
		return nil, err
	}

	counts, err := r.GroupMemberCounts()
	if err != nil {
		return nil, err
	}
	memberships, err := r.GroupMemberships(viewerID)
	if err != nil {
		return nil, err
	}
	pendingJoins, err := r.PendingJoinRequestGroups(viewerID)
	if err != nil {
		return nil, err
	}

	items := make([]*model.GroupListItem, 0, len(groups))
	for _, group := range groups {
		item := &model.GroupListItem{
			Group:       *group,
			MemberCount: counts[group.ID],
			IsMember:    memberships[group.ID],
			PendingJoin: pendingJoins[group.ID],
			IsCreator:   group.CreatorID == viewerID,
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) GroupMemberCounts() (map[int64]int, error) {
	rows, err := r.db.Query(`SELECT group_id, COUNT(*) FROM group_members GROUP BY group_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[int64]int)
	for rows.Next() {
		var groupID int
		var count int
		if err := rows.Scan(&groupID, &count); err != nil {
			return nil, err
		}
		counts[int64(groupID)] = count
	}
	return counts, rows.Err()
}

// ----------------------------------------------------------- consistency

// AcceptGroupInvitationTx creates the membership and marks the invitation
// accepted atomically. It returns ErrNotFound when the invitation is missing,
// ErrExists when it is no longer pending or the user is already a member.
func (r *Repository) AcceptGroupInvitationTx(invitationID, userID int64) error {
	return r.withTx(func(tx *sql.Tx) error {
		var toUserID int64
		var status string
		err := tx.QueryRow(
			`SELECT to_user_id, status FROM group_invitations WHERE id = ?`,
			invitationID,
		).Scan(&toUserID, &status)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if toUserID != userID {
			return ErrNotOwner
		}
		if status != model.GroupInvitationPending {
			return ErrExists
		}

		var exists int
		if err := tx.QueryRow(
			`SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = (SELECT group_id FROM group_invitations WHERE id = ?) AND user_id = ?)`,
			invitationID, userID,
		).Scan(&exists); err != nil {
			return err
		}
		if exists == 1 {
			return ErrExists
		}

		if _, err := tx.Exec(
			`INSERT INTO group_members (group_id, user_id)
			 SELECT group_id, to_user_id FROM group_invitations WHERE id = ?`,
			invitationID,
		); err != nil {
			return err
		}

		if _, err := tx.Exec(
			`UPDATE group_invitations SET status = ? WHERE id = ?`,
			model.GroupInvitationAccept, invitationID,
		); err != nil {
			return err
		}
		return nil
	})
}

// AcceptGroupJoinRequestTx creates the membership and marks the join request
// accepted atomically. It returns ErrNotFound when the request is missing,
// ErrExists when it is no longer pending or the user is already a member.
func (r *Repository) AcceptGroupJoinRequestTx(requestID, groupID int64) error {
	return r.withTx(func(tx *sql.Tx) error {
		var reqGroupID int64
		var reqUserID int64
		var status string
		err := tx.QueryRow(
			`SELECT group_id, user_id, status FROM group_join_requests WHERE id = ?`,
			requestID,
		).Scan(&reqGroupID, &reqUserID, &status)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if reqGroupID != groupID {
			return ErrNotFound
		}
		if status != model.GroupJoinPending {
			return ErrExists
		}

		var exists int
		if err := tx.QueryRow(
			`SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = ? AND user_id = ?)`,
			reqGroupID, reqUserID,
		).Scan(&exists); err != nil {
			return err
		}
		if exists == 1 {
			return ErrExists
		}

		if _, err := tx.Exec(
			`INSERT INTO group_members (group_id, user_id) VALUES (?, ?)`,
			reqGroupID, reqUserID,
		); err != nil {
			return err
		}

		if _, err := tx.Exec(
			`UPDATE group_join_requests SET status = ? WHERE id = ?`,
			model.GroupJoinAccept, requestID,
		); err != nil {
			return err
		}
		return nil
	})
}

// RefuseGroupInvitationTx marks the invitation declined only when it is still
// pending and belongs to userID. Returns ErrNotFound / ErrNotOwner / ErrExists.
func (r *Repository) RefuseGroupInvitationTx(invitationID, userID int64) error {
	return r.withTx(func(tx *sql.Tx) error {
		var toUserID int64
		var status string
		err := tx.QueryRow(
			`SELECT to_user_id, status FROM group_invitations WHERE id = ?`,
			invitationID,
		).Scan(&toUserID, &status)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if toUserID != userID {
			return ErrNotOwner
		}
		if status != model.GroupInvitationPending {
			return ErrExists
		}
		if _, err := tx.Exec(
			`UPDATE group_invitations SET status = ? WHERE id = ?`,
			model.GroupInvitationDecline, invitationID,
		); err != nil {
			return err
		}
		return nil
	})
}

// RefuseGroupJoinRequestTx marks the join request declined only when it is
// still pending and belongs to groupID. Returns ErrNotFound / ErrExists.
func (r *Repository) RefuseGroupJoinRequestTx(requestID, groupID int64) error {
	return r.withTx(func(tx *sql.Tx) error {
		var reqGroupID int64
		var status string
		err := tx.QueryRow(
			`SELECT group_id, status FROM group_join_requests WHERE id = ?`,
			requestID,
		).Scan(&reqGroupID, &status)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if reqGroupID != groupID {
			return ErrNotFound
		}
		if status != model.GroupJoinPending {
			return ErrExists
		}
		if _, err := tx.Exec(
			`UPDATE group_join_requests SET status = ? WHERE id = ?`,
			model.GroupJoinDecline, requestID,
		); err != nil {
			return err
		}
		return nil
	})
}

// withTx runs fn inside a transaction, rolling back on error.
func (r *Repository) withTx(fn func(tx *sql.Tx) error) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// isUniqueConstraint reports whether err is a SQLite UNIQUE violation, used
// to translate duplicate inserts into ErrExists for the service layer.
func isUniqueConstraint(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
