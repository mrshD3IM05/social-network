# WebSocket Redo — Project Docs

This folder contains the implementation guide for the WebSocket redo.
Follow the order below; each doc builds on the previous.

## Reading Order

1. **[TEAM-ROLES.md](./TEAM-ROLES.md)** — who does what (4-person split)
2. **[WEBSOCKET-PROTOCOL.md](./WEBSOCKET-PROTOCOL.md)** — message format for all WS events
3. **[API.md](./API.md)** — HTTP endpoints the frontend needs (chat history, notifications)
4. **[BACKEND.md](./BACKEND.md)** — step-by-step backend tasks (whoever picks up backend work)
5. **[FRONTEND.md](./FRONTEND.md)** — step-by-step frontend tasks (whoever picks up frontend work)

## What Changed From the Old Code

| Area | Before | After |
|------|--------|-------|
| WS endpoint | `/ws` — messages only | `/ws` — messages + notifications |
| Session tracking | `trackClient`/`untrackClient` never called — logout didn't close WS | Tracking called on connect/disconnect — logout revokes connections |
| Chat history | None — chat started empty every time | `GET /messages/conversations`, `GET /messages/{userId}`, `GET /messages/group/{groupId}` |
| Notification history | None — only real-time | `GET /notifications` with pagination + mark-read |
| Unread counts | None | Notification unread count only (no per-conversation message badges) |
| Frontend WS | Separate connection per page | Single shared connection via React Context |
| Reconnection | None — WS drop = dead page | Exponential backoff reconnect |
| Chat UI | User ID input box | Conversation list sidebar, message threads, emoji picker |

## Dependency Graph

```
BACKEND                           FRONTEND
──────                            ────────
Fix session tracking              WebSocketProvider (connects to /ws)
  │                                 │
  ├─ Message HTTP endpoints        ├─ UnreadProvider (fetches counts)
  │  (conversations, history)      │
  │                                 ├─ ChatLayout (conversation list)
├─ Notification HTTP endpoints   │
  │  (list, mark-read)             ├─ MessageThread (message bubbles)
  │                                 │
  ├─ Wire notifications into       ├─ EmojiPicker (standalone)
  │  existing flows                │
  │                                 ├─ Rewrite messages/page.js
  ├─ Register routes in            │
  │  server.go                     ├─ Rewrite notifications/page.js
  └─                                ├─ Update SocialShell (badge)
                                    └─ Update layout.js (providers)
```

## API Contract

The backend person MUST publish request/response shapes before the frontend person starts
coding against them. See [API.md](./API.md) for the agreed contracts.
