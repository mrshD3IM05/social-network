# GROUPS BACKEND IMPLEMENTATION

**Date:** 2026-09-18 · **Branch:** `oualid` · **Scope:** backend + database only. Frontend untouched.

---

## 1. Implementation Summary

The complete Groups backend is implemented according to the subject:

- **Create group** — authenticated user creates a group (title + description); the creator is automatically inserted as the first group member and becomes the group creator (`groups.creator_id`).
- **Browse groups** — authenticated users list all groups with member counts and their own relationship flags (`is_member`, `pending_join`, `is_creator`).
- **Get group** — group detail with creator (public fields only), full member list, member count, and the caller's status (`is_member`, `is_creator`, `pending_invite`, `pending_join`). Non-members who have no relationship to the group get **404** (no enumeration of groups they cannot see).
- **Invite users** — any group member invites an existing, non-member user; self-invites, unknown users, already-members and duplicate invitations are rejected.
- **Accept / refuse invitation** — only the invitation recipient, only while pending; accept joins the group **atomically** (transaction).
- **Join request** — authenticated non-members request to join; members, the creator himself, and duplicate requesters are rejected.
- **Accept / refuse join request** — only the group creator, only while pending; accept joins the user **atomically** (transaction).
- **Notifications** — group invitations and join requests (plus accept/refuse confirmations) are persisted into the existing `notifications` table and pushed in real time through the existing WebSocket hub (`PublishNotification`), matching the payload the frontend already consumes.

No new migrations were needed: the existing schema (migrations `000005`, `000006`, `000007`) already covers groups, members, invitations, join requests and notifications with proper FKs and unique constraints.

---

## 2. Backend Architecture

The project uses a layered architecture: `middleware → handler → service → repository → sqlite`, with the WebSocket hub beside it. Groups follows exactly this flow:

| Layer | File | Role |
|---|---|---|
| Route registration | `backend/internal/server/server.go` | Maps `POST /groups`, `GET /groups`, etc. to handlers behind `auth.Authorized` |
| Handler | `backend/internal/handler/grouphandler/handler.go` | Parses forms/path IDs, resolves the session user via `common.CurrentUserID`, calls the service, maps errors to HTTP status codes |
| Wiring | `backend/internal/handler/handlers.go` | Constructs `grouphandler.New(groupsvc.New(repo, webSocket), sessionsvc.New(repo))` |
| Service | `backend/internal/service/groupsvc/service.go` | All business rules: validation, membership checks, creator checks, duplicate prevention, status transitions, authorization; persists + publishes notifications |
| Repository | `backend/internal/repository/group.go`, `backend/internal/repository/notification.go` | All SQL: group CRUD, member/invitation/join-request queries, transactional accept/refuse helpers, notification inserts |
| Database | SQLite (WAL), embedded migrations `backend/internal/db/migrations/sqlite/` | `groups`, `group_members`, `group_invitations`, `group_join_requests`, `notifications` tables |
| Real-time | `backend/internal/websocket/hub.go` (`PublishNotification`, reused — not modified) | Pushes `{"type":"notification","notification":{...}}` to the recipient's sockets |

Request example (`POST /groups/{id}/invitations`):

```
client → middleware (rate limit + Authorized session check)
       → grouphandler.InviteUser (parse group id + user_id form field, resolve session user)
       → groupsvc.Invite (member check → target exists → not member → no duplicate pending)
       → repository.CreateGroupInvitation / CreateNotification
       → hub.PublishNotification (live push)
       ← 201 + invitation JSON
```

---

## 3. Endpoints

| Method | Endpoint | Description | Authentication | Authorization |
|--------|----------|-------------|----------------|---------------|
| POST | `/groups` | Create a group (form: `title`, `description`) | Session cookie | Any authenticated user |
| GET | `/groups` | Browse all groups with viewer relationship flags | Session cookie | Any authenticated user |
| GET | `/groups/{id}` | Group detail: info, creator, members, viewer status | Session cookie | Members + creator always; invited/pending-request users see limited detail; others 404 |
| GET | `/groups/{id}/members` | Member list | Session cookie | Group members only (403 otherwise) |
| POST | `/groups/{id}/invitations` | Invite a user (form: `user_id`) | Session cookie | Group members only |
| GET | `/group-invitations` | List the caller's pending invitations | Session cookie | Any authenticated user (own invitations only) |
| POST | `/group-invitations/{id}/accept` | Accept a pending invitation | Session cookie | Invitation recipient only |
| POST | `/group-invitations/{id}/decline` | Refuse a pending invitation | Session cookie | Invitation recipient only |
| POST | `/groups/{id}/join-requests` | Request to join the group | Session cookie | Any authenticated non-member (creator himself rejected) |
| GET | `/groups/{id}/join-requests` | List pending join requests | Session cookie | Group creator only |
| POST | `/group-join-requests/{id}/accept` | Accept a join request (creates membership) | Session cookie | Group creator only |
| POST | `/group-join-requests/{id}/decline` | Refuse a join request | Session cookie | Group creator only |

Status-code mapping (consistent with the project's flat-text `http.Error` style):

| Situation | Status |
|---|---|
| Unauthenticated request | 401 |
| Invalid path/body values (bad ID, empty title, self-invite, self-request) | 400 |
| Not found (group, user, invitation, join request) or detail hidden from outsiders | 404 |
| Not a member / not the creator / not the recipient (authorization failures) | 403 |
| Already a member, duplicate pending invitation/request, already-processed invitation/request | 409 |
| Unexpected database failure | 500 |

---

## 4. Database Changes

**Existing tables used (no changes):**

- `groups` — `id` PK, `creator_id` → users CASCADE, `title`, `description`, `created_at` (migration 000005)
- `group_members` — composite PK `(group_id, user_id)` (hard duplicate-membership guarantee), FKs CASCADE (migration 000005)
- `group_invitations` — `id` PK, FKs to groups/users CASCADE, `status` default `'pending'`, **`UNIQUE (group_id, to_user_id)`**, index `(to_user_id, status)` (migration 000005)
- `group_join_requests` — `id` PK, FKs CASCADE, `status` default `'pending'`, **`UNIQUE (group_id, user_id)`**, index `(group_id, status)` (migration 000005)
- `notifications` — `id` PK, FKs to users/actors CASCADE, `group_id` SET NULL, index `(user_id, read)` (migration 000007)

**New migrations:** none — the existing schema already supports every Groups requirement, including duplicate prevention (PK + UNIQUE constraints) and status columns (`pending` / `accepted` / `declined`).

**New constraints/indexes:** none added; the existing ones were verified (primary keys, foreign keys, unique pairs, status indexes).

**Transactions:** new in the repository layer (`withTx` helper): `AcceptGroupInvitationTx`, `AcceptGroupJoinRequestTx`, `RefuseGroupInvitationTx`, `RefuseGroupJoinRequestTx`. Accept flows re-validate ownership + pending status + non-membership **inside** the transaction, insert the membership and update the status in one commit; refuse flows re-validate ownership + pending status before the status update. No partial state is possible on failure. These are the first transactions in the codebase — added only for Groups flows, as instructed, without touching unrelated features.

**Race-condition safety:** beyond the transactional re-checks, concurrent duplicate inserts are stopped by the DB UNIQUE/PK constraints, and the service maps `UNIQUE constraint failed` driver errors to its `Err*Exists` sentinels (verified with 5 concurrent duplicate requests — exactly one 201, the rest 409, one DB row, one notification).

---

## 5. Authorization Rules

| Action | Who is allowed | Enforcement |
|---|---|---|
| Create a group | Any authenticated user | Route behind `auth.Authorized`; handler resolves user from session cookie, never from the request body |
| Browse groups / view own pending invitations | Any authenticated user | Same |
| View group detail | Members and the creator (full); invited / pending-request users (flags only); everyone else gets 404 | `groupsvc.Detail` |
| View member list | Group members only | `groupsvc.Members` → 403 |
| Invite users | Group members only; target must exist, not be a member, not be the inviter, and have no existing invitation row | `groupsvc.Invite` |
| Accept / refuse an invitation | The invitation recipient only, and only while `pending` | Re-validated inside the DB transaction (`ErrNotOwner` → 403) |
| Request to join | Authenticated non-members only; the creator cannot request to join his own group | `groupsvc.RequestJoin` |
| Accept / refuse a join request | The group creator only, and only while `pending`; the request's group is re-checked against the path | `groupsvc.RespondJoinRequest` + transaction re-check |

IDOR hardening: every object-level action resolves ownership server-side — the caller ID always comes from the session cookie; invitation recipient and join-request group are re-verified inside the accepting/refusing transaction, so manipulated IDs cannot grant membership.

---

## 6. Notifications

Integrated with the **existing** notification infrastructure — no second architecture:

- **Persistence** — `repository/notification.go` inserts into the existing `notifications` table (migration 000007) with types `group_invitation`, `group_join_request`, `group_invite_response`, `group_join_response`, plus actor, human-readable content and the group reference.
- **Real-time push** — the service calls the existing `hub.PublishNotification` (`backend/internal/websocket/hub.go`, previously dead code with zero call sites), which emits `{"type":"notification","notification":{...}}` — exactly the event shape `frontend/app/(main)/notifications/page.jsx` already listens for. **The frontend file was not modified.**
- Triggers implemented:
  - invited user gets `group_invitation` when a member invites them;
  - group creator gets `group_join_request` when a user requests to join;
  - inviter gets `group_invite_response` when their invitation is accepted;
  - requester gets `group_join_response` when the creator accepts their request.
- Notification failures are logged and do **not** roll back the group operation (membership is the source of truth), mirroring the hub's best-effort delivery semantics.

Verified end-to-end: a listening WebSocket client received the live `group_invitation` event, and all 7 notification rows were correctly persisted in the DB.

---

## 7. Testing

All scenarios executed against a live server (fresh DB, 6 registered users) via `curl`, plus a raw WebSocket client and direct DB assertions. Build: `go build ./...` and `go vet ./...` both clean.

**Create / browse / detail**
- Unauthenticated create → 401 ✅
- Empty title → 400 ✅; valid create → 201 ✅; creator auto-membership confirmed in DB ✅
- Browse returns groups with `member_count`, `is_member`, `pending_join`, `is_creator` ✅
- Creator sees detail with members and public creator info — **no password hash, email or DOB leak** (fixed during testing) ✅
- Outsider detail → 404; nonexistent group → 404; invalid ID `abc` → 400 ✅

**Invitations**
- Non-member invites → 403 ✅; member invites → 201 ✅
- Duplicate pending invite → 409 ✅; invite to nonexistent group → 403 ✅
- Invite nonexistent user → 404 ✅; self-invite → 400 ✅; bad `user_id` → 400 ✅
- Recipient lists pending invitations ✅; invited user sees limited detail with `pending_invite=true` ✅
- Third party accepting someone else's invitation → 403 ✅
- Recipient accepts → 204, membership created, status `accepted` ✅; re-accept → 409 ✅
- Recipient declines → 204, status `declined` ✅; re-decline → 409 ✅; non-recipient declines → 403 ✅
- Nonexistent invitation → 404 ✅

**Join requests**
- Member requesting to join again → 409 ✅; non-member requests → 201 ✅; duplicate → 409 ✅
- Requester detail shows `pending_join=true` ✅
- Pending list is creator-only: non-creator → 403, creator → 200 ✅
- Non-creator accept → 403; requester self-accept → 403 ✅
- Creator accepts → 204, membership created, status `accepted`, requester notified ✅; re-accept → 409 ✅
- Creator declines → 204, status `declined` ✅; re-decline → 409 ✅; nonexistent request → 404 ✅

**Membership**
- Member list visible to members (200) and hidden from outsiders (403) ✅
- Final DB state: exactly the expected members; all statuses consistent; `PRAGMA foreign_key_check` clean ✅
- Composite PK + UNIQUE constraints verified — no duplicate memberships/invitations/requests possible ✅

**Race conditions**
- 5 concurrent duplicate join requests: one 201, four 409, **one** DB row, **one** notification ✅
- 5 concurrent duplicate invitations: one 201, four 409, **one** DB row ✅

**Notifications**
- WebSocket client (session-cookie auth) received `{"type":"notification","notification":{"type":"group_invitation",...}}` in real time ✅
- All four notification types persisted with correct user/actor/group/content ✅

---

## 8. Remaining Issues

These pre-existing findings from `PROJECT_AUDIT.md` are **outside Groups scope** and were not modified (report-only, per instructions):

1. **BUG-002 (High)** — SQLite `PRAGMA foreign_keys` is applied per-connection but executed once on a pooled connection, so FK enforcement may be off on some connections. Groups code never relies on FK cascades for correctness (explicit deletes/updates), but fixing this globally would strengthen the whole schema.
2. **BUG-003 / BUG-001 (frontend + WS)** — the frontend hardcodes `ws://host:8080` (unreachable in Docker) and websocket session tracking is unwired, so notification *real-time delivery* inherits those limitations. Groups notifications are persisted regardless, so no data is lost.
3. **No pagination** — the project has no pagination pattern anywhere (`GET /posts` returns everything), so `GET /groups` follows suit, as instructed. Easy to add if/when a pattern is introduced.
4. **Invitation/join-request reuse after refusal** — the schema's `UNIQUE (group_id, to_user_id)` / `(group_id, user_id)` keeps one row per pair; after a decline, a new request/invite for the same pair returns 409 rather than resetting the row. Changing this would require altering the existing migration; flagged as a product decision.
5. **Not implemented here (separate subject features, not part of this task):** group posts (`posts.group_id` still never set by `postsvc`), group comments, group chat UI reachability, group events + going/not-going. Tables and the WS group-chat branch already exist for a follow-up.
6. **No unit tests committed** — the repo `.gitignore` excludes `*_test.go`; verification was done with the live integration suite documented above instead.

---

# Files Changed

## Created
```text
backend/internal/service/groupsvc/service.go
backend/internal/handler/grouphandler/handler.go
backend/internal/repository/notification.go
GROUPS_BACKEND_IMPLEMENTATION.md
```

## Modified
```text
backend/internal/model/group.go
backend/internal/repository/group.go
backend/internal/repository/repository.go
backend/internal/server/server.go
backend/internal/handler/handlers.go
backend/readme.md
```

| File | Change |
|---|---|
| `backend/internal/model/group.go` | Mapped `Group` to real schema columns (`title`/`creator_id`), added status constants and member/invitation/join-request/detail/list models |
| `backend/internal/repository/group.go` | Added listing, membership, invitation, join-request queries, aggregate helpers and transactional accept/refuse operations |
| `backend/internal/repository/repository.go` | Added `ErrExists` and `ErrNotOwner` sentinel errors for duplicate/ownership handling |
| `backend/internal/server/server.go` | Registered all 12 Groups routes behind the auth middleware |
| `backend/internal/handler/handlers.go` | Wired the Groups handler and service into the DI graph |
| `backend/readme.md` | Documented the Groups endpoints and notification behavior |

(No unrelated lines were touched in `repository.go`, `handlers.go` or `server.go` beyond the additions above.)

## Deleted
None

**Frontend files modified: NO**
