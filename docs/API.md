# API Contract — HTTP Endpoints

All endpoints require authentication (session cookie) unless noted.
Base path: `/api/v1`

---

## Chat Endpoints

### List conversations

```
GET /messages/conversations
```

Response `200`:
```json
{
  "private": [
    {
      "user": { "id": 5, "first_name": "Alice", "last_name": "Smith", "avatar": "abc123" },
      "last_message": {
        "id": 42,
        "from_user_id": 3,
        "to_user_id": 5,
        "content": "See you tomorrow!",
        "created_at": "2026-09-07T14:30:00Z"
      }
    }
  ],
  "groups": [
    {
      "group": { "id": 12, "title": "Dev Team", "description": "..." },
      "last_message": {
        "id": 99,
        "from_user_id": 7,
        "group_id": 12,
        "content": "Deployed!",
        "created_at": "2026-09-07T15:00:00Z"
      }
    }
  ]
}
```

Logic:
- Private: find all users where a message exists between current user and that user
- Groups: find all groups where current user is a member
- `last_message`: most recent message in that conversation
- No read/unread state is tracked for messages

---

### Get private message history

```
GET /messages/{userId}?before=42&limit=50
```

Path params:
- `userId` — the other user's ID

Query params:
- `before` (optional) — message ID to paginate before (for infinite scroll upward)
- `limit` (optional, default 50, max 100) — number of messages

Response `200`:
```json
{
  "messages": [
    {
      "id": 40,
      "from_user_id": 3,
      "to_user_id": 5,
      "content": "Hello!",
      "created_at": "2026-09-07T13:55:00Z"
    },
    {
      "id": 41,
      "from_user_id": 5,
      "to_user_id": 3,
      "content": "Hi there!",
      "created_at": "2026-09-07T14:00:00Z"
    }
  ],
  "has_more": true
}
```

Logic:
- Permission: either user must follow the other, or recipient has public profile
- Returns messages WHERE (from=A AND to=B) OR (from=B AND to=A)
- Ordered by `created_at DESC`, then reversed for display
- If `before` is provided, return messages with id < before

---

### Get group message history

```
GET /messages/group/{groupId}?before=42&limit=50
```

Path params:
- `groupId` — the group ID

Query params:
- `before` (optional) — message ID to paginate before
- `limit` (optional, default 50, max 100)

Response `200`:
```json
{
  "messages": [
    {
      "id": 50,
      "from_user_id": 7,
      "group_id": 12,
      "content": "Deployed!",
      "created_at": "2026-09-07T15:00:00Z"
    }
  ],
  "has_more": false
}
```

Logic:
- Permission: user must be a member of the group
- Returns messages WHERE group_id = groupId
- Ordered by `created_at DESC`, reversed for display

---

## Notification Endpoints

### List notifications

```
GET /notifications?before=10&unread=true
```

Query params:
- `before` (optional) — notification ID to paginate before
- `unread` (optional, boolean) — filter to only unread notifications

Response `200`:
```json
{
  "notifications": [
    {
      "id": 10,
      "type": "follow_request",
      "actor_id": 7,
      "actor_name": "Bob",
      "content": "Bob wants to follow you",
      "group_id": null,
      "read": false,
      "created_at": "2026-09-07T14:35:00Z"
    }
  ],
  "has_more": true,
  "unread_count": 3
}
```

Logic:
- Fetches from `notifications` table WHERE user_id = current user
- Join with `users` table on `actor_id` to get `actor_name`
- If `before` provided, WHERE id < before
- If `unread=true`, WHERE read = 0
- `unread_count` is total unread regardless of pagination

---

### Mark notification as read

```
POST /notifications/{id}/read
```

Response `200`:
```json
{ "status": "ok" }
```

Logic:
- UPDATE notifications SET read = 1 WHERE id = ? AND user_id = current_user

---

### Mark all notifications as read

```
POST /notifications/read-all
```

Response `200`:
```json
{ "status": "ok" }
```

Logic:
- UPDATE notifications SET read = 1 WHERE user_id = current_user AND read = 0

---

## Existing Endpoints (unchanged, for reference)

| Method | Path | Purpose |
|--------|------|---------|
| POST | /register | Create account |
| POST | /login | Sign in |
| POST | /logout | Sign out (revokes WS connections) |
| GET | /me | Current user info |
| GET | /user/{id} | Get user profile |
| POST | /users/{id}/follow | Send follow request |
| DELETE | /users/{id}/follow | Unfollow |
| POST | /follow-requests/{id}/accept | Accept follow request |
| POST | /follow-requests/{id}/decline | Decline follow request |
| GET | /posts | List posts |
| POST | /posts | Create post |
| PUT | /posts/{id} | Update post |
| DELETE | /posts/{id} | Delete post |
| POST | /files | Upload image(s) — multipart `files[]` (max 3, jpeg/png/gif), optional `post_id` **or** `message_id` to attach |
| POST /avatar | Set avatar |
| GET | /fs/{id} | Download file (private cacheable) |
| GET | /ws | WebSocket connection |
