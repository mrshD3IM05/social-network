# Team Roles — Work Split for 4 People

## Overview

Each person owns a domain. Dependencies flow downward — top rows should start first.

---

## Person 1 — Backend: Groups + Events

Owns all group-related backend logic. Depends on nothing (can start immediately).

### Tasks

1. **Repository** (`internal/repository/group.go` — extend existing)
   - Group CRUD (create, get, list all)
   - Group membership (add member, remove member, list members)
   - Group invitations (create, accept, decline, list pending)
   - Group join requests (create, accept, decline, list pending)
   - Group posts (create, list by group with visibility)
   - Group events (create, list by group)
   - Event responses (create/update RSVP, list responses)

2. **Service** (`internal/service/groupsvc/service.go` — new)
   - `CreateGroup(creatorID, title, description)` — create group, auto-add creator as member
   - `InviteUser(groupID, fromUserID, toUserID)` — create invitation
   - `RespondToInvite(invitationID, userID, status)` — accept/decline
   - `RequestJoin(groupID, userID)` — create join request (if not private profile bypass)
   - `RespondToJoinRequest(requestID, creatorID, status)` — only group creator can accept/decline
   - `AddMember(groupID, userID)` — add to group_members
   - `RemoveMember(groupID, userID)` — remove from group_members
   - `ListGroups()` — list all groups for browse page
   - `GetGroup(groupID)` — get group info
   - `CreateEvent(groupID, creatorID, title, description, dateTime)` — create event
   - `RespondToEvent(eventID, userID, choice)` — "going" or "not_going"
   - `GetGroupPosts(groupID, userID)` — posts visible to group members
   - `IsMember(groupID, userID)` — check membership

3. **Handler** (`internal/handler/grouphandler/handler.go` — new)
   - `POST /groups` — create group
   - `GET /groups` — list all groups
   - `GET /groups/{id}` — get group detail
   - `POST /groups/{id}/invite` — invite user
   - `POST /groups/{id}/join` — request to join
   - `POST /groups/invitations/{id}/accept` — accept invite
   - `POST /groups/invitations/{id}/decline` — decline invite
   - `POST /groups/join-requests/{id}/accept` — accept join (creator only)
   - `POST /groups/join-requests/{id}/decline` — decline join (creator only)
   - `GET /groups/{id}/members` — list members
   - `POST /groups/{id}/events` — create event
   - `GET /groups/{id}/events` — list events
   - `POST /events/{id}/respond` — RSVP to event
   - `GET /groups/{id}/posts` — list group posts

4. **Route registration** — add group routes to `internal/server/server.go`

### Files to create/modify

| Action | File |
|--------|------|
| Modify | `internal/repository/group.go` |
| New | `internal/service/groupsvc/service.go` |
| New | `internal/handler/grouphandler/handler.go` |
| Modify | `internal/handler/handlers.go` — add Group handler |
| Modify | `internal/server/server.go` — register group routes |

### Notification triggers (coordinate with Person 2)

- After creating a group invitation → call notification service to notify invitee
- After creating a join request → call notification service to notify group creator
- After creating an event → call notification service to notify all group members

---

## Person 2 — Backend: Notifications + Comments + Reactions + WS Extension

Owns notification system, comments, reactions, and WebSocket extensions. Depends on nothing (can start immediately).

### Tasks

1. **Notification repository** (`internal/repository/notification.go` — new)
   - `ListNotifications(userID, before, unreadOnly, limit)` — paginated list with actor name join
   - `UnreadNotificationCount(userID)` — count unread
   - `MarkNotificationRead(userID, notificationID)` — mark one as read
   - `MarkAllNotificationsRead(userID)` — mark all as read
   - `CreateNotification(notification)` — insert

2. **Notification service** (`internal/service/notificationsvc/service.go` — new)
   - `List(userID, before, unreadOnly)` — returns notifications + has_more + unread_count
   - `MarkRead(userID, notificationID)`
   - `MarkAllRead(userID)`
   - `CreateAndPublish(userID, type, actorID, content, groupID)` — insert + publish via Hub
   - Needs `*ws.Hub` injected to call `hub.PublishNotification()`

3. **Notification handler** (`internal/handler/notificationhandler/handler.go` — new)
   - `GET /notifications` — list with ?before= and ?unread=true params
   - `POST /notifications/{id}/read` — mark one read
   - `POST /notifications/read-all` — mark all read

4. **Message repository extensions** (`internal/repository/message.go` — extend)
   - `PrivateConversations(userID)` — list conversation partners with last message
   - `GroupConversations(userID)` — list group chats with last message
   - `MessagesBetween(userA, userB, before, limit)` — paginated private history
   - `GroupMessages(groupID, before, limit)` — paginated group history

5. **Message service** (`internal/service/messagesvc/service.go` — new)
   - `ListConversations(userID)` — returns private + group conversations
   - `GetPrivateMessages(userID, otherUserID, before, limit)` — fetch history
   - `GetGroupMessages(userID, groupID, before, limit)` — fetch history

6. **Message handler** (`internal/handler/messagehandler/handler.go` — new)
   - `GET /messages/conversations` — list all conversations
   - `GET /messages/{userId}` — private message history with ?before= and ?limit=
   - `GET /messages/group/{groupId}` — group message history

7. **WebSocket hub extension** (`internal/websocket/hub.go` — extend)
   - Extend `Repository` interface: add `GetUserByID`
   - Add `"typing"` handler in readPump: validate permissions, broadcast to recipient/group (no DB)

8. **Session tracking fix** (`internal/websocket/hub.go` — fix)
   - In `ServeHTTP`: call `trackClient(cookie.Value, client)` after creating client
   - In `readPump` defer: call `untrackClient(c)` before `remove`

9. **Route registration** — add message + notification routes to `internal/server/server.go`

10. **Wire notifications into follow flow** (`internal/service/followsvc/service.go` — modify)
    - When follow request created for private profile → `notifSvc.CreateAndPublish(targetUserID, "follow_request", ...)`
    - When follow request accepted → `notifSvc.CreateAndPublish(fromUserID, "follow_accepted", ...)`

### Files to create/modify

| Action | File |
|--------|------|
| New | `internal/repository/notification.go` |
| Modify | `internal/repository/message.go` |
| New | `internal/service/notificationsvc/service.go` |
| New | `internal/service/messagesvc/service.go` |
| New | `internal/handler/notificationhandler/handler.go` |
| New | `internal/handler/messagehandler/handler.go` |
| Modify | `internal/websocket/hub.go` |
| Modify | `internal/handler/handlers.go` |
| Modify | `internal/server/server.go` |
| Modify | `internal/service/followsvc/service.go` |

---

## Person 3 — Frontend: Pages + Components + Routing

Owns all page-level UI. Can start with mock data, connect to real APIs once Person 1 & 2 define contracts (see `docs/API.md`).

### Tasks

1. **App providers** (`src/app/providers/` — new folder)
   - `WebSocketProvider.js` — single shared WS connection, reconnect logic, send/subscribe API
   - `UnreadProvider.js` — client-side per-conversation unread tracking + notification counts, listens to WS events
   - `AppProviders.js` — client wrapper combining both providers (needed if layout.js is server component)

2. **Conversation list** (`src/app/components/ChatLayout.js` — new)
   - Fetches `GET /messages/conversations`
   - Sidebar with search, conversation items (avatar, name, last message preview, timestamp, unread badge)
   - Unread badges tracked client-side (per conversation, reset on open — no backend read tracking)
   - Real-time: new messages move conversations to top

3. **Message thread** (`src/app/components/MessageThread.js` — new)
   - Fetches `GET /messages/{userId}` or `GET /messages/group/{groupId}`
   - Message bubbles (sent vs received), timestamps
   - Typing indicator, auto-scroll, infinite scroll upward for older messages

4. **Emoji picker** (`src/app/components/EmojiPicker.js` — new)
   - Grid of common emojis in categories, search input, click to insert
   - Popover positioned above the chat input

5. **Rewrite messages page** (`src/app/messages/page.js` — rewrite)
   - Replace user-ID-input chat with `ChatLayout` + `MessageThread`
   - Layout: sidebar | main thread

6. **Rewrite notifications page** (`src/app/notifications/page.js` — rewrite)
   - Fetch history from `GET /notifications`
   - Real-time via WS, accept/decline buttons for follow requests + group invites
   - Mark all as read, unread visual distinction

7. **Update SocialShell** (`src/app/components/SocialShell.js` — modify)
   - Add notification bell icon with unread count badge in topbar
   - Add messages icon with unread count badge in nav

8. **Update layout** (`src/app/layout.js` — modify)
   - Wrap children with `AppProviders` (WebSocketProvider + UnreadProvider)

9. **CSS additions** (`src/app/page.module.css` — extend)
   - Chat layout grid, conversation sidebar, message bubbles, typing dots animation
   - Emoji picker grid, notification cards, unread badges, responsive breakpoints

### Files to create/modify

| Action | File |
|--------|------|
| New | `src/app/providers/WebSocketProvider.js` |
| New | `src/app/providers/UnreadProvider.js` |
| New | `src/app/providers/AppProviders.js` |
| New | `src/app/components/ChatLayout.js` |
| New | `src/app/components/MessageThread.js` |
| New | `src/app/components/EmojiPicker.js` |
| Rewrite | `src/app/messages/page.js` |
| Rewrite | `src/app/notifications/page.js` |
| Modify | `src/app/components/SocialShell.js` |
| Modify | `src/app/layout.js` |
| Modify | `src/app/page.module.css` |

---

## Person 4 — Integration + DevOps + Chat WebSocket Client

Owns infrastructure, shared frontend utilities, and end-to-end integration testing. Can start immediately on Docker/DevOps, then moves to frontend integration once Person 3 has providers ready.

### Tasks

1. **Frontend Dockerfile** (new: `frontend/Dockerfile`)
   - Multi-stage build: node:20-alpine builder, node:20-alpine runner
   - `npm run build`, expose port 5500 (or 3000)
   - Must work with the existing `compose.yml` service definition

2. **Verify compose.yml**
   - Ensure all 3 services (backend, frontend, caddy) start correctly
   - Fix any port/ networking issues
   - Test: `docker compose up --build`

3. **Shared API client** (`src/app/lib/api.js` — new, extracted from SocialShell)
   - Single `api(path, options)` function used by all pages
   - Base URL config via `NEXT_PUBLIC_API_BASE`
   - Error handling, response parsing

4. **Shared utilities** (`src/app/lib/utils.js` — new)
   - `displayName(user)` — format user name
   - `fileUrl(id)` — build file URL
   - `dateLabel(value)` — format dates
   - `timeAgo(value)` — relative time ("2m ago", "1h ago")

5. **WebSocket client hook** (`src/app/hooks/useWebSocket.js` — new, if Person 3's provider needs refinement)
   - Auto-reconnect with exponential backoff
   - Connection state management
   - Message routing by type

6. **End-to-end integration testing**
   - Test full flows: register → login → create post → follow → message → group → notification
   - Verify WS reconnection works
   - Test session revocation on logout (WS closes)
   - Test group chat (multiple members, messages delivered to all)
   - Test notification delivery (follow request → notification appears)

7. **Error handling + loading states**
   - Ensure all pages show proper loading spinners
   - Ensure all API errors show user-friendly messages
   - Ensure WebSocket disconnect shows offline indicator

8. **Responsive polish**
   - Test all pages on mobile viewports
   - Fix any layout issues
   - Chat: conversation list full-width on mobile, clicking opens thread with back button

### Files to create/modify

| Action | File |
|--------|------|
| New | `frontend/Dockerfile` |
| Modify | `compose.yml` (if needed) |
| New | `src/app/lib/api.js` |
| New | `src/app/lib/utils.js` |
| New | `src/app/hooks/useWebSocket.js` (if needed) |

---

## Coordination Points

### API Contract (MUST be agreed before frontend starts coding)

Person 1 and Person 2 define their API shapes in `docs/API.md`. Person 4 verifies they match the implementation. Person 3 codes against the contracts.

### Notification Wiring

Person 2 owns notification creation logic. Person 1 calls into it when group events happen. Both need to agree on notification types and payloads:

| Trigger | Type | Actor | Content |
|---------|------|-------|---------|
| Follow request (private profile) | `follow_request` | requester | "X wants to follow you" |
| Follow request accepted | `follow_accepted` | accepter | "X accepted your follow request" |
| Group invitation | `group_invite` | inviter | "X invited you to GroupName" |
| Join request to your group | `group_join_request` | requester | "X wants to join GroupName" |
| Event in your group | `group_event` | creator | "X created event: Title" |

### WebSocket Protocol

Both backend (Person 2) and frontend (Person 3) must follow the protocol in `docs/WEBSOCKET-PROTOCOL.md`. Person 4 verifies it works end-to-end.

---

## Start Order

```
Day 1:  Person 1 (groups)  ─── start immediately
        Person 2 (notifs + WS) ─── start immediately
        Person 4 (Docker)   ─── start immediately
        Person 3 (frontend) ─── start with providers + mock data

Day 2+: Person 3 connects to real APIs as Person 1 & 2 finish endpoints
        Person 4 tests end-to-end as pieces come together
```
