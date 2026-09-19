# PROJECT_AUDIT.md

**Project:** Facebook-like social network (Go backend + Next.js frontend + SQLite)
**Audited against:** `subject.md` (official subject)
**Audit date:** 2026-09-18
**Method:** Full source read (backend, frontend, migrations, Docker), feature tracing frontend → backend → database, `go build ./...` + `go vet ./...` executed (both clean).

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
22. [Extra features](#22-extra--not-required)
23. [Final compliance table](#23-final-compliance-table)
24. [Final summary](#24-final-summary)

Legend: ✅ implemented · ⚠️ partial · ❌ missing · 🔴 implemented but buggy

---

## 1. Requirements checklist

| # | Requirement (subject) | Status | Where |
|---|---|---|---|
| 1 | JS framework used (frontend) | ✅ | Next.js 16 App Router, `frontend/` |
| 2 | HTML/CSS/JS, responsiveness, performance | ✅ | `globals.css` (mobile breakpoint), client components |
| 3 | Backend with sessions + cookies | ✅ | `sessionsvc`, `middleware/auth.go` |
| 4 | Image handling JPEG/PNG/GIF, stored files | ✅ | `filesvc`, `filehandler`, `files` table |
| 5 | WebSocket real-time | ✅/🔴 | `websocket/hub.go` (works in dev; broken behind proxy, BUG-003) |
| 6 | SQLite + migrations run at startup | ✅ | `db/sqlite/sqlite.go`, embedded migrations |
| 7 | Two Docker images (backend, frontend) | ✅ | `backend/Dockerfile`, `frontend/Dockerfile`, `compose.yml` |
| 8 | Registration: email, password, first/last name, DOB | ✅ | `authsvc.Register` |
| 9 | Registration: avatar, nickname, about-me **present but optional** | ⚠️ | Nickname is *required* (BUG-005); no avatar input in form (BUG-012/§20) |
| 10 | Stay logged in, logout available at all times | ✅ | 30-day cookie; logout in navbar |
| 11 | Follow / unfollow, accept / decline requests | ✅ backend / ⚠️ UI | `followsvc`; no UI or list endpoint for pending requests (BUG-008) |
| 12 | Public profile ⇒ auto-follow (no request) | ✅ | `followsvc.Follow` |
| 13 | Profile: all register info (minus password) | ⚠️ | `common.PublicUser` omits email + DOB (BUG-011) |
| 14 | Profile: user activity, own posts | ⚠️ | Feed filtered client-side in `profile/[id]/page.jsx` |
| 15 | Profile: followers + following lists | ❌ | No endpoint, no UI |
| 16 | Public/private profiles + visibility rules | ✅ | `usersvc.CanViewProfile`, `GET /user/{id}` 403 |
| 17 | Toggle own profile public/private | ❌ | No endpoint, no UI (BUG-007) |
| 18 | Posts: create with image or GIF | ✅ | `POST /posts` + `POST /files?post_id=` |
| 19 | Posts: public / almost-private / private | ✅/✅/🔴 | `private` posts broken — `post_visibility` never written (BUG-001) |
| 20 | Comments on posts (with image/GIF) | ❌ | No endpoints, no UI (table exists) |
| 21 | Groups: create, invite, accept/refuse | ✅ | `groupsvc` (12 endpoints) |
| 22 | Members can invite; join requests; creator accepts/refuses | ✅ | `groupsvc` |
| 23 | Browse all groups section | ✅ backend / ❌ UI | `GET /groups` exists; `groups/page.jsx` is a "coming soon" placeholder |
| 24 | Group posts + comments, member-only visibility | ❌ | `posts.group_id` never set; no endpoints/UI |
| 25 | Group events (title/desc/date, going/not-going) | ❌ | Tables exist; zero code uses them |
| 26 | Private chat with follow-relationship rule | ✅ | `repository.CanMessage` |
| 27 | Real-time delivery over WebSocket | 🔴 | Works only when backend is on port 8080 of the hostname (BUG-003) |
| 28 | Emojis in chat | ⚠️ | Unicode text passes through; no emoji picker |
| 29 | Group chat room (members only) | ⚠️ | Backend branch complete; no UI, no history endpoint |
| 30 | Notifications visible on **every page** | ❌ | Only `/notifications` page, realtime-only; no navbar badge |
| 31 | Notification: follow request (private profile) | ❌ | Nothing creates it (BUG-006) |
| 32 | Notification: group invitation | ✅ | `groupsvc.notify` → `notifications` + WS push |
| 33 | Notification: join request to group creator | ✅ | Same |
| 34 | Notification: event created for members | ❌ | No events exist |
| 35 | Notifications separate from private messages | ✅ | Distinct `{"type":"notification"}` WS event |

---

## 2. Frontend audit

**Framework:** Next.js (`^16.3.4`) with React 19, App Router, `frontend/package.json`. Genuinely used — routing (`app/`), layouts (`(auth)`, `(main)` route groups), client components (`'use client'`), `next/link`, `next/navigation`. This is a real framework usage, not a library.

**Pages found:**

| Route | File | State |
|---|---|---|
| `/` | `app/page.jsx` | redirect to `/home` |
| `/login`, `/register` | `app/(auth)/…` | ✅ complete, validated |
| `/home` | `app/(main)/home/page.jsx` | feed + composer |
| `/profile/[id]` | `app/(main)/profile/[id]/page.jsx` | profile + posts (filtered client-side) |
| `/people` | `app/(main)/people/page.jsx` | client-side search over feed authors |
| `/groups` | `app/(main)/groups/page.jsx` | **static "Coming soon" placeholder — no functionality** |
| `/notifications` | `app/(main)/notifications/page.jsx` | realtime-only list, no history |
| `/settings` | `app/(main)/settings/page.jsx` | avatar change; account details read-only |
| `/chat`, `/chat/[id]` | `app/(main)/chat/…` | private chat (live-only) |
| 404 | `app/not-found.jsx` | ✅ |

**Communication:** `lib/api.js` centralizes fetch with `credentials: 'include'`; `/api/v1` prefix. In dev, `next.config.js` rewrites `/api/v1/*` → `http://localhost:8080/*`. In Docker, Caddy routes `/api/v1/*` → `backend:8080`.

**Responsiveness:** `globals.css` has a real `@media (max-width: 860px)` block (sidebar becomes a top bar, grids collapse). ✅

**Navigation:** `components/Navbar.jsx` sidebar with active-link highlighting, logout button. ✅

**Performance / obvious issues**

- `/home` fetches *all* posts then renders all — no pagination anywhere (backend admits this in `backend/readme.md`).
- `people/page.jsx` and `chat/page.jsx` duplicate the same "derive people from feed posts" logic.
- Profile page derives "user's posts" by fetching the whole feed and filtering — an N-posts endpoint would be correct.
- Two pages open **two separate WebSockets** (`chat/[id]` and `notifications`); no shared socket or reconnect logic (the docs/WEBSOCKET-PROTOCOL.md describes reconnection with backoff — **not implemented**).

**Missing UI for existing backend endpoints:** group browsing/creation/invitations/join-requests, accept/decline follow requests, public/private profile toggle, group chat, events, selected-followers picker, notification history/badge.

---

## 3. Backend audit

**Server & structure:** `cmd/server/main.go` → `http.ListenAndServe(":8080", middleware.RateLimit(mux))`, routes in `internal/server/server.go`. Layered: `middleware → handler → service → repository → sqlite`, with the WS hub beside it (`backend/readme.md` matches reality).

**Routes (all registered, verified against handlers):**

- Auth: `POST /register` (Guest), `POST /login` (Guest), `POST /logout`, `GET /me`
- Users: `GET /user/{id}`, `POST|DELETE /users/{id}/follow`, `POST /follow-requests/{id}/accept|decline`
- Posts: `GET|POST /posts`, `PUT|DELETE /posts/{id}`, `POST|DELETE /posts/{id}/reactions`
- Files: `POST /files`, `POST /avatar`, `GET /fs/{id}`
- Groups: 12 routes (create, list, detail, members, invitations, join-requests + responses)
- WS: `GET /ws`

**Middleware:** rate limiting (per-IP window) + `Authorized`/`Guest` session checks. Note: rate limit constant is **1000/min** (`rate_limit.go`), the `X-RateLimit-Limit` header says **100**, and the readme says 100/min — inconsistent (BUG-013).

**Error handling:** consistent `http.Error` flat-text style; handlers map sentinel errors → 400/401/403/404/409/500. Group flows are the most rigorous (errors re-checked inside transactions).

**Authorization:** every handler resolves the user from the session cookie, never from the body. Object-level checks verified: post update/delete are `WHERE author_id = ?`; invitation/join-request responses re-verify ownership inside the DB transaction; group membership/creator checks present; file serving has a visibility query.

**Known gaps:** no comments endpoints, no group posts, no events, no notification list/read endpoints, no "list my pending follow requests", no profile-update/privacy-toggle endpoint, no message-history endpoint. See §20.

`go build ./...` and `go vet ./...` pass.

---

## 4. Authentication audit

| Item | Status | Evidence |
|---|---|---|
| Registration required fields (email, password, first/last, DOB) | ✅ | `authsvc.validateRegisterInput` (regexes, 13+ age check) |
| Nickname optional | ❌ | **Required** by backend and frontend — contradicts subject (BUG-005) |
| Avatar/Nickname/About-me "present but skippable" | ⚠️ | About-me ✅ optional; nickname forced; **avatar field absent from the register form** (only `POST /avatar` after login) |
| Password handling | ✅ | bcrypt `GenerateFromPassword`/`CompareHashAndPassword`, hash never serialized (responses use `common.PublicUser`/`PrivateUser` maps) |
| Login | ✅ | By email **or** nickname; generic "invalid email or password" (no enumeration) |
| Sessions | ✅ | 32 random bytes (`crypto/rand`), stored in `sessions`, 30-day expiry, expiry checked on read |
| Cookies | ✅ | `HttpOnly`, `SameSite=Lax`, `Path=/` (no `Secure` — fine for the HTTP-only deployment, would need it for HTTPS) |
| Logout | ✅/🔴 | Deletes session, clears cookie; **but does not actually close the session's WebSockets** — `RevokeSessionClients` iterates a map that is never populated (BUG-004), contradicting the readme |
| Stay logged in | ✅ | Cookie + session; `(main)/layout.jsx` gate via `GET /me` |
| `Guest` middleware | ✅ | Logged-in users cannot re-register/login (403) |

Note: register responds `201` and immediately creates a session — good UX, not required by subject.

---

## 5. Followers audit

**Backend flow (`followsvc/service.go`, `repository/follow.go`):**

- `POST /users/{id}/follow` → if target `private = 0` → row inserted with `status='accepted'` (auto-follow ✅); if private → `status='pending'`.
- Re-follow after decline re-activates the row; duplicate pending/accepted → 409.
- `POST /follow-requests/{id}/accept|decline` → recipient-only check (`follow.ToUserID != recipient → 403`), only while pending.
- `DELETE /users/{id}/follow` → deletes the row. Unfollow without following is a silent 204 (harmless).
- `IsFollowing` (accepted only) is used for private-profile access and almost-private posts. Correct directionality everywhere (`from = viewer, to = author`).

**Gaps:**

- **No endpoint to list pending follow requests received by the current user** — the accept/decline endpoints are unusable from any UI because nobody can learn the request IDs (BUG-008).
- **No notification when a follow request arrives** — explicitly required by the subject (BUG-006).
- **No UI** for accept/decline (frontend only has Follow/Unfollow buttons on the profile page).
- The readme's sequence diagram claims the WS hub notifies the target on follow — **not implemented** anywhere in `followsvc`.

---

## 6. Profile audit

- `GET /user/{id}` (`userhandler.GetUser`): returns `common.PublicUser` (id, first/last, avatar, nickname, about_me, private, created_at). **403 "profile is private"** when target is private and viewer doesn't follow — correct privacy rule, and private-profile 403 is what the frontend uses to render the locked card.
- **Privacy leak check:** email, DOB, password never included in `PublicUser` ✅. Password hash also never serialized anywhere (models have JSON tags but responses are hand-built maps). ✅
- **BUG-011 (Medium):** the subject says the profile shows *every* piece of register info except password — email and date of birth are missing from `GET /user/{id}` even for the owner (owner's own email only appears on `/me`).
- **User activity / posts:** the frontend filters the whole feed by `author_id` — works for posts the viewer may see, but there is no dedicated "posts of user X respecting privacy" endpoint; a visitor who can see the profile sees only posts that also pass feed visibility. Acceptable approximation, marked partial.
- **Followers / following lists:** ❌ nothing exists (subject requires displaying them).
- **Toggle public/private:** ❌ `repository.UpdateUser` writes the `private` column but the only caller is `SetAvatar`; there is no endpoint and no settings UI (the settings page literally says "Editing these needs an API endpoint that does not exist yet").

---

## 7. Posts audit

- **Create:** `POST /posts` (form `content`, `privacy`) → 201. Privacy validated against the three constants. Content length is *not* validated server-side (frontend caps 1000) — BUG-017.
- **Images:** created separately via `POST /files` with `post_id`; only the post author may attach (`filesvc.Upload` checks `post.AuthorID != ownerID`); max 3 images/10 MB; served by `GET /fs/{id}` with per-viewer visibility. `Post.Images` hydrated from the `files` table. ✅
- **public** — visible to all authenticated users (`ListVisiblePosts` condition). ✅
- **almost_private** — visible when viewer follows the author (accepted follow). ✅ Correct direction.
- **private** ("only the followers chosen by the creator") — 🔴 **BUG-001 (Critical):** visibility depends on rows in `post_visibility`, but **no code path ever inserts into that table**. There is no API to declare chosen followers and no UI to pick them. A `private` post is therefore visible only to its author, forever. The subject requirement is not satisfied.
- **Update/Delete:** owner-only via `WHERE id = ? AND author_id = ?`, `RowsAffected` → 404. ✅
- **Reactions:** extra feature (see §22), correctly gated by `CanViewPost`.
- **Comments:** ❌ entirely missing — no handler/service/repository code, no endpoints, no UI. Only the `comments` table and the unused `files.comment_id` column exist. The subject explicitly requires creating comments with image/GIF.
- **Group posts:** ❌ `posts.group_id` exists and `ListVisiblePosts` correctly excludes group posts (`WHERE p.group_id IS NULL`), but nothing can ever create one — `postsvc.Create` has no group parameter and no route accepts `group_id`.

---

## 8. Groups audit

The backend is the strongest part of the project (recently built; see `GROUPS_BACKEND_IMPLEMENTATION.md`, which matches the code):

- **Create** (`POST /groups`): title ≤100 required, description ≤1000 optional; creator auto-inserted into `group_members`. ✅
- **Browse** (`GET /groups`): all groups + `member_count`, `is_member`, `pending_join`, `is_creator`. Satisfies the subject's "browse through all groups" — **but only in the API; the groups page is a placeholder**.
- **Detail** (`GET /groups/{id}`): members+creator see everything; invited/pending users get flags; everyone else 404 (anti-enumeration). Creator is projected through `GroupCreator` — no email/DOB/password leak. ✅
- **Invitations**: members invite non-members; self-invite 400, unknown user 404, duplicate 409 (pre-check + `UNIQUE(group_id,to_user_id)` + `isUnique()` race mapping). Recipient-only accept/decline, re-validated **inside a transaction** (`AcceptGroupInvitationTx`), membership insert + status update atomic. ✅
- **Join requests**: non-members only (creator can't request own group); creator-only pending list; creator-only accept/decline with the same transactional re-validation. ✅
- **Notifications** on invite / join-request / accept — persisted to `notifications` and pushed live (`hub.PublishNotification`). ✅ (types: `group_invitation`, `group_join_request`, `group_invite_response`, `group_join_response`)
- **Group posts/comments**: ❌ missing (see §7).
- **Group events**: ❌ missing — `group_events` and `event_responses` tables (migration 000006) have **zero** references in Go code. Title/description/date-time/going/not-going are all unimplemented.
- Known product quirk: after a decline, the `UNIQUE` pair constraint prevents a new invitation/request for the same (group,user) — permanent 409 (BUG-016).

---

## 9. Private chat audit

- **Eligibility rule** (subject: "at least one of the users must be following the other", plus the public-profile exception): `repository.CanMessage` allows private messages when recipient `private = 0` **or** an accepted follow exists in either direction. Matches the subject exactly. ✅
- **Real-time:** sender's message is persisted (`CreateMessage`) and pushed to recipient **and** echoed to sender (`hub.publish`). Multiple tabs per user supported (per-user client map). ✅ in dev.
- **Delivery in the real deployment:** 🔴 BUG-003 — both chat pages hardcode `ws://${window.location.hostname}:8080/api/v1/ws`. Behind Caddy (host port **8000**) or in Docker (8080 not published), the socket cannot connect. Chat and live notifications only work when the backend is reachable at `hostname:8080` (dev).
- **Persistence/conversations:** messages are stored in `messages`, but there is **no endpoint to fetch history** and the UI says "Messages are live only and are not saved when you reload." Data is saved yet unreachable — partial.
- **Emojis:** no picker; UTF-8 emoji text survives the JSON round-trip, so emojis *work* but the feature is minimal. ⚠️
- **Authorization on send:** enforced server-side (`CanMessage` before persist); failures produce `{"type":"error"}`. ✅

---

## 10. Group chat audit

- **Backend:** the WS hub accepts `{"type":"message","group_id":X}`, checks membership via `CanMessage`, persists, and fans out to `GroupMemberIDs`. ✅ backend logic complete.
- **Frontend:** ❌ no group chat UI exists (groups page is a placeholder), and there is no message-history endpoint, so even a hand-crafted client couldn't read the room's past messages.
- Net status: **not usable end-to-end**; backend ready for a follow-up.

---

## 11. Notifications audit

- **Created today:** group invitations, join requests, and the two "response" variants (`groupsvc.notify` → `repository.CreateNotification` → `hub.PublishNotification`). Content strings are human-readable; `group_id` attached; failures logged, never roll back the group operation. ✅
- **Not created:** follow-request notifications (subject-required ❌), event notifications (no events ❌).
- **Visibility on every page:** ❌ The only consumer is `/notifications`, which builds its list **only from live WS events received while the page is open**. There is no `GET /notifications` endpoint, no unread badge on the navbar, and the `read` column is never updated. A notification that arrives while you're on another page is invisible.
- **Separation from private messages:** ✅ distinct `{"type":"notification", ...}` vs `{"type":"message", ...}` WS events, distinct UI treatment.
- `docs/WEBSOCKET-PROTOCOL.md` documents `follow_request`/`follow_accepted`/`group_event` notification types — **documentation only, not implemented**.

---

## 12. SQLite audit

**Tables (16 app tables + `schema_migrations`):**

```
users (id PK, email UNIQUE, password, first_name, last_name, date_of_birth,
       avatar, nickname, about_me, private, created_at)
 ├─< sessions (id PK, user_id FK→users CASCADE, expires_at)
 ├─< posts (author_id FK→users CASCADE, group_id FK→groups CASCADE NULL, privacy, content)
 │    ├─< post_visibility (PK(post_id,user_id), FKs CASCADE)          ← never written (BUG-001)
 │    ├─< comments (post_id FK, author_id FK)                          ← unused by app
 │    └─< files (id TEXT PK, owner FK CASCADE, post_id FK SET NULL,
 │              comment_id FK SET NULL, message_id FK SET NULL)
 ├─< follow_requests (from,to FKs CASCADE, status, UNIQUE(from,to))
 ├─< groups (creator_id FK→users CASCADE, title, description)
 │    ├─< group_members (PK(group_id,user_id), FKs CASCADE)
 │    ├─< group_invitations (FKs CASCADE, status, UNIQUE(group_id,to_user_id))
 │    ├─< group_join_requests (FKs CASCADE, status, UNIQUE(group_id,user_id))
 │    ├─< group_events (FKs CASCADE)                                   ← unused by app
 │    │    └─< event_responses (UNIQUE(event_id,user_id))              ← unused by app
 │    └─< notifications (user_id FK, actor_id FK, group_id FK SET NULL, read)
 └─< messages (from FK, to FK NULL, group_id FK NULL,
               CHECK (from_user_id <> to_user_id OR to_user_id IS NULL))
reactions (target_type CHECK post|comment, target_id, user_id FK, reaction CHECK like|dislike,
           UNIQUE(target_type,target_id,user_id))   ← no FK on target_id (polymorphic)
```

**Good:** composite PK on `group_members`; `UNIQUE` pairs on invitations/join-requests/follows; indexes on all FK lookup paths (`idx_follow_requests_to/from`, `idx_posts_author/group`, `idx_notifications_user`, `idx_messages_to/group`, …); messages CHECK constraint fixed in 000017 to allow group messages; FKs with sensible CASCADE / SET NULL.

**Issues:**

- 🔴 **BUG-002:** `PRAGMA foreign_keys = ON` is per-connection in SQLite, but `sqlite.Open` executes it **once** on one pooled connection (`db.Exec`). Other pool connections run with FK enforcement off. Same for `busy_timeout`. (`journal_mode=WAL` is persistent, so that one is fine.) Cascades and FK-based integrity are therefore unreliable.
- `users.nickname` has **no UNIQUE constraint** and registration never checks nickname collisions, while login accepts nickname as identifier → ambiguous login (BUG-010).
- Unused-but-present: `comments`, `group_events`, `event_responses`, `post_visibility`, `files.comment_id` — schema exists for missing features.
- `reactions.target_id` is polymorphic without FK (acceptable pattern, worth knowing).

---

## 13. Migration audit

- **System:** `golang-migrate` + `iofs` source; **migrations are embedded** (`//go:embed sqlite/*.sql`) and applied in `sqlite.Migrate()` at every startup (`main.go → sqlite.InitDB`). A clean database gets all 18 versions. ✅
- **Structure:** `backend/internal/db/migrations/sqlite/000NNN_name.{up,down}.sql` — 18 up + 18 down, all pairs present. The subject's suggested layout is `backend/pkg/db/...` but explicitly allows "organized as you wish"; the applied path is internal, which satisfies "the application of migrations and the file organization will be tested" as long as the runner works — and it runs from code, not from a filesystem path.
- **Order:** 000001 users → 000002 sessions → 000003 posts/comments/post_visibility → 000004 follows → 000005 groups → 000006 events → 000007 notifications/messages → 000008 messages rebuild → 000009 reactions → 000010 files → 000011 files.message_id → 000012/13 URL-prefix data fixes (still-safe column present) → 00014 drop legacy `image` columns → 00015 avatar prefix strip → 00016 add `posts.type` → 00017 messages CHECK rebuild → 00018 drop `posts.type`. Data migrations precede the drops that would break them. ✅
- **Schema matches `schema.sql` snapshot** (reference file, not executed).
- Verification note: migration *execution* against a clean DB was not run during this audit (would require creating a DB); correctness is established by code reading + the successful build. **Actual clean-database run cannot be verified from the provided source code alone.**

---

## 14. Images audit

- **Types:** content sniffed via `http.DetectContentType` on the first 512 bytes; only `image/jpeg`, `image/png`, `image/gif` accepted — subject's three types covered, extension never trusted. ✅
- **Size/count:** 10 MB per image (`MaxBytesReader` + `LimitReader`), max 3 per request. ✅
- **Storage:** random 32-hex ID → `uploads/<id>` (`0o750` dir, `0o640` file, `O_EXCL`); metadata in `files` with `storage_path` hidden from JSON (`json:"-"`). Original filename stored via `filepath.Base` — **no path traversal** (user input never used to build the path). ✅
- **Retrieval:** `GET /fs/{id}` is authenticated and runs `CanViewFile` (owner, avatar references, post visibility, message participants, group members) → 404 otherwise; `Cache-Control: private, immutable` is safe for content-addressed IDs. ✅
- **Cleanup:** on DB insert failure the file is removed; on write failure the partial file is removed. ✅
- **Gaps:** no re-encode/sanitization (subject doesn't require it); avatar at **registration** not possible (form field missing) — only `POST /avatar` after login; register still accepts a legacy `avatar` *string* form field that is stored verbatim in `users.avatar` without validation (BUG-012).

---

## 15. Docker audit

- **Images:** `backend/Dockerfile` — two-stage `golang:1.25-alpine` with `CGO_ENABLED=1` (needed by `go-sqlite3`), runtime `alpine:3.22` + `libgcc`/`ca-certificates`. `frontend/Dockerfile` — `node:22-alpine`, `npm ci`, `next build`, `next start -p 3000`. Both look correct. ✅
- **compose.yml:** backend (`expose 8080`), frontend (`expose 3000`), Caddy (published `8000:80`, `8443:443`) with `Caddyfile.docker` routing `/api/v1/*` → `backend:8080` and `/` → `frontend:3000` (prefix stripped via `handle_path`, matching the Go mux). Volumes for DB and uploads. Subject's "two images" satisfied; Caddy is a legitimate third service (the subject itself suggests Caddy).
- **Database in containers:** `sn.db` is created in the working directory `/app`, persisted through the `backend-data` named volume; uploads through `backend-uploads`.
  - ⚠️ BUG-018: `backend-data` is mounted **over `/app`** — the volume shadows the image's `/app` (works only because Docker copies image content into an empty named volume on first use; a non-empty or reused volume would hide `/server`). Mounting `/app/data` would be robust.
- **Communication:** HTTP works through Caddy. **WebSocket does not**: the frontend connects to `ws://hostname:8080` (BUG-003), and neither 8080 nor 3000 is published to the host; in the composed deployment, chat and live notifications fail.
- `next.config.js` rewrite to `http://localhost:8080` is dead in Docker (harmless; Caddy owns `/api/v1`).
- No healthchecks; `depends_on` only orders startup (migrations run fast, so acceptable).
- **A real clean-environment container start was not executed during this audit** ("cannot be verified from the provided source code"); the assessment is config-level.

---

## 16. WebSocket audit

- **Upgrade/auth:** `GET /ws` behind `auth.Authorized`, then the hub re-reads the session cookie before `Upgrade` (gorilla). Unauthenticated → 401 before upgrade. ✅
- **Client registry:** `Hub.clients map[userID]map[*Client]struct{}` guarded by `sync.RWMutex`; `add`/`remove` are locked; supports **multiple tabs/devices per user**. ✅
- **Concurrency:** one goroutine per client for `writePump`; `readPump` runs in the request goroutine; sends are non-blocking `select`/`default` on a buffered chan (16). No obvious races; `go vet` clean.
- **Health:** write-side ping every 45s, 60s read deadline refreshed by pong handler, 64 KB read limit, 10s write deadline. ✅
- **Cleanup:** on read error the client is removed and the connection closed; `writePump` exit closes the socket too. ✅
- **🔴 BUG-004:** `websocket/session.go` (`trackClient`/`untrackClient`/`RevokeSessionClients`) is **dead code** — `ServeHTTP` never calls `trackClient` and `readPump`'s defer never calls `untrackClient`. Consequence: `POST /logout` deletes the session, but already-open sockets stay authenticated for as long as the TCP connection lives (the hub doesn't re-check the session per message), and the readme's claim "force closes that session's websockets" is false.
- **Message drops:** when a client's send buffer is full the payload is silently discarded (no close, no error) — low severity (BUG-015).
- **`CheckOrigin: func(*http.Request) bool { return true }`** — accepts any origin. Cross-Site WebSocket Hijacking is mitigated in practice by the `SameSite=Lax` cookie (browsers don't attach it to cross-site WS handshakes), but the permissive origin check is still worth tightening (see §18).

---

## 17. Bug audit

| ID | Severity | Location | Problem / Why / Expected vs current / Fix |
|---|---|---|---|
| **BUG-001** | **Critical** | `repository/post.go` (`postVisibleCondition`), `postsvc.Create`, `PostForm.jsx` | **`private` (selected-followers) posts can never be shared.** Visibility requires `post_visibility` rows, but nothing ever inserts there and no API/UI lets the author pick followers. Expected: chosen followers see the post. Current: only the author ever sees it. Fix: accept a list of allowed user IDs on post creation (validate they follow the author), insert into `post_visibility`, and add the follower-picker UI. |
| **BUG-002** | High | `db/sqlite/sqlite.go` `Open()` | **FK enforcement unreliable.** `PRAGMA foreign_keys`/`busy_timeout` are per-connection but executed once on one pooled connection. Expected: FKs enforced on every connection. Fix: set them in the DSN (`file:sn.db?_foreign_keys=on&_busy_timeout=5000`) or via a `ConnectHook`. |
| **BUG-003** | High | `chat/[id]/page.jsx:29`, `notifications/page.jsx:13` | **Hardcoded `ws://hostname:8080/api/v1/ws`.** Breaks behind Caddy (host port 8000) and in Docker (8080 unpublished). Expected: WS through the same origin/proxy. Fix: derive scheme/host from `window.location` (`wss?://location.host/api/v1/ws`). |
| **BUG-004** | High | `websocket/session.go`, `hub.go ServeHTTP`, `authhandler.Logout` | **Session-revocation dead code.** `trackClient`/`untrackClient` are never called, so `RevokeSessionClients` is a no-op; logout leaves sockets live. Expected: logout closes that session's sockets (readme claims it). Fix: call `trackClient(sessionID, client)` after `add`, `untrackClient` in `readPump`'s defer. |
| **BUG-005** | Medium | `authsvc.validateRegisterInput`, `register/page.jsx` | **Nickname required** although the subject marks it optional. Fix: make nickname optional backend+frontend (keep uniqueness when provided). |
| **BUG-006** | Medium | `followsvc` | **No notification on follow request** (subject-required; `docs/WEBSOCKET-PROTOCOL.md` even documents the type). Fix: create + push a `follow_request` notification in `Follow` when status is pending. |
| **BUG-007** | Medium | `usersvc`/`userhandler`, `settings/page.jsx` | **No way to toggle profile public/private** (subject-required). `UpdateUser` supports it but no endpoint/UI exists. Fix: `PATCH /me` or `POST /users/me/privacy` + settings toggle. |
| **BUG-008** | Medium | server.go, `followsvc` | **No endpoint listing my pending follow requests**, so accept/decline routes are unreachable from any client. Fix: `GET /follow-requests` (recipient-scoped) + UI. |
| **BUG-009** | Medium | notifications page, Navbar | **Notifications not visible on every page**: no history endpoint, no navbar badge/counter; events missed while off-page are lost to the user. Fix: `GET /notifications` (+ mark-read), global listener in the `(main)` layout with a badge. |
| **BUG-010** | Medium | migration 000001, `authsvc` | **Nickname not unique** but used as a login identifier → first-match login, ambiguous accounts. Fix: `UNIQUE` on `users.nickname` (new migration) + registration check. |
| **BUG-011** | Medium | `common/response.go` `PublicUser`, `userhandler.GetUser` | **Profile omits email and date of birth**, though the subject says the profile carries every register field except password. Fix: include email/DOB when viewer is owner or an accepted follower. |
| **BUG-012** | Low | `authhandler.Register`, `authsvc.RegisterInput.Avatar` | Legacy `avatar` form string is stored verbatim in `users.avatar` with no validation (garbage breaks `<img>` URLs). Fix: ignore the field or accept an uploaded file ID only. |
| **BUG-013** | Low | `middleware/rate_limit.go`, `backend/readme.md` | Rate limit is 1000/min in code, `X-RateLimit-Limit: 100` in header, "100/min" in readme. Fix: pick one value; align header + docs. |
| **BUG-014** | Low | `middleware/rate_limit.go` | Client map never pruned → unbounded memory growth over time. Fix: evict stale windows. |
| **BUG-015** | Low | `hub.go publish/sendError` | Full send buffer ⇒ silent message drop (no close/error). Fix: close slow clients or log. |
| **BUG-016** | Low | `group_invitations`/`group_join_requests` UNIQUE pairs | After a decline, the same (group,user) pair can never invite/request again (permanent 409). Flagged as product decision in `GROUPS_BACKEND_IMPLEMENTATION.md`; fix would be status-reset or new-row schema change. |
| **BUG-017** | Low | `postsvc.Create/Update` | Post content length unvalidated server-side (frontend caps 1000; API accepts arbitrary length). Fix: mirror the 1000-char cap. |
| **BUG-018** | Low | `compose.yml` | `backend-data` volume mounted over `/app` shadows the image workdir (works only via Docker's first-use copy). Fix: mount a subdirectory (e.g. `/app/data`). |
| **BUG-019** | Low | `sessionsvc.Get` | Expired sessions are deleted only when touched; dead rows accumulate. Fix: periodic cleanup or delete-on-expiry sweep. |

**Compile errors:** none (`go build`, `go vet` clean). **SQL injection:** none found — every query is parameterized. **Runtime verification** of server behavior beyond the code (live requests, container start) was not performed in this audit: *"Cannot be verified from the provided source code."*

---

## 18. Security audit

| Area | Finding |
|---|---|
| Password hashing | ✅ bcrypt, default cost; hashes never leave the server (hand-built JSON maps). |
| Sessions | ✅ 32-byte CSPRNG tokens; server-side store; expiry enforced. ⚠️ Cookie lacks `Secure` (fine for the HTTP-only Caddy setup; required for HTTPS). |
| Cookies | ✅ `HttpOnly`, `SameSite=Lax` — Lax also blocks the cookie on cross-site WS handshakes and most cross-site posts (de-facto CSRF mitigation). No CSRF token (acceptable given Lax + JSON/form APIs; note it). |
| Authorization / IDOR | ✅ Verified: post update/delete owner-scoped in SQL; follow responses recipient-checked; group invitation/join responses re-checked **inside transactions**; group member/creator checks on every group route; files visibility-checked per viewer; caller ID always from the session cookie. |
| SQL injection | ✅ All queries parameterized (`?`), no string-built values. |
| XSS | ✅ React escapes by default; no `dangerouslySetInnerHTML` anywhere. |
| Uploads | ✅ Magic-byte sniffing (extension ignored), size caps, random storage names, no path traversal, `json:"-"` hides storage paths. |
| WebSocket | ⚠️ `CheckOrigin` always true — currently mitigated by the Lax cookie; tighten to expected host(s). Auth enforced before upgrade and per message (via `CanMessage`). |
| Information exposure | ✅ Password/email/DOB never in public payloads; group detail projects creator through a safe struct; 404-instead-of-403 for hidden groups. ⚠️ Register returns 409 for taken emails (standard, minor enumeration). |
| Rate limiting | ✅ present (per-IP). ⚠️ Header/value/doc mismatch + unbounded map (BUG-013/014). |
| Secrets | ✅ None hardcoded; no `.env` committed (`.gitignore`). |
| Misc | ⚠️ `POST /register` accepts a raw `avatar` string into `users.avatar` (BUG-012). ⚠️ No input length caps on some backend fields (posts BUG-017; group fields *are* capped). |

---

## 19. Architecture audit

**Strengths**

- Clean layering (`model / middleware / handler / service / repository / db`) with dependency injection composed in one place (`handler/handlers.go`). Interfaces are consumer-defined (`authsvc.Repository`, `groupsvc.Repository`…) — good Go style.
- The groups vertical is exemplary: transactions in the repository, sentinel-error mapping, race-safe duplicate handling, documented (`GROUPS_BACKEND_IMPLEMENTATION.md`).
- Frontend mirrors the backend limits in one module (`lib/validate.js`) — nice consistency.
- Documentation (`backend/readme.md`, `docs/`, `schema.sql`) is unusually complete and mostly accurate — except the follow-notification and logout-WS claims (see BUG-004/006).

**Issues (not subject requirements)**

1. **Dead code:** `websocket/session.go` entirely unused; `repository.isUniqueConstraint`, `CreateNotificationTx`, `UpdateGroupInvitationStatus`, `UpdateGroupJoinRequestStatus`, `GetPendingInvitationsForGroup`, `UpdateGroup`, `DeleteGroup`, `DeleteUser` have no callers. `groupsvc.isUnique` duplicates `repository.isUniqueConstraint`.
2. **Repo directly in WS hub:** the hub bypasses the service layer (documented) — pragmatic, but it means chat has no validation/service home for future rules (length caps live only in the frontend).
3. **Duplication:** frontend `people/page.jsx` and `chat/page.jsx` duplicate the feed-derivation logic; backend has parallel follow/invitation/join-request service shapes (acceptable).
4. **No pagination** anywhere (`GET /posts`, `GET /groups`, member lists).
5. **No tests committed:** `.gitignore` excludes `*_test.go`, so the "tests" claimed in `CHANGELOG.md` are not in the repo; verification relies on manual/integration sessions.
6. **Mixed response styles:** handlers return hand-built maps (`common.PublicUser`) instead of model JSON tags — safe but easy to drift (already caused the email/DOB omission).
7. Naming and file organization are consistent; no circular imports; `internal/` layout follows Go conventions.

---

## 20. Missing requirements

| Requirement | Subject reference | What's missing | Affected parts |
|---|---|---|---|
| Comments on posts (with image/GIF) | "create posts **and comments** on already created posts… can include an image or GIF" | Entire feature: no endpoints, service, repo code, UI. Only the `comments` table exists. | Backend posts, frontend PostCard, files (`comment_id` unused) |
| Private posts with selected followers | "private (only the followers chosen by the creator)" | API to declare chosen followers + picker UI; `post_visibility` never written | `postsvc`, new endpoint, `PostForm.jsx`, DB writes |
| Group posts + group comments | "in a group a user can create posts and comment… only displayed to members" | No creation path (`posts.group_id` never set), no member-scoped endpoints/UI | `postsvc`, grouphandler, groups UI |
| Group events | "create an event… title, description, day/time, going / not going" | Entire feature; tables exist unused | New service/handler/routes, groups UI, `group_events`/`event_responses` |
| Group UI (browse/create/invite/join/chat) | whole Groups section | Backend is done; the page is a "coming soon" placeholder | `frontend/app/(main)/groups/` |
| Follow-request notification | "notified if… some other user sends him/her a following request" | No creation anywhere | `followsvc`, notification fan-out |
| Event-creation notification | "notified if… an event is created" | Blocked on events existing | Events + `groupsvc`/eventsvc |
| Notifications on every page | "see the notifications in every page" | No history endpoint, no global badge/listener; page-only realtime | `(main)/layout.jsx`, Navbar, new `GET /notifications` |
| Toggle profile public/private | "option that allows the user to turn its profile public or private" | No endpoint, no UI | `usersvc`/`userhandler`, settings page |
| Followers/following lists on profile | "display the users that are following… and who he/she is following" | No endpoints, no UI | `followsvc`, profile page |
| Accept/decline follow requests (usable) | "recipient can choose to accept or decline" | Backend respond endpoints exist but no list endpoint + no UI | `GET /follow-requests`, frontend |
| Avatar in the registration form | "Avatar/Image (Optional)… should be present in the form" | Field absent; only post-login `POST /avatar` | `register/page.jsx` (multipart submit) |
| Emojis in chat (picker-level) | "send emojis to each other" | Unicode text works; no picker UI | `chat/[id]/page.jsx` |
| Message history / conversations | implied by chat + persistence already built | `messages` saved but never fetchable | `GET /messages` endpoint, chat UI load |

---

## 21. Partially implemented requirements

1. **Private posts (selected followers)** — What works: the three-level privacy model, `private` level validation, and the visibility *query* (which checks `post_visibility`). What doesn't: nothing can ever populate `post_visibility`; no selection UI. Fix: see BUG-001.
2. **Real-time chat delivery** — What works: full hub, auth, persistence, echo, group fan-out (in dev against `hostname:8080`). What doesn't: unreachable behind the Caddy port mapping / in Docker (BUG-003).
3. **Notifications** — What works: group invite/join-request/response notifications, persisted + pushed, distinct message shape. What doesn't: follow-request notifications, history endpoint, read/unread, global visibility (BUG-006, BUG-009).
4. **Logout WebSocket revocation** — What works: session deletion, cookie clearing. What doesn't: sockets of that session stay open (BUG-004).
5. **Registration form** — Works: required fields, DOB 13+, about-me optional. Doesn't: nickname forced (BUG-005), avatar field missing (BUG-012/§20).
6. **Profile information** — Works: public/private gating, 403 for hidden profiles, safe field projection. Doesn't: email/DOB omitted (BUG-011), followers/following lists missing, no privacy toggle, "activity" approximated by feed filtering.
7. **Follow system** — Works: follow/unfollow, auto-accept for public, recipient-only accept/decline with duplicate handling. Doesn't: no pending-request list endpoint/UI (BUG-008), no notification (BUG-006).
8. **Private chat** — Works: eligibility rule (follow-or-public), live delivery in dev, persistence to DB. Doesn't: history fetch, emoji picker, delivery through proxy.
9. **Group chat** — Works: backend send/authorize/fan-out branch. Doesn't: UI, history.
10. **Docker deployment** — Works: images build plausibly, HTTP path end-to-end via Caddy. Doesn't: WS path; fragile `/app` volume (BUG-018).
11. **Rate limiting** — Works: per-IP window, 429 + `Retry-After`. Doesn't: coherent limits/docs (BUG-013), map pruning (BUG-014).

---

## 22. Extra / not required

| Feature | Location | Assessment |
|---|---|---|
| Post like/dislike reactions | `posthandler.ReactionPost/DeleteReaction`, `reactions` table, `PostCard.jsx` | Harmless, well-gated by `CanViewPost`; extra polish |
| Group invite/join **response** notifications | `groupsvc` | Extra; the subject explicitly welcomes extra notifications |
| Per-IP rate limiting | `middleware/rate_limit.go` | Extra; harmless beyond the doc/header mismatch |
| Third Caddy container | `compose.yml`, `caddy/` | Subject suggests Caddy; fine |
| Login with nickname | `authsvc.Login` | Extra convenience; caused the nickname-required contradiction |
| People directory derived from feed | `people/page.jsx` | Extra workaround; slightly misleading ("Everyone who appears in your feed") |
| Message image attachments | `files.message_id`, `CanAttachToMessage` | Extra; implemented with proper authorization |
| Group detail relationship flags | `GroupDetail`/`GroupListItem` | Supporting extra; good API hygiene |
| Char counters / shared validation module | `CharCount.jsx`, `lib/validate.js` | Extra polish; harmless |
| `POST /avatar` endpoint + settings UI | `filehandler.SetAvatar` | Supports the subject's avatar requirement (post-registration) |

None of these are errors; the notification extras and reactions add review surface but no subject violations.

---

## 23. Final compliance table

| Category | Requirement | Status | Evidence | Problem |
|---|---|---|---|---|
| Frontend | JS framework genuinely used | ✅ | Next.js App Router throughout | — |
| Frontend | Responsiveness | ✅ | `globals.css` mobile block | — |
| Frontend | Navigation | ✅ | `Navbar.jsx`, route groups | — |
| Frontend | FE/BE communication | ✅ | `lib/api.js`, cookies | — |
| Frontend | Register form (optional fields present) | ⚠️ | `register/page.jsx` | nickname required; no avatar input |
| Frontend | Posts composer (privacy + images) | ⚠️ | `PostForm.jsx` | no selected-followers picker |
| Frontend | Comments UI | ❌ | — | feature absent |
| Frontend | Groups UI | ❌ | `groups/page.jsx` | static placeholder |
| Frontend | Events UI | ❌ | — | feature absent |
| Frontend | Private chat UI | ⚠️ | `chat/[id]/page.jsx` | no history; hardcoded WS |
| Frontend | Group chat UI | ❌ | — | absent |
| Frontend | Notifications on every page | ❌ | `notifications/page.jsx` | page-only, no badge/history |
| Frontend | Accept/decline follow UI | ❌ | — | no pending-request list |
| Frontend | Public/private toggle UI | ❌ | `settings/page.jsx` | read-only |
| Auth | Register required fields | ⚠️ | `authsvc` | nickname rule contradicts subject |
| Auth | Login | ✅ | `authhandler.Login` | — |
| Auth | Logout | ⚠️ | `authhandler.Logout` | WS not revoked (BUG-004) |
| Auth | Sessions & cookies | ✅ | `sessionsvc` | no `Secure` flag (HTTP-only deploy) |
| Auth | bcrypt | ✅ | `authsvc` | — |
| Auth | Stay logged in | ✅ | 30-day session | — |
| Followers | Follow request / accept / decline | ⚠️ | `followsvc` | no list endpoint → unusable end-to-end |
| Followers | Public auto-follow | ✅ | `followsvc.Follow` | — |
| Followers | Unfollow | ✅ | `DeleteFollow` | — |
| Followers | Follow-request notification | ❌ | — | not created |
| Profile | Public/private visibility rules | ✅ | `CanViewProfile` | — |
| Profile | All register info minus password | ⚠️ | `PublicUser` | email/DOB omitted |
| Profile | Own posts on profile | ⚠️ | `profile/[id]/page.jsx` | feed-filter approximation |
| Profile | Followers/following lists | ❌ | — | absent |
| Profile | Toggle public/private | ❌ | — | absent |
| Profile | Password never displayed | ✅ | hand-built JSON maps | — |
| Posts | Create with image/GIF | ✅ | `POST /posts` + `/files` | — |
| Posts | Public posts | ✅ | `postVisibleCondition` | — |
| Posts | Almost-private posts | ✅ | follow-based condition | — |
| Posts | Private (selected followers) | 🔴 | `post_visibility` | never populated (BUG-001) |
| Posts | Comments with images | ❌ | — | absent |
| Posts | Group posts | ❌ | — | `group_id` never set |
| Groups | Create / invite / accept / refuse | ✅ | `groupsvc` | — |
| Groups | Members invite others | ✅ | `Invite` | — |
| Groups | Join requests + creator decisions | ✅ | `RequestJoin`/`RespondJoinRequest` | — |
| Groups | Browse all groups | ⚠️ | `GET /groups` | no UI |
| Groups | Group posts/comments member-only | ❌ | — | absent |
| Groups | Events + going/not-going | ❌ | — | tables unused |
| Chat | Private-message eligibility rule | ✅ | `CanMessage` | — |
| Chat | Real-time delivery | 🔴 | `hub.go` | hardcoded WS URL (BUG-003) |
| Chat | Message persistence | ⚠️ | `messages` table | saved but never readable |
| Chat | Emojis | ⚠️ | text passthrough | no picker |
| Chat | Group chat room | ❌ | backend branch only | not end-to-end |
| Notifications | Group invitation / join-request | ✅ | `groupsvc.notify` | — |
| Notifications | Follow-request notification | ❌ | — | absent |
| Notifications | Event notification | ❌ | — | no events |
| Notifications | History / read-unread | ❌ | — | no endpoints |
| SQLite | SQLite used, schema & constraints | ✅ | migrations, `schema.sql` | — |
| SQLite | FK enforcement reliable | 🔴 | `sqlite.go Open()` | per-connection pragma (BUG-002) |
| Migrations | System + files + run at startup | ✅ | golang-migrate embedded | clean-DB run not executed here |
| Images | JPEG/PNG/GIF + validation + serving | ✅ | `filesvc` | — |
| Images | Avatar at registration | ❌ | — | post-login only |
| Docker | Backend + frontend images | ✅ | Dockerfiles | — |
| Docker | Containers communicate | ⚠️ | Caddyfile.docker | WS path broken |
| Docker | Clean-environment startup | ⚠️ | compose.yml | volume-over-/app quirk; not live-tested |
| WebSocket | Upgrade, auth, multi-tab, ping/pong | ✅ | `hub.go` | — |
| WebSocket | Logout revocation | 🔴 | `session.go` | dead code (BUG-004) |

---

## 24. Final summary

### ✅ Working correctly
- Next.js frontend with real routing, validation, responsive layout; login/register flows.
- Session-cookie authentication with bcrypt, HttpOnly cookies, guest/authorized middleware, 30-day persistence.
- Follow system backend: request/accept/decline/unfollow + public auto-follow, duplicate-safe.
- Posts: create/update/delete with ownership checks; public and almost-private visibility; multi-image attachments (JPEG/PNG/GIF sniffed, size-capped, per-viewer file serving).
- **Groups backend end-to-end**: create, browse, detail with anti-enumeration, invitations and join requests with transactional accept/refuse, race-safe duplicates, group notifications persisted and pushed.
- Private-chat eligibility rule exactly as the subject defines it (follow-relationship or public profile), message persistence, multi-tab WebSocket hub with ping/pong.
- SQLite schema with strong constraints; embedded golang-migrate migrations applied at startup; clean build (`go build`/`go vet`).

### ⚠️ Needs improvement
- Real-time delivery only works against `hostname:8080` — must go through the same origin/proxy.
- Notification system is group-only and page-local; needs history endpoint, read state, and a global badge.
- Profile completeness (email/DOB), followers/following lists, privacy toggle, pending-follow-request listing.
- Logout doesn't close WebSockets (dead session-tracking code).
- Docker: WS path, volume-over-`/app`, rate-limit consistency.

### Missing
- Comments (with images); private-post selected followers; group posts/comments; group events + RSVP; groups UI; group chat UI; follow-request & event notifications; notifications on every page; profile privacy toggle; followers/following lists; avatar in registration form; message-history/conversations endpoint; emoji picker.

### Errors / bugs
- 19 findings: 1 critical (BUG-001 private posts unusable), 3 high (FK pragma, hardcoded WS URL, dead session revocation), 7 medium, 8 low — details in §17.

### Extra
- Post reactions, response notifications, rate limiting, Caddy service, nickname login, people directory, chat image attachments, group detail flags, validation/counter polish — see §22.

### Security issues
- Permissive `CheckOrigin` (mitigated by SameSite=Lax); no `Secure` cookie flag (HTTP-only deployment); raw `avatar` string accepted at register; rate-limit header/value mismatch; unbounded limiter map. No SQLi, no XSS, no path traversal, no password/PII leaks found; object-level authorization consistently enforced.

### Docker issues
- WS unreachable in the composed deployment (BUG-003); `backend-data` volume mounted over `/app` (BUG-018); no healthchecks; clean-container run not executed in this audit (*cannot be verified from the provided source code*).

### Database/Migration issues
- Per-connection `PRAGMA foreign_keys` misapplied once (BUG-002); `nickname` not unique while used for login (BUG-010); unused tables/columns (`comments`, `group_events`, `event_responses`, `post_visibility`, `files.comment_id`); expired sessions only reaped on access. Migrations themselves are well-formed, ordered, paired up/down, and executed automatically.

### Overall compliance

| Metric | Count |
|---|---|
| Requirements fully implemented | **34** |
| Requirements partially implemented | **14** |
| Requirements missing | **21** |
| Requirements implemented but with a blocking bug | **4** |
| Distinct bug findings (§17) | **19** (1 critical, 3 high, 7 medium, 8 low) |
| Extra (not required) features | **10** |

Largest compliance gaps to close before presentation, in order of subject impact: **selected-followers private posts (BUG-001)**, **comments**, **group UI (posts/comments/events/chat over the finished groups backend)**, **events**, **notifications on every page + follow-request notification**, **profile privacy toggle**, and **the WebSocket URL fix** that makes chat work in the graded Docker setup.
