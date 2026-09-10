# Backend Implementation Guide

Owner: Person A (backend)
Depends on: nothing (start immediately)
Blocks: Frontend person needs API shapes from step 3+ to code against

---

## Step 1 — Fix session tracking bug

**File:** `backend/internal/websocket/hub.go`

The `trackClient` and `untrackClient` functions in `session.go` exist but are never called.
This means `RevokeSessionClients` (called on logout) does nothing.

### Changes to `hub.go`:

In `ServeHTTP`, after creating the client and before starting pumps:
```go
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request, sessions *sessionsvc.Service) {
    // ... existing cookie/session/upgrade code ...

    client := &Client{hub: h, connection: connection, userID: session.UserID, send: make(chan []byte, 16)}

    // ADD: get session ID from cookie and track the client
    sessionID := cookie.Value
    trackClient(sessionID, client)

    h.add(client)
    go client.writePump()
    client.readPump()
}
```

In `readPump` defer, add `untrackClient`:
```go
func (c *Client) readPump() {
    defer func() {
        untrackClient(c)   // ADD THIS
        c.hub.remove(c)
        c.connection.Close()
    }()
    // ... rest unchanged
}
```

### Verify:
- Login, connect WS, then logout → WS connection should close
- Connect from 2 tabs, logout from one → only that tab's WS closes

---

## Step 2 — Repository queries for messages

**File:** `backend/internal/repository/message.go`

Add these methods to the existing `Repository`:

### Conversations (private)
```go
type Conversation struct {
    User        *model.User     `json:"user"`
    LastMessage *model.Message  `json:"last_message"`
}

func (r *Repository) PrivateConversations(userID int64) ([]Conversation, error)
```

Query logic:
```sql
-- Get distinct conversation partners
WITH partners AS (
    SELECT DISTINCT
        CASE WHEN from_user_id = ? THEN to_user_id ELSE from_user_id END AS partner_id
    FROM messages
    WHERE (from_user_id = ? OR to_user_id = ?)
      AND to_user_id IS NOT NULL
)
SELECT p.partner_id,
       -- last message
       m.id, m.from_user_id, m.to_user_id, m.content, m.created_at
FROM partners p
JOIN messages m ON m.id = (
    SELECT id FROM messages
    WHERE (from_user_id = ? AND to_user_id = p.partner_id)
       OR (from_user_id = p.partner_id AND to_user_id = ?)
    ORDER BY created_at DESC LIMIT 1
)
ORDER BY m.created_at DESC
```

Then for each partner_id, fetch the user record.

### Group conversations
```go
type GroupConversation struct {
    Group       *model.Group    `json:"group"`
    LastMessage *model.Message  `json:"last_message"`
}

func (r *Repository) GroupConversations(userID int64) ([]GroupConversation, error)
```

Query logic:
```sql
WITH my_groups AS (
    SELECT group_id FROM group_members WHERE user_id = ?
)
SELECT g.id, g.title, g.description,
       m.id, m.from_user_id, m.group_id, m.content, m.created_at
FROM my_groups mg
JOIN groups g ON g.id = mg.group_id
LEFT JOIN messages m ON m.id = (
    SELECT id FROM messages WHERE group_id = g.id ORDER BY created_at DESC LIMIT 1
)
ORDER BY COALESCE(m.created_at, g.created_at) DESC
```

### Private message history
```go
func (r *Repository) MessagesBetween(userA, userB int64, before *int64, limit int) ([]model.Message, error)
```

```sql
SELECT id, from_user_id, to_user_id, group_id, content, created_at
FROM messages
WHERE ((from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?))
  AND (? IS NULL OR id < ?)
ORDER BY created_at DESC
LIMIT ?
```

### Group message history
```go
func (r *Repository) GroupMessages(groupID int64, before *int64, limit int) ([]model.Message, error)
```

```sql
SELECT id, from_user_id, to_user_id, group_id, content, created_at
FROM messages
WHERE group_id = ? AND (? IS NULL OR id < ?)
ORDER BY created_at DESC
LIMIT ?
```

### Message read state

There is no read tracking in the backend — no `read_at` column and no repository
methods for marking messages as read. Unread badges are handled client-side
(see `docs/FRONTEND.md`).

### Note on Message model
No changes needed to the `model.Message` struct — its current fields
(`ID`, `FromUserID`, `ToUserID`, `GroupID`, `Content`, `CreatedAt`) are sufficient.

---

## Step 3 — Message service

**New file:** `backend/internal/service/messagesvc/service.go`

```go
package messagesvc

type Service struct {
    repo *repository.Repository
}

func New(repo *repository.Repository) *Service

func (s *Service) ListConversations(userID int64) (private []repository.Conversation, groups []repository.GroupConversation, err error)
func (s *Service) GetPrivateMessages(userID, otherUserID int64, before *int64, limit int) ([]model.Message, bool, error)
func (s *Service) GetGroupMessages(userID, groupID int64, before *int64, limit int) ([]model.Message, bool, error)
```

`GetPrivateMessages` logic:
1. Call `CanMessage` to verify permission
2. Call `MessagesBetween` with limit+1 (to determine `has_more`)
3. Return messages (reversed for chronological order) and has_more

---

## Step 4 — Message handler

**New file:** `backend/internal/handler/messagehandler/handler.go`

```go
package messagehandler

type Handler struct {
    Service *messagesvc.Service
    Session *sessionsvc.Service
}

func New(service *messagesvc.Service, session *sessionsvc.Service) *Handler

func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request)
func (h *Handler) GetPrivateMessages(w http.ResponseWriter, r *http.Request)
func (h *Handler) GetGroupMessages(w http.ResponseWriter, r *http.Request)
```

Each handler:
1. Gets `userID` from session via `common.CurrentUserID(r, h.Session)`
2. Parses path/query params
3. Calls service method
4. Returns JSON via `common.WriteJSON`

---

## Step 5 — Notification repository, service, handler

### Repository (`backend/internal/repository/notification.go` — new file)

```go
type NotificationRow struct {
    ID        int64
    UserID    int64
    Type      string
    ActorID   int64
    ActorName string  // joined from users table
    Content   string
    GroupID   *int64
    Read      bool
    CreatedAt time.Time
}

func (r *Repository) ListNotifications(userID int64, before *int64, unreadOnly bool, limit int) ([]NotificationRow, bool, error)
func (r *Repository) UnreadNotificationCount(userID int64) (int, error)
func (r *Repository) MarkNotificationRead(userID, notificationID int64) error
func (r *Repository) MarkAllNotificationsRead(userID int64) error
func (r *Repository) CreateNotification(n *model.Notification) error
```

`ListNotifications` query:
```sql
SELECT n.id, n.user_id, n.type, n.actor_id, u.first_name || ' ' || u.last_name, n.content, n.group_id, n.read, n.created_at
FROM notifications n
JOIN users u ON u.id = n.actor_id
WHERE n.user_id = ? AND (? IS NULL OR n.id < ?) AND (? = 0 OR n.read = 0)
ORDER BY n.created_at DESC
LIMIT ?
```

### Service (`backend/internal/service/notificationsvc/service.go` — new file)

```go
package notificationsvc

type Service struct {
    repo *repository.Repository
    hub  *ws.Hub
}

func New(repo *repository.Repository, hub *ws.Hub) *Service

func (s *Service) List(userID int64, before *int64, unreadOnly bool) ([]repository.NotificationRow, bool, int, error)
func (s *Service) MarkRead(userID, notificationID int64) error
func (s *Service) MarkAllRead(userID int64) error
func (s *Service) CreateAndPublish(userID int64, notifType string, actorID int64, content string, groupID *int64) error
```

`CreateAndPublish` logic:
1. Insert into DB
2. Call `s.hub.PublishNotification(notification)` for real-time delivery

### Handler (`backend/internal/handler/notificationhandler/handler.go` — new file)

```go
package notificationhandler

type Handler struct {
    Service *notificationsvc.Service
    Session *sessionsvc.Service
}

func New(service *notificationsvc.Service, session *sessionsvc.Service) *Handler

func (h *Handler) List(w http.ResponseWriter, r *http.Request)
func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request)
func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request)
```

---

---

## Step 6 — Wire notifications into existing flows

### In `followsvc/service.go`
When a follow request is created for a private profile:
```go
// After creating the follow request
if targetUser.Private {
    notifSvc.CreateAndPublish(targetUserID, "follow_request", fromUserID, actorName+" wants to follow you", nil)
}
```

When a follow request is accepted:
```go
// After accepting
notifSvc.CreateAndPublish(fromUserID, "follow_accepted", targetUserID, actorName+" accepted your follow request", nil)
```

### In group flows (when group handler/service exists)
- Group invite → notify the invitee
- Join request → notify the group creator
- Event creation → notify all group members

Note: The notification service needs to be injected into followsvc and group services.
This means `handlers.go` and service constructors need updating.

---

## Step 7 — Register routes

**File:** `backend/internal/server/server.go`

Add to `RegisterRoutes`:
```go
// message routes
mux.Handle("GET /messages/conversations", auth.Authorized(http.HandlerFunc(h.Message.ListConversations)))
mux.Handle("GET /messages/{userId}", auth.Authorized(http.HandlerFunc(h.Message.GetPrivateMessages)))
mux.Handle("GET /messages/group/{groupId}", auth.Authorized(http.HandlerFunc(h.Message.GetGroupMessages)))

// notification routes
mux.Handle("GET /notifications", auth.Authorized(http.HandlerFunc(h.Notification.List)))
mux.Handle("POST /notifications/{id}/read", auth.Authorized(http.HandlerFunc(h.Notification.MarkRead)))
mux.Handle("POST /notifications/read-all", auth.Authorized(http.HandlerFunc(h.Notification.MarkAllRead)))
```

**File:** `backend/internal/handler/handlers.go`

Add to `Handlers` struct:
```go
type Handlers struct {
    Auth         *authhandler.Handler
    User         *userhandler.Handler
    Post         *posthandler.Handler
    File         *filehandler.Handler
    Message      *messagehandler.Handler      // ADD
    Notification *notificationhandler.Handler  // ADD
    WebSocket    *ws.Hub
}
```

Update `New()` to wire everything together.

---

## Step 8 — Update schema.sql reference

**File:** `backend/schema.sql`

No change needed — `read_at` was not added.

---

## File Checklist

| Action | File |
|--------|------|
| Modify | `backend/internal/websocket/hub.go` |
| Modify | `backend/internal/repository/message.go` |
| Modify | `backend/internal/handler/handlers.go` |
| Modify | `backend/internal/server/server.go` |
| Modify | `backend/internal/service/followsvc/service.go` (notification wiring) |
| New | `backend/internal/repository/notification.go` |
| New | `backend/internal/service/messagesvc/service.go` |
| New | `backend/internal/service/notificationsvc/service.go` |
| New | `backend/internal/handler/messagehandler/handler.go` |
| New | `backend/internal/handler/notificationhandler/handler.go` |
