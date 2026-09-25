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
