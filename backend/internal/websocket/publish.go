package websocket

import "sn-backend/internal/model"

// PublishMessage pushes a private message to both sides of the conversation,
// in the same shape the chat page already listens for. It is used by the HTTP
// send endpoint, which is the path that can carry images.
func (h *Hub) PublishMessage(message *model.Message) {
	event := map[string]any{"type": "message", "message": message}
	if message.ToUserID != nil {
		h.publish(*message.ToUserID, event)
	}
	h.publish(message.FromUserID, event)
}

// PublishGroupMessage pushes a group message to every member of the group.
func (h *Hub) PublishGroupMessage(message *model.Message) {
	if message.GroupID == nil {
		return
	}
	members, err := h.repo.GroupMemberIDs(*message.GroupID)
	if err != nil {
		return
	}
	event := map[string]any{"type": "message", "message": message}
	for _, memberID := range members {
		h.publish(memberID, event)
	}
}
