package websocket

// relayTyping tells one user that somebody is writing to them.
//
// Typing is not saved anywhere: it only means "right now", so it is passed
// straight to the other person and forgotten. The recipient is checked with the
// same rule as a real message, so nobody can ping a user they may not write to.
func (h *Hub) relayTyping(fromUserID int64, toUserID *int64) {
	if toUserID == nil {
		return
	}
	allowed, err := h.repo.CanMessage(fromUserID, toUserID, nil)
	if err != nil || !allowed {
		return
	}
	h.publish(*toUserID, map[string]any{"type": "typing", "from_user_id": fromUserID})
}

// relayGroupTyping tells the other members of a group that somebody is writing.
// The sender is skipped, and only members are told.
func (h *Hub) relayGroupTyping(fromUserID int64, groupID *int64) {
	if groupID == nil {
		return
	}
	allowed, err := h.repo.CanMessage(fromUserID, nil, groupID)
	if err != nil || !allowed {
		return
	}
	members, err := h.repo.GroupMemberIDs(*groupID)
	if err != nil {
		return
	}
	for _, memberID := range members {
		if memberID == fromUserID {
			continue
		}
		h.publish(memberID, map[string]any{
			"type": "typing", "from_user_id": fromUserID, "group_id": *groupID,
		})
	}
}
