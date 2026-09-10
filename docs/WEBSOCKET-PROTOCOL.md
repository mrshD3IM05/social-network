# WebSocket Protocol

Single shared endpoint: `GET /ws`

All messages are JSON. Every message has a `"type"` field.

---

## Client → Server

### Send a chat message

Private message (either user must follow the other, or recipient has public profile):

```json
{
  "type": "message",
  "to_user_id": 5,
  "content": "Hey, how are you?"
}
```

Group message (sender must be a member of the group):

```json
{
  "type": "message",
  "group_id": 12,
  "content": "Meeting at 3pm?"
}
```

Rules:
- Exactly one of `to_user_id` or `group_id` must be present (not both, not neither)
- `content` must be non-empty
- Server validates permissions via `CanMessage()` before persisting

---

## Server → Client

### New chat message

```json
{
  "type": "message",
  "message": {
    "id": 42,
    "from_user_id": 3,
    "to_user_id": 5,
    "group_id": null,
    "content": "Hey, how are you?",
    "created_at": "2026-09-07T14:30:00Z"
  }
}
```

Delivery:
- Private message → sent to both sender and recipient
- Group message → sent to all group members

### Notification

```json
{
  "type": "notification",
  "notification": {
    "id": 1,
    "type": "follow_request",
    "actor_id": 7,
    "actor_name": "Bob",
    "content": "Bob wants to follow you",
    "group_id": null,
    "read": false,
    "created_at": "2026-09-07T14:35:00Z"
  }
}
```

Notification types:
- `follow_request` — someone wants to follow you (private profile)
- `follow_accepted` — your follow request was accepted
- `group_invite` — someone invited you to a group
- `group_join_request` — someone wants to join your group
- `group_event` — new event in a group you belong to

### Error

```json
{
  "type": "error",
  "error": "message is not permitted"
}
```

---

## Connection Lifecycle

1. Frontend opens `ws(s)://host/api/v1/ws` with session cookie
2. Server authenticates via cookie, upgrades to WebSocket
3. Server calls `trackClient(sessionID, client)` for session revocation
4. On logout, server calls `RevokeSessionClients(sessionID)` which closes all connections for that session
5. On disconnect, server calls `untrackClient(client)` and removes from hub

## Reconnection

Frontend uses exponential backoff:
- Attempt 1: 1s delay
- Attempt 2: 2s delay
- Attempt 3: 4s delay
- Attempt 4: 8s delay
- Max: 30s delay
- Reset to 1s on successful connection
