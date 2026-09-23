# PROJECT AUDIT — Social Network vs. subject.md

Audit date: 2026-09-21 · Branch `oualid` · Backend: Go 1.25 (module `sn-backend`) · Frontend: Next.js 16 / React 19 · DB: SQLite (golang-migrate, embedded)

Method: every backend `.go` file (40), every migration (18 versions × up/down), every frontend page/component, `compose.yml`, both Dockerfiles, both Caddyfiles and `backend/readme.md` were read in full. `go build ./...` and `go vet ./...` pass with no errors. Nothing below is inferred from file names alone; each claim cites the code that proves it.

---

## Table of contents

1. [Requirements checklist](#1-requirements-checklist)
2. [Frontend audit](#2-frontend-audit)
3. [Backend audit](#3-backend-audit)
4. [Authentication audit](#4-authentication-audit)
5. [Followers audit](#5-followers-audit)
6. [Profile audit](#6-profile-audit)
7. [Posts audit](#7-posts-audit)
8. [Groups audit](#8-groups-audit)
9. [Private chat audit](#9-private-chat-audit)
10. [Group chat audit](#10-group-chat-audit)
11. [Notifications audit](#11-notifications-audit)
12. [SQLite audit](#12-sqlite-audit)
13. [Migration audit](#13-migration-audit)
14. [Images audit](#14-images-audit)
15. [Docker audit](#15-docker-audit)
16. [WebSocket audit](#16-websocket-audit)
17. [Bug audit](#17-bug-audit)
18. [Security audit](#18-security-audit)
19. [Architecture audit](#19-architecture-audit)
20. [Missing requirements](#20-missing-requirements)
21. [Partially implemented requirements](#21-partially-implemented-requirements)
22. [Extra / not required](#22-extra--not-required)
23. [Final compliance table](#23-final-compliance-table)
24. [Final summary](#24-final-summary)

Status legend: ✅ IMPLEMENTED · 🟡 PARTIALLY IMPLEMENTED · ❌ MISSING · 🔴 ERROR/BUG · ➕ EXTRA

---

## 1. Requirements checklist

| # | Subject requirement | Status | Evidence / problem |
|---|---|---|---|
| 1 | JS framework for frontend | ✅ | Next.js 16 App Router used genuinely (pages, layouts, client components). `frontend/` |
| 2 | HTML/CSS/JS frontend, responsive | ✅ | `frontend/app/globals.css` (532 lines, `@media (max-width: 860px)` mobile layout) |
| 3 | Go backend web server | ✅ | `backend/cmd/server/main.go`, `internal/server/server.go` (31 routes) |
| 4 | Sessions & cookies | ✅ | `internal/service/sessionsvc`, cookie `session` HttpOnly SameSite=Lax, 30-day TTL, DB-backed |
| 5 | Images: JPEG/PNG/GIF, stored, served | ✅ | `internal/service/filesvc`, `internal/handler/filehandler`, `GET /fs/{id}` |
| 6 | WebSocket for real time | ✅ (core) | `internal/websocket/hub.go` (gorilla/websocket) — see §16 for caveats |
| 7 | SQLite | ✅ | `internal/db/sqlite/sqlite.go`, WAL + FK pragmas |
| 8 | Migrations create tables on every run | ✅ | `go:embed sqlite/*.sql` + golang-migrate `m.Up()` in `sqlite.go` |
| 9 | Two Docker images (backend, frontend) | 🟡 | Both Dockerfiles exist, but the pair cannot talk to each other as configured — BUG-009/010 |
| 10 | Registration fields (email, password, first/last name, DOB required; avatar/nickname/about optional) | 🟡 | Nickname is *required* (authsvc + register form); avatar field absent from form; about_me present |
| 11 | Stay logged in until logout; logout always available | ✅ | 30-day cookie + `POST /logout` in navbar (`Navbar.jsx`) |
| 12 | Follow / unfollow / request / accept / decline | 🟡 | API complete (`followsvc`), but no way to *list* pending requests and no notification → flow dead-ends |
| 13 | Public profile auto-follow bypass | ✅ | `followsvc.Follow` sets status `accepted` when `target.Private == false` |
| 14 | Profile: user info | ✅ | `GET /user/{id}` → `common.PublicUser` (no password) |
| 15 | Profile: user activity | ❌ | No endpoint, no UI |
| 16 | Profile: user's posts | 🟡 | Profile page filters the global feed client-side; no per-user posts endpoint |
| 17 | Profile: followers & following lists | ❌ | No endpoints, no UI |
| 18 | Public vs private profile visibility rules | ✅ | `usersvc.CanViewProfile` + 403 in `userhandler.GetUser` |
| 19 | Toggle own profile public/private | ❌ | `users.private` column and `repo.UpdateUser` exist, but no route/handler/UI exposes the toggle |
| 20 | Posts: create with image/GIF | ✅ | `POST /posts` + `POST /files` (post_id), `PostForm.jsx` |
| 21 | Posts: public / almost-private / private | 🟡 | `public` and `almost_private` enforced in SQL; `private` uses `post_visibility` which is **never populated** and no UI to pick followers |
| 22 | Comments on posts (with image/GIF) | ❌ | `comments` table exists (migration 3/14), zero Go code, zero UI |
| 23 | Groups: create (title, description) | ✅ | `groupsvc.Create`, `POST /groups` |
| 24 | Groups: invite / accept / refuse invitation, members invite | ✅ | `groupsvc.Invite/RespondInvitation`, transactional accept |
| 25 | Groups: join request, creator accepts/refuses | ✅ | `groupsvc.RequestJoin/RespondJoinRequest`, creator-only checks |
| 26 | Groups: browse all groups | ✅ | `GET /groups` → `GroupListPayload` |
| 27 | Group posts visible to members only | ❌ | `posts.group_id` exists but no create/list path; `ListVisiblePosts` explicitly filters `group_id IS NULL` |
| 28 | Group comments | ❌ | Same as #22 |
| 29 | Group events (title/description/day-time/going/not going) | ❌ | `group_events` + `event_responses` tables exist (migration 6), zero Go code, zero UI |
| 30 | Private chat: follow-relationship rule | 🟡 | `CanMessage` allows messaging **any public-profile user** with no relationship — broader than subject |
| 31 | Private chat: real-time delivery + persistence | 🟡 | Live delivery ✅, persisted to `messages` ✅, but **no history endpoint** and UI says "not saved when you reload" |
| 32 | Chat emojis | ✅ | Plain-text UTF-8 over WS; emoji typeable (no picker — cosmetic) |
| 33 | Group chat room, members only | 🟡 | Backend fully supports `group_id` messages with member check; **no frontend UI** |
| 34 | Notifications visible on every page | ❌ | No global component/badge; only a live-only `/notifications` page |
| 35 | Notification: private-profile follow request | ❌ | `followsvc` never creates a notification |
| 36 | Notification: group invitation | ✅ | `groupsvc.Invite` → `notify()` |
| 37 | Notification: join request to creator | ✅ | `groupsvc.RequestJoin` → `notify()` |
| 38 | Notification: event created for members | ❌ | Events don't exist |
| 39 | Notifications ≠ private messages (displayed differently) | 🟡 | Separate page/type exists, but page is ephemeral (live-only) |
| 40 | Docker: containers start from clean environment | 🔴 | Builds fine, but frontend container cannot reach backend (`localhost:8080` rewrite) and WS bypasses the proxy |

---

## 2. Frontend audit

**Framework.** Next.js `^16.3.4` + React `^19.2.8` (`frontend/package.json`). Genuinely used: App Router file-based routes (`app/(main)/…`, `app/(auth)/…`), layouts, client components, `useRouter/useParams/usePathname`, route groups, `not-found.jsx`. This satisfies "you must use a JS framework".

**Structure.** 8 pages: home feed, people, groups (list + detail), chat (list + conversation), notifications, settings, profile, login/register, plus 8 reusable components (`Navbar`, `PostCard`, `PostForm`, `GroupCard`, `Modal`, `Avatar`, `CharCount`, `Icon`, `PageHeader`). Shared API helpers in `lib/api.js` and validation in `lib/validate.js` mirroring the Go rules (good duplication-with-a-purpose, documented as such).

**Navigation.** Sidebar with 6 links, active-link highlighting via `usePathname`, `/` redirects to `/home`, `(main)/layout.jsx` guards by calling `/me` and redirecting to `/login` on 401.

**Responsiveness.** `globals.css` line 508: `@media (max-width: 860px)` converts sidebar to a sticky top bar, single-column auth grid, full-width buttons. Present and reasonable.

**Frontend↔backend communication.** All HTTP goes through `/api/v1` prefix; `next.config.js` rewrites to `http://localhost:8080`. `credentials: 'include'` everywhere, so the session cookie flows. WebSockets bypass the rewrite and hardcode `ws://${hostname}:8080/api/v1/ws` (see BUG-009).

**UI coverage of subject features present:** register/login, feed + post composer with privacy selector and image attach, reactions, profile card with follow/unfollow/message buttons and "private profile" locked state, groups browse/create/invite/join-request flows, 1:1 live chat, live notifications page, settings (avatar upload, read-only account details).

**UI gaps vs subject:** no comment UI (none exists server-side), no group post/event UI, no follower lists, no profile-privacy toggle, no private-post audience picker, no follow-request inbox, no group-chat room UI.

**Performance notes (minor).**
- `lib/people.js` derives the "people" list by downloading the *entire* feed (`/posts`) and de-duplicating authors. Works, but scales poorly and means users who never posted are invisible to People/Chat/Invite pickers.
- `profile/[id]/page.jsx` loads the whole feed to filter one author's posts (no per-user endpoint).

**Cannot be verified from the provided source code:** actual rendering in a browser; visual responsiveness beyond the CSS present.

---

## 3. Backend audit

**Architecture.** Clean layering, matching `backend/readme.md`:

```
middleware (rate limit, auth) → handler → service → repository → SQLite
websocket hub ↘ repository (direct)
```

- `cmd/server/main.go` — `sqlite.InitDB("sn.db")` → `repository.New` → `server.RegisterRoutes` → `:8080`.
- `internal/server/server.go` — 31 routes on Go 1.22+ method-pattern mux (`"POST /posts/{id}/reactions"`).
- Repositories: raw `database/sql` with **only** parameterized queries (`?` placeholders everywhere — no string-built SQL found).
- Services: pure business logic behind small `Repository` interfaces (dependency-injection friendly, easy to mock).
- Handlers: parse form/path, map service errors to status codes via `common` / `writeError`.
- Error handling: consistent `http.Error` flat-text bodies; `repository.ErrNotFound/ErrExists/ErrNotOwner` sentinel errors mapped to 404/409/403.

**Route inventory (server.go):** auth (register/login/logout/me), user (get/follow/unfollow/respond), posts (CRUD + reactions), files (upload/avatar/download), groups (create/list/detail/members/invitations/join-requests/respond), `GET /ws`.

**Middleware.**
- `middleware/auth.go` — `Authorized` (401 without valid session) and `Guest` (403 when already authenticated) — applied consistently to every route, including `/ws`.
- `middleware/rate_limit.go` — per-IP fixed-window limiter, **extra** feature. Bugs noted (BUG-017, BUG-021).

**Gaps.** No endpoint for: comments, group posts, group events, notification listing, message history, follow-request listing, followers/following lists, profile update/privacy toggle, per-user posts. Several of these are subject requirements — see §20.

**README drift (backend/readme.md).** Claims `GET /users` returns 501 (route not registered at all), claims follow-request notifications over WS (not implemented), claims rate limit of 100/min (code enforces 1000). Treat the README as aspirational in those spots.

---

## 4. Authentication audit

| Requirement | Status | Evidence |
|---|---|---|
| Registration with email/password/first/last/DOB | ✅ | `authsvc.Register` + `validateRegisterInput` (regex email, DOB parse `2006-01-02`, no future dates) |
| Nickname optional | 🔴 BUG-018 | `authsvc`: nickname **required**; frontend `register/page.jsx` marks it required too. Subject lists it as optional |
| Avatar optional in form | 🔴 BUG-019 | Register form has **no avatar input**; backend reads `avatar` as a plain form *string* (a file ID, not an upload) — avatar can only be set later via `POST /avatar` |
| About me optional | ✅ | Optional both sides |
| Password storage | ✅ | bcrypt `GenerateFromPassword`/`CompareHashAndPassword` (`authsvc`); hash never serialized to clients (`common.PublicUser/PrivateUser` omit it) |
| Sessions | ✅ | 32 random bytes (`crypto/rand`), stored in `sessions` table, 30-day TTL, expiry checked on every `Get`, expired rows deleted |
| Cookies | ✅ | `session` cookie: `HttpOnly`, `SameSite=Lax`, `Path=/`; cleared with `MaxAge=-1` on logout. No `Secure` flag (BUG-023) |
| Stay logged in | ✅ | Persistent cookie + DB session; `(main)/layout.jsx` restores user via `/me` |
| Logout available at all times | ✅ | Navbar logout button on every page → `POST /logout` → session deleted, cookie cleared |
| Login with email or nickname | ➕ | `authsvc.Login` accepts either identifier (not required by subject; harmless) |
| Authorization model | ✅ | Every route wrapped in `Authorized`; per-object checks in services (`ErrNotRecipient`, `ErrNotGroupCreator`, owner-only post update/delete) |

Guest-only enforcement on `/register` and `/login` (403 when a cookie is presented) is slightly stricter than the subject but coherent.

---

## 5. Followers audit

Implemented in `internal/service/followsvc/service.go` on top of the `follow_requests` table (one row per (from,to), `status` = pending/accepted/declined — the table doubles as the follow graph, which is fine but the name is misleading).

- **Send follow request** ✅ — `POST /users/{id}/follow`. Public target → row created/updated with `status=accepted` (auto-follow bypass ✅). Private target → `pending`.
- **Accept / decline** ✅ at API level — `POST /follow-requests/{id}/accept|decline`; `Respond` verifies the caller is the recipient (`ErrNotRecipient`) and the request is still pending.
- **Unfollow** ✅ — `DELETE /users/{id}/follow` deletes the row (so re-follow is possible — good).
- **Privacy-aware visibility** ✅ — `usersvc.CanViewProfile` grants private-profile content only to accepted followers; `ListVisiblePosts` uses the same rule in SQL.
- 🔴 **BUG-007/008 — the request flow dead-ends.** There is **no endpoint to list pending follow requests** for the recipient, and `followsvc` **never creates a notification** (the required notification for private-profile follow requests). The frontend has no request inbox. A recipient can only accept a request if they guess its numeric ID. Combined, private-profile following is not completable through the product.
- 🟡 Frontend `profile/[id]/page.jsx` always renders both Follow and Unfollow buttons without knowing current relationship state (minor UX, API rejects duplicates with 409).

---

## 6. Profile audit

- **User information** ✅ — `GET /user/{id}` returns `common.PublicUser`: id, first/last name, avatar, nickname, about_me, private flag, created_at. **Password is never included** ✅, email/DOB only in `PrivateUser` (self) ✅.
- **Visibility rules** ✅ — private profile → 403 `profile is private` unless the viewer follows the owner (`usersvc.CanViewProfile`). Own profile always visible.
- **User activity** ❌ — nothing (no endpoint, no UI).
- **Posts of the user** 🟡 — frontend filters `/posts` by `author_id`. Because the feed already applies privacy rules, this leaks nothing, but it misses posts the viewer may legitimately see only via other filters and is inefficient.
- **Followers / following lists** ❌ — required by subject; no endpoints, no UI.
- **Toggle public/private** ❌ — BUG-001. `users.private` (INTEGER, default 0) and `repo.UpdateUser` exist, but no route, handler or UI exposes the switch. Settings page literally says: *"Editing these needs an API endpoint that does not exist yet."* This is a headline subject requirement.
- **Privacy-leak check** ✅ — group payloads go through `model.GroupCreator` (comment in `repository/group.go` documents that password/email/DOB never reach clients). `GET /user/{id}` on a private profile returns 403 rather than a redacted body — no partial leak.

---

## 7. Posts audit

- **Create** ✅ — `POST /posts` (content, privacy ∈ {public, almost_private, private}; validated in `postsvc.validPrivacy`).
- **Update/Delete own post** ➕ — owner-only (`UpdatePostOwned`/`DeletePostOwned` with `author_id` predicate). Not required by subject; harmless.
- **Privacy enforcement** (`repository/post.go::postVisibleCondition`):
  - public → everyone ✅
  - almost_private → viewer has accepted follow **to the author** ✅
  - private → row in `post_visibility` for (post, viewer) ✅ *in SQL*…
- 🔴 **BUG-002 — private posts are unusable.** Nothing ever inserts into `post_visibility`: no endpoint accepts a list of chosen followers, `postsvc.Create` doesn't take one, the composer has no audience picker. A "private" post is therefore visible to its author only. The subject's "only the followers chosen by the creator" is not deliverable.
- **Images/GIFs** ✅ — composer uploads up to 3 files to `POST /files` with `post_id`, images listed from `files` table (`ListPostFileIDs`), rendered via `GET /fs/{id}` with per-viewer visibility (`CanViewFile`). GIF accepted (`image/gif` in `allowedImageType`).
- **Comments** ❌ — BUG-003. The `comments` table exists (created migration 000003, `image` column dropped in 000014) but there is no Go code, route, or UI. The subject requires commenting, with optional image/GIF, respecting post permissions. `model.ReactionTargetComment` and `files.comment_id` are vestiges of an abandoned start.
- **Reactions** ➕ — like/dislike with toggle + summary (`repository/reaction.go`), properly gated by `CanViewPost`. Extra feature.

---

## 8. Groups audit

Backend `internal/service/groupsvc/service.go` + `internal/handler/grouphandler/handler.go` — the strongest part of the project.

- **Create group (title + description)** ✅ — validation (title ≤100 chars, description ≤1000), creator auto-added to `group_members` ✅.
- **Invite users** ✅ — any *member* can invite (`IsGroupMember` gate); rejects self-invite, unknown users, existing members, duplicates (409); invitation persisted in `group_invitations` with UNIQUE(group,to_user).
- **Accept / refuse invitation** ✅ — `AcceptGroupInvitationTx` inserts membership **and** flips status inside one transaction; `RefuseGroupInvitationTx` refuses only pending + recipient-owned. Recipient-only enforced (`ErrNotOwner`).
- **Join requests** ✅ — `POST /groups/{id}/join-requests` (non-members only, creator excluded); `GET /groups/{id}/join-requests` creator-only; `RespondJoinRequest` verifies `group.CreatorID == creatorID` before accepting/refusing, accept is transactional.
- **Browsing** ✅ — `GET /groups` returns all groups with member_count, is_member, pending_join, is_creator (batched map lookups, no N+1).
- **Group detail privacy** ✅ (design choice) — outsiders get 404 unless invited/pending (`groupsvc.Detail`), members list 403 for non-members.
- **Notifications for invite / join-request / responses** ✅ — persisted via `CreateNotification` + live push `hub.PublishNotification`.
- 🔴 **BUG-015** — after an invitation or join request is declined, the UNIQUE(group_id,to_user_id) constraint plus the service's "processed rows are history" logic make re-inviting / re-requesting impossible forever (returns 409). A declined user can never enter the group.
- **Group posts** ❌ — BUG-004. `posts.group_id` column exists, but `CreatePost` never sets it, `ListVisiblePosts` filters `p.group_id IS NULL`, and there is no member-gated group-post route. The subject's "posts and comments only displayed to members of the group" is absent.
- **Group comments** ❌ — same as post comments.
- **Group events** ❌ — BUG-005. `group_events` (title/description/date_time) and `event_responses` (choice, UNIQUE(event,user)) are fully modeled in migration 000006 and then never touched by any Go code, route, or UI. Going/Not-going is entirely missing, as is the "event created" notification.
- **Group chat** — backend ready, UI missing (§10).

---

## 9. Private chat audit

- **Transport** ✅ — single WebSocket `GET /ws` (gorilla), cookie-authenticated before upgrade, per-user client map, sender echo, multiple tabs supported (map of clients per user ID).
- **Persistence** 🟡 — messages are written to `messages` (`CreateMessage`) ✅, but there is **no endpoint to read history**; `chat/[id]/page.jsx` shows only messages received since page load and even prints *"Messages are live only and are not saved when you reload."* The data is stored; the feature (conversations) isn't retrievable. BUG-013.
- **Authorization to send** 🔴 BUG-014 — subject: "at least one of the users must be following the other". `repository.CanMessage` allows messaging whenever `target.private = 0` **regardless of any follow relationship**, i.e. any user can DM any public-profile stranger. Broader than the spec.
- **Delivery rule** — subject: recipient receives instantly if they follow the sender **or** the recipient has a public profile. The implementation just publishes to the recipient's connected clients whenever the send was allowed; combined with BUG-014 the effective behavior is "anyone public receives", which covers the subject's public-profile exception but not its restriction.
- **Emojis** ✅ — content is arbitrary UTF-8 text; emojis work. No emoji picker (cosmetic).
- **Conversations list** 🟡 — `/chat` derives contacts from feed authors (`fetchPeople`); no conversation-history basis.
- **Error handling** ✅ — server emits `{"type":"error"}` for invalid payloads, non-permitted sends, and save failures; client displays it.
- **Lifecycle** 🟡 — client closes socket on unmount; no auto-reconnect, no `onclose`/`onerror` UI state. Logout does **not** close sockets (BUG-011).

---

## 10. Group chat audit

- **Backend** ✅ — the hub accepts `{"type":"message","group_id":N,"content":…}`, `CanMessage` verifies sender membership in `group_members`, the message is persisted with `group_id`, and `GroupMemberIDs` fans the event out to every member (including the sender). The `messages` table CHECK constraint was rebuilt twice (migrations 000008, 000017) specifically to permit group rows (`to_user_id IS NULL`).
- **Frontend** ❌ — no group chat room exists in the UI; `chat/[id]/page.jsx` is strictly 1:1 (`to_user_id`). The subject's "common chat room" is therefore not user-reachable. BUG-006.
- **History** — same gap as private chat: nothing reads `messages` back.

---

## 11. Notifications audit

- **Types implemented** (`model/group.go` consts): `group_invitation`, `group_join_request`, `group_invite_response`, `group_join_response`. Created in `groupsvc.notify()` → persisted in `notifications` → `hub.PublishNotification` pushes `{"type":"notification",…}` to the recipient's sockets in real time ✅.
- **Required by subject but absent:**
  - follow request on a private profile — `followsvc` has no notification code at all (❌ BUG-007);
  - event creation for group members — no events (❌).
- **Visible on every page** ❌ — the navbar has a bell *link*, but no badge/count/live feed; nothing通知-like renders on other pages. BUG-016.
- **Separate from private messages** 🟡 — separate page and separate WS event type ✅, but the page is ephemeral: it only accumulates events received while open, and there is **no GET endpoint** for stored notifications, so the persisted rows are never displayed after a reload.
- **Read/unread** — `notifications.read` column exists and is indexed, but nothing ever sets or filters on it; `model.Notification.Read` is only read back at insert.
- **Recipient correctness** ✅ — each notification's `user_id` is the invitee / creator / inviter / requester as appropriate (verified per call site in `groupsvc`).

---

## 12. SQLite audit

Driver: `mattn/go-sqlite3`; opened in `db/sqlite/sqlite.go` with `PRAGMA foreign_keys=ON`, `journal_mode=WAL`, `busy_timeout=5000`. **Connection:** file `sn.db` in the working directory.

**Actual schema (after migrations 000001–000018; snapshot also kept in `backend/schema.sql`):**

```
users (id PK, email UNIQUE, password, first_name, last_name, date_of_birth,
       avatar, nickname, about_me, private, created_at)
 ├─1:N─ sessions (id PK, user_id FK→users CASCADE, expires_at)
 ├─N:M─ follow_requests (from_user_id FK, to_user_id FK, status, UNIQUE(from,to))   [also = follow graph]
 ├─1:N─ posts (author_id FK→users CASCADE, privacy, group_id FK→groups CASCADE, created_at)
 │        ├─1:N─ comments (post_id FK CASCADE, author_id FK CASCADE)      ← UNUSED by code
 │        ├─1:N─ post_visibility (post_id,user_id PK, FK CASCADE)         ← NEVER POPULATED
 │        └─1:N─ files (id TEXT PK, owner FK CASCADE, post_id FK SET NULL,
 │                    comment_id FK SET NULL, message_id FK SET NULL)
 ├─1:N─ reactions (target_type post|comment, target_id, user_id FK, UNIQUE triple) [comment target unused]
 ├─N:M─ group_members (group_id FK CASCADE, user_id FK CASCADE, PK(group,user))
 ├─1:N─ groups (creator_id FK→users CASCADE, title, description)
 │        ├─1:N─ group_invitations (group FK, from/to user FK, status, UNIQUE(group,to))
 │        ├─1:N─ group_join_requests (group FK, user FK, status, UNIQUE(group,user))
 │        ├─1:N─ group_events (group FK, creator FK, title, description, date_time) ← UNUSED
 │        │        └─1:N─ event_responses (event FK, user FK, choice, UNIQUE(event,user)) ← UNUSED
 │        └─1:N─ messages (group_id FK CASCADE) + 1:1 private (to_user_id FK)
 ├─1:N─ messages (from_user_id FK, to_user_id FK?, group_id FK?, CHECK(from<>to OR to IS NULL))
 └─1:N─ notifications (user_id FK, actor_id FK, type, content, group_id FK SET NULL, read)
schema_migrations (golang-migrate bookkeeping)
```

**Strengths.** Proper FKs with intentional CASCADE/SET-NULL choices; composite PKs/UNIQUE constraints prevent duplicates exactly where needed (follows, memberships, invitations, join requests, event responses, reactions); sensible indexes on every hot lookup path (`to_user_id+status`, `group_id+status`, `user_id+read`, posts by author/group, messages by to/group); transactions used for accept flows (`withTx`); reactions use `ON CONFLICT … DO UPDATE` upsert.

**Issues.**
- 🔴 **BUG-012** — `PRAGMA foreign_keys` and `busy_timeout` are per-connection in SQLite; `db.Exec("PRAGMA …")` applies them to *one* pooled connection, and `SetMaxOpenConns(1)` is not called. FK enforcement is therefore unreliable for queries served by other pool connections. (WAL is persistent, so that one sticks.)
- 🟡 `messages` was rebuilt twice (000008 then 000017) — migration history noise, harmless.
- 🟡 `comments`, `post_visibility`, `group_events`, `event_responses`, `files.comment_id`, reactions-on-comments are schema-only — the unused-schema footprint of the missing features.
- 🟡 `follow_requests` naming (it stores accepted follows too) — works, but confusing.
- 🟡 No composite index for `messages` conversation reads (from+to) — moot until a history endpoint exists.

**No SQL-injection surface:** every query in `internal/repository/*.go` uses placeholders; no fmt-concatenated SQL found.

---

## 13. Migration audit

- **System** ✅ — `golang-migrate/migrate/v4` with `source/iofs` over `//go:embed sqlite/*.sql` (`internal/db/migrations/migrations.go`) and `database/sqlite3.WithInstance`. Applied automatically in `sqlite.InitDB` → `Migrate` → `m.Up()` (tolerating `ErrNoChange`) on **every** start, satisfying "every time the application runs, it creates the specific tables".
- **Files** ✅ — 18 versions, each with a matching `.up.sql` **and** `.down.sql` (37 SQL files verified), numbered `000001_…` → `000018_…` in `backend/internal/db/migrations/sqlite/`.
- **Folder structure** 🟡 — subject's example tree is `backend/pkg/db/migrations/sqlite`; the project uses `backend/internal/db/migrations/sqlite`. The subject explicitly allows other organization ("It can be organized as you wish"), and the naming pattern matches what testers look for (`file://…/migrations/sqlite` style). Note it for evaluators anyway.
- **Order** ✅ — users → sessions → posts/comments → follows → groups → events → notifications/messages → reactions → files → data fixes → schema fixes. Dependencies respected (e.g. files references messages, created after; `post_type` added in 16 and dropped in 18 — redundant but valid).
- **Clean-database behavior** ✅ (by code path) — `InitDB` opens, pragmas, migrates; `go build` passes; migration 1 creates `users` first. Not executed live during this audit — runtime creation from a pristine DB **cannot be verified from the provided source code** beyond the code path review.
- **`sqlite.go` role** ✅ — connection + pragmas + migration application, exactly what the subject asks that file to do.
- **Down migrations** present; 000014's down correctly rebuilds `comments` with the `image` column — consistent.

---

## 14. Images audit

`internal/service/filesvc/service.go`, `internal/handler/filehandler/handler.go`, served by `GET /fs/{id}`.

- **Formats** ✅ — `http.DetectContentType` magic-byte sniffing (512-byte header), allow-list `image/jpeg | image/png | image/gif`. Extension/MIME from the client is ignored — spoof-resistant.
- **Limits** ✅ — 10 MB per file (`MaxImageSize`), max 3 per upload (`MaxImages`), `http.MaxBytesReader` on the body, `io.LimitReader` while copying.
- **Storage** ✅ — random 16-byte hex ID (`crypto/rand`), file written to `uploads/<id>` with `O_EXCL`, mode 0640, dir 0750; metadata (original name via `filepath.Base`, mime, size, owner, post/message linkage) in the `files` table. On DB failure the file is removed.
- **Path traversal** ✅ — the client never influences the storage path; serving uses the stored `storage_path`, and the ID is server-generated. `original_name` is sanitized with `filepath.Base`.
- **Retrieval authorization** ✅ — `CanViewFile` mirrors post privacy (owner, any-user avatar, public/followers/selected posts, message participants, group members) before `http.ServeFile`; invisible → 404. `Cache-Control: private, immutable` is appropriate for content-addressed IDs.
- **Avatar** ✅ — `POST /avatar` (single file, same checks) updates `users.avatar` to the file ID; rendered by `Avatar.jsx` via `imageUrl(id)`.
- **Gaps** — register-time avatar not supported (BUG-019); orphan files never garbage-collected (old avatars/uploaded-but-unattached files accumulate) — minor.

---

## 15. Docker audit

**Images.** `backend/Dockerfile`: multi-stage `golang:1.25-alpine` (build-base for CGO/sqlite) → `alpine:3.22` runtime, `CGO_ENABLED=1 go build`, exposes 8080, runs `/server`. Sound for mattn/go-sqlite3. `frontend/Dockerfile`: `node:22-alpine`, `npm ci`, `npm run build`, `npm run start`, exposes 3000. Both are valid single images each — the "two Docker images" requirement is structurally met (Caddy is a third, extra container).

**Compose** (`compose.yml`): backend + frontend on a private network, `expose` (not published) for 8080/3000, Caddy published on 8000:80 proxying `/api/v1/*` → `backend:8080` and everything else → `frontend:3000`; named volumes `backend-data:/app` (DB) and `backend-uploads:/app/uploads`. Caddy terminates and proxies WebSocket upgrades correctly *by configuration*.

🔴 **BUG-010 — the two required containers cannot talk.** `next.config.js` rewrites `/api/v1/*` to `http://localhost:8080`. Rewrites are executed by the Next **server** inside the frontend container, where `localhost:8080` is nothing (backend is a different container). Every API call fails in the composed deployment. Destination must be `http://backend:8080` (or via env var).

🔴 **BUG-009 — WebSocket bypasses everything.** The chat/notifications pages dial `ws://${hostname}:8080/api/v1/ws`. In the composed setup nothing publishes 8080 to the host → connections fail. (Next rewrites can't proxy WS either, which is presumably why 8080 was hardcoded for dev.)

**Other observations.**
- No environment variables anywhere; ports/URLs are hardcoded (host dev assumptions baked into images).
- DB persists via the `backend-data` volume; a clean `docker compose up` should boot and migrate — but the app is unusable due to BUG-009/010, so *realistic clean-environment startup* fails today.
- `backend-data` mounted over `/app` (the image's WORKDIR) is unusual, though harmless since the binary lives at `/server`; the nested `backend-uploads` mount works.
- Subject ports guidance is satisfied only through Caddy (8000→frontend), not by the frontend container itself — acceptable, worth noting.

---

## 16. WebSocket audit

`internal/websocket/hub.go` (+ `session.go`).

- **Upgrade & auth** ✅ — `Authorized` middleware + explicit cookie/session re-check in `ServeHTTP` before `websocket.Upgrader.Upgrade`. Unauthenticated upgrade → 401 before the connection exists.
- **Registration/disconnect** ✅ — `Hub.clients map[userID]map[*Client]struct{}` guarded by `sync.RWMutex`; `add` on connect, `remove` in `readPump`'s defer, empty per-user maps deleted. Multiple tabs/devices per user supported.
- **Concurrency** ✅ — one read goroutine (the handler's) + one `writePump` goroutine per client; all sends go through the buffered `send` channel (16) owned by `writePump` (single writer per conn — correct per gorilla). `publish` takes `RLock` and uses non-blocking sends (slow clients drop rather than block the hub).
- **Routing** ✅ — 1:1 (`publish(to)` + echo to sender) and group fan-out via `GroupMemberIDs`.
- **Health** ✅ — 45 s ping ticker with 10 s write deadline; 60 s read deadline reset by pong handler; 64 KB read limit.
- **Errors** ✅ — malformed frames, unauthorized targets and save failures emit `{"type":"error",…}` without killing the connection; read errors tear down cleanly.
- 🔴 **BUG-011** — `websocket/session.go` (`trackClient`/`untrackClient`/`RevokeSessionClients`) is **dead code**: nothing calls track/untrack (confirmed by search; the project's own `docs/BACKEND.md` admits it). Logout's `RevokeSessionClients(cookie.Value)` therefore finds nothing — a logged-out tab keeps a live, authenticated socket indefinitely.
- 🟡 `CheckOrigin: func(*http.Request) bool { return true }` — combined with cookie auth this permits cross-site WebSocket hijacking (BUG-022).
- 🟡 No deadline on the time between `add` and `readPump` starting — fine in practice since it's sequential.

---

## 17. Bug audit

Only code-supported findings. Severity: Critical = breaks a subject requirement end-to-end; High = breaks a feature or deployment; Medium = functional deviation; Low = cosmetic/robustness.

| ID | Severity | Location | Problem / Why / Expected vs current / Fix |
|---|---|---|---|
| BUG-001 | **Critical** | `internal/server/server.go` (no route), `internal/repository/user.go::UpdateUser` (orphaned), `frontend/app/(main)/settings/page.jsx` | **No way to toggle profile public/private.** Subject requires the option on the own profile; the column and repo method exist but no handler/route/UI. Expected: user can switch. Current: impossible; settings page admits it. Fix: add `PUT /me` (or `/users/{id}/privacy`) handler calling `UpdateUser`, wire a toggle in Settings/Profile. |
| BUG-002 | **Critical** | `internal/service/postsvc/service.go::Create`, `repository/post.go` (`post_visibility` condition), `frontend/components/PostForm.jsx` | **Private posts have no audience selection.** `post_visibility` is never written; no endpoint/UI picks followers. Expected: only *chosen* followers see the post. Current: only the author sees it. Fix: accept a `viewers[]` list on create/update, insert `post_visibility` rows, add UI picker of accepted followers. |
| BUG-003 | **Critical** | `internal/db/migrations/sqlite/000003…` (table) vs. no Go code/routes | **Comments not implemented at all** (nor comment images). Subject requires commenting on posts with permission inheritance. Fix: repository+service+handler for `GET/POST /posts/{id}/comments` reusing `CanViewPost`; extend `PostCard.jsx`. |
| BUG-004 | **Critical** | `repository/post.go::ListVisiblePosts` (`p.group_id IS NULL`), `postsvc.Create`, no group-post route/UI | **Group posts not implemented.** Column exists; nothing sets or reads it. Subject: group posts visible only to members. Fix: member-gated `POST/GET /groups/{id}/posts`, drop the NULL filter for member viewers or query separately, add UI. |
| BUG-005 | **Critical** | `migrations/000006_*` (tables) vs. no Go code/routes/UI | **Group events + going/not-going not implemented**, incl. the required event-created notification. Fix: event CRUD + `event_responses` endpoints, member-gated; notification to members on create; UI in group detail. |
| BUG-006 | High | `internal/websocket/hub.go` (group branch works) vs. `frontend/app/(main)/chat/*` | **Group chat has no UI.** Backend routes/persists/fans out group messages; frontend only ever sends `to_user_id`. Subject: members chat in a common room. Fix: add group room page reusing the hub protocol with `group_id`. |
| BUG-007 | **Critical** | `internal/service/followsvc/service.go` (no notification call) | **No follow-request notification.** Subject explicitly requires it for private profiles; groupsvc shows the pattern exists. Fix: after creating a pending request, `CreateNotification` + `hub.PublishNotification` to the target. |
| BUG-008 | High | `internal/server/server.go`, frontend | **No listing of pending follow requests** (no endpoint, no inbox UI); recipient can only respond by ID. With BUG-007 the private-follow flow is unreachable in practice. Fix: `GET /follow-requests` + inbox UI (notifications page or profile). |
| BUG-009 | High | `frontend/app/(main)/chat/[id]/page.jsx` line ~26, `frontend/app/(main)/notifications/page.jsx` line ~12 | **WS URL hardcoded to `ws://hostname:8080`** — fails behind Caddy/compose (nothing on host :8080) and breaks TLS. Fix: derive from `location` (e.g. `wss?://host/api/v1/ws`) and proxy WS through Caddy. |
| BUG-010 | High | `frontend/next.config.js` | **Rewrite target `http://localhost:8080` unreachable from the frontend container** — the whole API is dead in `docker compose up`. Fix: `http://backend:8080`, ideally from `process.env.BACKEND_URL`. |
| BUG-011 | Medium | `internal/websocket/session.go` (dead code), `internal/handler/authhandler/handler.go::Logout` | **Logout doesn't close live WebSockets.** `trackClient/untrackClient` never called; `RevokeSessionClients` no-ops. Expected: revoked session loses its sockets. Fix: call `trackClient(cookie.Value, client)` after add and `untrackClient(c)` in readPump's defer. |
| BUG-012 | Medium | `internal/db/sqlite/sqlite.go::Open` | **FK/busy_timeout pragmas applied to a single pooled connection** (per-connection pragmas + `database/sql` pool). FK enforcement unreliable. Fix: `db.SetMaxOpenConns(1)` (fine for SQLite/WAL) or set pragmas in a `connect hook`. |
| BUG-013 | Medium | `internal/server/server.go` (no message-read route), `frontend/app/(main)/chat/[id]/page.jsx` | **No message history** despite persistence; UI states "not saved when you reload". Subject implies conversations. Fix: `GET /messages?user_id=` / `?group_id=` with participation checks; load on page open. |
| BUG-014 | Medium | `internal/repository/message.go::CanMessage` | **DM authorization broader than subject**: `target.private = 0 OR <follow exists>` allows DMing any public stranger with no relationship. Subject: at least one must follow the other. Fix: drop the `private = 0` short-circuit, keep the follow-exists clause (optionally keep public-recipient rule only for *delivery*, per subject wording). |
| BUG-015 | Medium | `internal/service/groupsvc/service.go::Invite/RequestJoin`, `migrations/000005` UNIQUE(group,to_user) | **Declined invitations/join requests are permanent blockers** — re-inviting or re-requesting returns 409 forever. Fix: on decline, delete the row (or relax the UNIQUE constraint to partial-index on pending). |
| BUG-016 | High | `frontend/components/Navbar.jsx`, no notification GET route | **Notifications not visible on every page** (no badge/global component) and **never listed after reload** (no GET endpoint) — persisted rows are unreadable. Subject: notifications visible on every page. Fix: `GET /notifications` (+ mark-read), global bell with unread count fed by the existing WS. |
| BUG-017 | Low | `internal/middleware/rate_limit.go` | Headers advertise `X-RateLimit-Limit: 100` while 1000 is enforced; comment contradicts itself. Fix: single constant used for both. |
| BUG-018 | Low | `internal/service/authsvc/service.go::validateRegisterInput`, `frontend/app/(auth)/register/page.jsx` | **Nickname required** though subject marks it optional (and `users.nickname` has `DEFAULT ''`). Fix: drop the required/regex gate when empty (keep uniqueness if enforced later). |
| BUG-019 | Low | `frontend/app/(auth)/register/page.jsx`, `authhandler.Register` | **No avatar field in the registration form** (subject: field must be present, skippable); backend accepts only a string, not an upload. Fix: add optional file input, upload after account creation via `/avatar`. |
| BUG-020 | Medium | `userhandler`/`server.go` | **No followers/following endpoints or UI** and **no user-activity feed** — required profile sections absent (overlap with §20). Fix: `GET /users/{id}/followers|following` gated by `CanViewProfile`; render lists on the profile. |
| BUG-021 | Low | `internal/middleware/rate_limit.go` | Client map grows unbounded (entries never evicted) — slow memory leak on hostile traffic. Fix: periodic sweep or eviction on window expiry. |
| BUG-022 | Low→Medium | `internal/websocket/hub.go::ServeHTTP` | `CheckOrigin` accepts all origins; with cookie auth this enables cross-site WebSocket hijacking. Fix: echo-only origin allow-list. |
| BUG-023 | Low | `internal/service/sessionsvc/service.go::SetCookie` | No `Secure` attribute on the session cookie (fine for local HTTP, wrong in production). Fix: `Secure: true` behind TLS. |
| BUG-024 | Low | `backend/readme.md` | Docs drift: claims `GET /users` 501 (route not registered at all), follow-request WS notifications (not implemented), 100/min rate limit (code: 1000). Fix: align docs. |
| BUG-025 | Low | `frontend/lib/people.js` | People/Chat/Invite pickers only show users who authored visible posts — you cannot discover or DM a follower who never posted. Fix: real `GET /users` search endpoint (or followers-derived list). |
| BUG-026 | Low | `frontend/app/(main)/chat/[id]/page.jsx` | No WS `onclose`/`onerror` handling or reconnection; the chat silently dies. Fix: status indicator + reconnect backoff. |

Non-bugs explicitly checked and found correct: bcrypt compare order; `Respond` recipient check; owner-only post update/delete predicates; group accept transactions; invitation/join-request duplicate handling (aside from BUG-015); `CanViewFile` parity with post visibility; `go build`/`go vet` clean.

---

## 18. Security audit

| Area | Finding |
|---|---|
| Password hashing | ✅ bcrypt (DefaultCost) in `authsvc`; hash never leaves the server (`PublicUser`/`PrivateUser` omit it; `model.User` JSON tags exist but that struct is never marshaled directly to clients for user endpoints) |
| Sessions | ✅ 256-bit random IDs, DB-backed, expiry enforced server-side, deleted on logout. 🟡 no server-side re-validation of the cookie value against timing attacks (acceptable), no rotation on privilege change |
| Cookies | 🟡 `HttpOnly`, `SameSite=Lax`; **no `Secure`** (BUG-023) |
| Authentication middleware | ✅ applied to all 31 routes incl. `/ws`; `Guest` prevents session fixation via re-login of an authed user |
| Authorization / IDOR | ✅ strong overall: post update/delete owner-scoped in SQL; follow response recipient-checked; group join-response creator-checked; invitation response recipient-checked (transactional); file downloads visibility-checked. ❌ **profile privacy toggle missing** (BUG-001) means the `private` flag is currently immutable — no bypass, but no feature either |
| Private profile visibility | ✅ enforced at both profile (403) and feed-SQL level |
| Private post visibility | 🔴 enforced in SQL but feature-incomplete (BUG-002) |
| Group authorization | ✅ membership/creator checks in service layer before every mutation; members list 403 for outsiders; detail 404 |
| WebSocket authorization | ✅ session checked pre-upgrade and per-message via `CanMessage`; 🟡 `CheckOrigin: true` (BUG-022), 🟡 logged-out sockets survive (BUG-011) |
| SQL injection | ✅ none found — 100% parameterized queries across `internal/repository` |
| XSS | ✅ React escapes all interpolated content; images served from same-origin `/api/v1/fs/<id>` with sniffed content-type; no `dangerouslySetInnerHTML` anywhere |
| CSRF | 🟡 `SameSite=Lax` + JSON-ish APIs mitigates the classic cases; no CSRF token (state-changing POSTs rely on Lax); WebSocket path covered by BUG-022 |
| File upload validation | ✅ magic-byte sniffing, size/count caps, `O_EXCL` random paths, `MaxBytesReader` — no path traversal, no executable storage paths |
| Sensitive exposure | ✅ password never serialized; email/DOB only on `PrivateUser` (self); group creator mapped to a public subset. 🟡 `model.User`'s JSON tags would leak if ever marshaled directly — currently never happens for other users |
| Secrets/API keys | ✅ none present |
| CORS | ✅ n/a (same-origin proxy); frontend never hits the backend cross-origin |
| Input validation | ✅ mirrored client+server for auth/posts/groups/messages; 🟡 message content length not capped server-side beyond the 64 KB frame (frontend caps 1000) |
| Rate limiting | ➕ present but mislabeled and leaky (BUG-017/021) |

---

## 19. Architecture audit

**Strengths.**
- Textbook layering with dependency injection via small service-level interfaces (`authsvc.Repository`, `groupsvc.Repository`, …) — testable and mockable.
- Consistent error taxonomy (`repository.ErrNotFound/ErrExists/ErrNotOwner` → HTTP mapping in one place per handler package).
- Transactional multi-table operations done right (`withTx`, `AcceptGroupInvitationTx`).
- Embedded migrations mean the binary is self-contained — good for Docker.
- Frontend mirrors backend validation in one module (`lib/validate.js`) with comments cross-referencing the Go rules.

**Issues (architecture-level, not subject violations).**
1. **Handler-layer session re-resolution.** `Authorized` resolves the session, then every handler calls `common.CurrentUserID` which re-reads cookie + session from the DB (e.g. `userhandler.GetUser` does it twice). Resolve once in middleware and inject via context.
2. **Hub ↔ repository bypass.** The hub talks straight to the repository (documented), duplicating authorization logic (`CanMessage`) outside the service layer. A `chatsvc` would keep policy in one place.
3. **Dead/orphan code.** `websocket/session.go` (never called), `repository.UpdateUser` (no route), `CreateNotificationTx`, `GetPendingInvitationsForGroup` (declared in the service interface, never invoked), `UpdateGroup/DeleteGroup/DeleteUser`, comment-reaction constants. Prune or wire up.
4. **N+1 query patterns.** `ListVisiblePosts` issues 2 extra queries per post (files + reaction summary); fine at lab scale, worth batching (`WHERE post_id IN (…)`) later.
5. **Flat-text error bodies** (`http.Error`) vs JSON everywhere else — the frontend surfaces raw strings like `"authentication required"` to users. Pick one response envelope.
6. **No tests at all** (0 test files). For a subject that emphasizes migration testing, at least repository/migration smoke tests would pay off.
7. **Naming:** `follow_requests` doubles as the follow graph; `almost_private`/`private` mapping is counterintuitive (private = "selected followers"); document or rename in a future migration.
8. **Config:** hardcoded ports/paths (`:8080`, `sn.db`, `uploads`, `localhost:8080`) — env-var-ify before Docker can work (BUG-010).

---

## 20. Missing requirements

| Requirement (subject reference) | What is missing | What to implement | Affected parts |
|---|---|---|---|
| **Comments on posts, with optional image/GIF, respecting post permissions** (Posts) | Everything above the `comments` table | Comment CRUD endpoints gated by `CanViewPost`, image attach via existing `files` table (`comment_id` already reserved), `PostCard` comment section | repository, new service/handler, routes, frontend PostCard |
| **Group posts + group comments, members-only visibility** (Groups) | Everything above the `posts.group_id` column | Member-gated create/list endpoints; visibility = group membership; UI in group detail | postsvc/repository, grouphandler, routes, frontend group page |
| **Group events: title, description, day/time, Going/Not-going choice** (Groups) | Everything above `group_events`/`event_responses` tables | Event create/list for members; response endpoint (choice upsert); UI; notification on creation | new eventsvc/handler, routes, frontend, notification hookup |
| **Notification: follow request on private profile** (Notifications) | No notification emitted by followsvc | `CreateNotification` + `PublishNotification` on pending request creation | followsvc (needs hub), frontend inbox |
| **Notification: event created** (Notifications) | Events missing (above) | Same as events | eventsvc |
| **Notifications visible on every page** (Notifications) | No global UI, no unread badge, no GET endpoint | `GET /notifications` + mark-read; global bell component in `Navbar` fed by the WS already connected per tab | handler/route, Navbar, layouts |
| **Profile: user activity** (Profile) | No notion of activity anywhere | Derive from posts/comments/reactions/events, endpoint + UI section | backend, profile page |
| **Profile: followers & following lists** (Profile) | No endpoints/UI; data exists in `follow_requests` | Two read endpoints gated by `CanViewProfile` + profile UI lists | userhandler/routes, profile page |
| **Toggle own profile public/private** (Profile) | No route/handler/UI despite DB support | `PUT /me` (or privacy endpoint) + toggle control | userhandler, server.go, settings/profile pages |
| **Private post audience selection** (Posts) | No way to populate `post_visibility` | Accepted-followers picker on the composer; persist selections on create/update | postsvc, files? no — post_visibility writes, PostForm UI |
| **Listing/acting on pending follow requests** (Followers, UX-level) | No GET endpoint/inbox | `GET /follow-requests` + accept/decline inbox | userhandler, frontend |
| **Group chat room UI** (Chat) | Backend ready, no frontend | Group room view sending `group_id` messages | frontend chat module |
| **Conversation/history retrieval** (Chat) | No read endpoint for `messages` | `GET /messages?user_id|group_id` + load-on-open UI | repository, handler, frontend |
| **Avatar field present (optional) at registration** (Authentication) | Form lacks it; backend takes a string not a file | Optional file input; upload after register | register page, authhandler |

---

## 21. Partially implemented requirements

1. **Private posts ("selected followers").** *Works:* privacy enum validated; SQL visibility condition for `post_visibility` correct; reactions respect it. *Doesn't work:* nothing can insert into `post_visibility`; no UI; net effect = author-only posts. *Change:* audience picker + persistence (BUG-002).
2. **Private chat.** *Works:* real-time 1:1 delivery, persistence to DB, multi-tab, error events, emoji text. *Doesn't:* history retrieval, DM rule broader than subject (BUG-013/014), no reconnect handling. *Change:* messages GET endpoint + `CanMessage` fix.
3. **Group chat.** *Works:* entire backend path (member check, persistence, fan-out). *Missing:* any UI surface (BUG-006).
4. **Notifications.** *Works:* group-related notifications persisted + pushed live, correct recipients, distinct event type. *Doesn't:* follow-request + event notifications; every-page visibility; history; read state (BUG-016, §20).
5. **Follow system.** *Works:* request/accept/decline/unfollow API with correct authorization, auto-follow for public profiles, privacy-coupled visibility. *Doesn't:* discovery of incoming requests + the required notification (BUG-007/008) → flow not completable by a normal user.
6. **Profiles.** *Works:* info display, privacy-gated access, no password exposure. *Missing:* activity, followers/following, per-user posts endpoint, privacy toggle (BUG-020/001, §20).
7. **Registration fields.** *Works:* all required fields, bcrypt, auto-login. *Deviates:* nickname forced required (BUG-018), avatar field absent (BUG-019).
8. **Docker.** *Works:* both images build; Caddy proxy config is correct; volumes persist DB/uploads. *Doesn't:* frontend container can't reach the API (BUG-010) and WS is unreachable (BUG-009) → composed app non-functional.
9. **Migrations.** *Works:* embedded, auto-applied, ordered, reversible. *Note:* `internal/` vs subject's example `pkg/` path (explicitly permitted, but call it out at defense); schema carries unused tables from unimplemented features.

---

## 22. Extra / not required

| Feature | Location | Assessment |
|---|---|---|
| Post like/dislike reactions (+toggle, summary) | `repository/reaction.go`, `PostCard.jsx` | Harmless; nicely gated by visibility |
| Post edit/delete | `postsvc`, `PostCard.jsx` | Harmless |
| Per-IP rate limiting | `middleware/rate_limit.go` | Good instinct; fix label + leak (BUG-017/021) |
| Caddy reverse-proxy container | `compose.yml`, `caddy/` | Extra vs "two images", but it's what makes single-port serving possible; keep |
| Login by nickname | `authsvc.Login` | Harmless |
| Age ≥ 13 check | `authsvc` | Harmless |
| Group creator/invite/join response notifications | `groupsvc` | Subject says extra notifications are welcome |
| Dark-mode support | `globals.css` | Harmless |
| Char counters, modals, icon set | frontend components | Harmless polish |
| `backend/schema.sql` snapshot + extensive README/docs | repo root, `docs/` | Helpful; fix drift (BUG-024) |

---

## 23. Final compliance table

| Category | Requirement | Status | Evidence | Problem |
|---|---|---|---|---|
| Frontend | JS framework | ✅ | Next.js 16 App Router, 8 pages, layouts | — |
| Frontend | HTML/CSS/JS + responsiveness | ✅ | `globals.css` @media 860px, mobile topbar | — |
| Frontend | FE/BE communication | ✅ (dev) | `lib/api.js`, `/api/v1` rewrite, credentials | breaks in Docker (BUG-010) |
| Backend | Server, routes, middleware, error mapping | ✅ | `server.go` 31 routes, 2 middlewares | session double-lookup (arch.) |
| Backend | DB interaction | ✅ | parameterized repos, transactions | FK pragma pool issue (BUG-012) |
| Auth | Sessions/cookies | ✅ | `sessionsvc` + HttpOnly cookie | no `Secure` (BUG-023) |
| Auth | Registration fields | 🟡 | `authsvc.validateRegisterInput` | nickname required; avatar field absent (BUG-018/019) |
| Auth | Logout always available | ✅ | `Navbar` + `POST /logout` | WS not revoked (BUG-011) |
| Followers | Follow/unfollow/request/accept/decline | 🟡 | `followsvc` | no request listing/notification (BUG-007/008) |
| Followers | Public auto-follow bypass | ✅ | `Follow()` status=accepted | — |
| Profile | Info w/o password | ✅ | `common.PublicUser` | — |
| Profile | Privacy-gated viewing | ✅ | `usersvc.CanViewProfile` | — |
| Profile | Activity / followers lists | ❌ | absent | §20 |
| Profile | Public/private toggle | ❌ | column+repo exist, no route/UI | BUG-001 |
| Posts | Create w/ image or GIF | ✅ | `POST /posts` + `/files` | — |
| Posts | Public / almost-private | ✅ | `postVisibleCondition` SQL | — |
| Posts | Private (chosen followers) | 🔴 | `post_visibility` never populated | BUG-002 |
| Posts | Comments (+image) | ❌ | table only | BUG-003 |
| Groups | Create/invite/accept/refuse | ✅ | `groupsvc` + txs | re-invite after decline blocked (BUG-015) |
| Groups | Join requests, creator-only decision | ✅ | `RequestJoin/RespondJoinRequest` | — |
| Groups | Browse all groups | ✅ | `GET /groups` | — |
| Groups | Group posts/comments | ❌ | column only, filter `IS NULL` | BUG-004 |
| Groups | Events + going/not-going | ❌ | tables only | BUG-005 |
| Chat | Private messaging, real-time | ✅ | hub + `messages` | — |
| Chat | Follow-relationship rule | 🔴 | `CanMessage` public short-circuit | BUG-014 |
| Chat | Persistence/history | 🟡 | persisted, never read | BUG-013 |
| Chat | Emojis | ✅ | UTF-8 text content | no picker (cosmetic) |
| Group chat | Members-only room | 🟡 | backend complete | no UI (BUG-006) |
| Notifications | Group invite / join-request / responses | ✅ | `groupsvc.notify` | — |
| Notifications | Follow-request / event notifications | ❌ | absent | BUG-007, events |
| Notifications | Visible on every page | ❌ | no global UI/endpoint | BUG-016 |
| SQLite | Schema, FKs, constraints, indexes | ✅ | migrations 1–18 | unused tables; BUG-012 |
| Migrations | Auto-applied, up/down, ordering | ✅ | embedded golang-migrate | `internal/` vs `pkg/` path note |
| Images | JPEG/PNG/GIF, storage, auth'd serving | ✅ | `filesvc`, `CanViewFile` | register avatar gap |
| Docker | Two images build | ✅ | Dockerfiles ×2 | — |
| Docker | Containers interoperate / clean start | 🔴 | rewrite→localhost:8080; WS→:8080 | BUG-009/010 |
| WebSocket | Auth, hub, routing, pings, cleanup | ✅ | `hub.go` | BUG-011/022 |

---

## 24. Final summary

### ✅ Working correctly
Next.js frontend with real routing/responsiveness; Go backend with clean layered architecture and 31 authorized routes; DB-backed cookie sessions with bcrypt; complete group membership lifecycle (create, invite, accept/decline, join requests, creator-only decisions) with atomic transactions; group invitation/join-request/response notifications persisted and pushed live; posts with three privacy values enforced in SQL plus image/GIF upload with magic-byte validation and per-viewer file serving; SQLite schema with proper FKs, constraints and indexes; 18 ordered up/down migrations embedded and auto-applied on every start; WebSocket hub with per-user fan-out, ping/pong health, correct single-writer goroutine model; JWT-free session model exactly as the subject demands.

### Needs improvement
Private-post audience selection (SQL ready, pipeline missing); private-chat authorization scope and history; group-chat UI; notification visibility/history/read-state; follow-request discoverability; registration nickname/avatar semantics; Docker env wiring; session-context reuse in handlers; N+1 list queries; docs drift.

### Missing
Comments (posts and groups) with images; group posts; group events with Going/Not-going; follow-request and event-creation notifications; notifications on every page; profile activity; followers/following lists; profile public/private toggle; private-post audience picker; follow-request inbox; message-history and group-chat UI; avatar field in the registration form.

### Errors / Bugs
26 tracked findings — headline items: BUG-001 (no privacy toggle), BUG-002 (private posts unreachable to anyone but author), BUG-003/004/005 (comments/group posts/events absent), BUG-007/008 (follow-request flow dead-ends), BUG-009/010 (Docker FE↔BE and WS connectivity), BUG-011 (logout doesn't revoke sockets), BUG-012 (FK pragma pool), BUG-013 (no history), BUG-014 (DM rule too permissive), BUG-015 (declined group requests are permanent), BUG-016 (no notification listing).

### Extra
Reactions, post edit/delete, rate limiting, Caddy container, nickname login, age gate, dark mode, response notifications — none harmful; prune or keep deliberately.

### Security issues
No `Secure` cookie flag; WS `CheckOrigin` open (CSWSH); logout leaves sockets alive; FK enforcement unreliable across the pool; unbounded rate-limit map; no CSRF tokens (mitigated by SameSite=Lax). No SQL-injection, XSS, path-traversal, password-hash, or IDOR findings.

### Docker issues
Frontend container cannot reach the backend (`localhost:8080` rewrite); WebSocket hardcodes host port 8080; no env-var configuration; frontend container itself doesn't publish a browser-facing port (Caddy does it on 8000). Backend/frontend images and Caddy WS proxying are otherwise sound.

### Database/Migration issues
Per-connection pragmas not guaranteed pool-wide (FK enforcement); four tables and two columns exist only for unimplemented features (`comments`, `post_visibility` unpopulated, `group_events`, `event_responses`, `files.comment_id`); double rebuild of `messages` in history; declined invitation/join rows permanently block re-entry; migration folder uses `internal/` instead of the subject's example `pkg/` (permitted, but be ready to explain it).

### Overall compliance
Counted against the 40-row checklist in §1 (excluding extras):

- **Fully implemented: 19** requirements.
- **Partially implemented: 10** requirements (work exists at one layer, incomplete at another — details in §21).
- **Missing: 11** requirements (§20), concentrated in comments, group content (posts/events), notifications-on-every-page, profile followers/activity/privacy-toggle, and private-post audience selection.
- **Requirements with detected bugs: 14** rows carry at least one of the 26 tracked bugs (§17); 6 of those bugs are Critical because they nullify a subject requirement end-to-end.
- **Extra features: 10** (§22), none flagged as harmful.

Largest single leverage for compliance: implement the group-content trio (group posts → group chat UI → events) on top of the already-complete membership/authorization machinery, then close the notification loop (follow-request notification + global notification surface), then fix the two Docker wiring constants.
