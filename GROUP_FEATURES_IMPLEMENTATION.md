# GROUP FEATURES IMPLEMENTATION — Group Posts, Group Comments, Group Events

**Date:** 2026-09-22 · **Branch:** `oualid` · Backend: Go 1.25 (`sn-backend`) · Frontend: Next.js 16 / React 19 · DB: SQLite (no schema change required)

Everything reuses the project's existing architecture: `handler → service → repository → SQLite`, service-level `Repository` interfaces, `common.WriteJSON` / flat-text `http.Error` responses, `CreateNotification` + `hub.PublishNotification` for notifications, and the existing `PostForm` / `PostCard` / `Modal` components on the frontend. No parallel architecture was created.

---

## 1. Implemented Features

| Feature | Status |
|---|---|
| Group posts (create + list, members only) | ✅ |
| Group comments (create + list, members only) | ✅ |
| Group events (create + list, members only) | ✅ |
| Event responses (going / not going, changeable, deduplicated) | ✅ |
| Group page UI (header bar, posts, comments, events, composer) | ✅ |
| Group Info modal (existing `GET /groups/{id}` data) | ✅ |
| Notifications (`event_created`, `comment_post`) | ✅ |
| Comment images (reuses `files.comment_id` reserved in migration 10) | ✅ |
| Comment count on every post payload | ✅ |

## 2. Backend Changes

### Authorization model (the core of the task)
`repository/post.go::postVisibleCondition` was rewritten into one SQL predicate used by **every** post read path (`ListVisiblePosts`, `CanViewPost`, `CanViewComment`, `CanViewFile`):

- **Group posts** (`group_id IS NOT NULL`): visible iff the viewer is in `group_members` for that group. The privacy column is ignored for group posts — membership *is* the visibility.
- **Normal posts** (`group_id IS NULL`): the previous author/public/followers/`post_visibility` logic, byte-for-byte behavior preserved.
- `CanViewFile` was extended so files attached to group posts and to comments on any post inherit exactly the post's visibility (previously a group post's images/comments would have leaked to non-members through `GET /fs/{id}` — an IDOR this task would otherwise have introduced). The group-message file branch was tightened to `m.group_id IS NOT NULL` so NULL group IDs can never match.
- `ListVisiblePosts` keeps `group_id IS NULL` (feed shows only non-group posts); group posts are served exclusively through member-gated `GET /groups/{id}/posts` and `GET /posts/{id}`.

Because comments and reactions all gate on `CanViewPost`, the whole chain `user → group membership → group post → comment` is enforced server-side; the client's `post_id` is never trusted.

### Routes (`internal/server/server.go`)
```
GET  /posts/{id}                    get one visible post (new, posthandler.GetPost)
GET  /posts/{id}/comments           list comments (commenthandler)
POST /posts/{id}/comments           create comment (commenthandler)
GET  /groups/{id}/posts             list group posts (grouphandler, member gate in postsvc)
POST /groups/{id}/posts             create group post (grouphandler, member gate in postsvc)
GET  /groups/{id}/events            list events with counts (grouphandler, eventsvc)
POST /groups/{id}/events            create event (grouphandler, eventsvc)
GET  /events/{id}/response          my response (grouphandler, eventsvc)
POST /events/{id}/response          set/change response (grouphandler, eventsvc)
```
All wrapped in the existing `auth.Authorized` middleware.

### Services
- **`postsvc`** — new `Get` (visibility-gated), `CreateGroupPost` (member check, sets `group_id`, pins privacy to `public` and ignores client privacy), `GroupPosts` (member check). `Repository` interface extended with `IsGroupMember`, `ListGroupPosts`.
- **`commentsvc` (new)** — `Create` (content validation 1–2000 chars, `CanViewPost` gate, notification to the post author — never to self) and `List` (`CanViewPost` gate).
- **`eventsvc` (new)** — `Create` (group must exist, caller must be a member — any member may create events, matching the subject; title ≤100, description ≤1000, date+time required and not in the past; notification to every member except the creator), `List` (member gate), `Respond` (choice ∈ {going, not_going}, membership verified via `GetGroupIDForEvent` → `IsGroupMember` — never from client input), `MyResponse`.
- **`filesvc`** — `Upload` accepts `comment_id` (owner-only attachment check via `GetComment`), same magic-byte/size/count rules.

### Handlers
- **`grouphandler`** — now also owns post-service and event-service references (`grouphandler.New(service, post, events, session)`): `CreateGroupPost`, `ListGroupPosts`, `CreateEvent`, `ListEvents`, `RespondEvent`, `MyEventResponse`. Date + time arrive as HTML `<input type=date>` / `<input type=time>` values and are parsed with `time.ParseInLocation("2006-01-02 15:04", …)`.
- **`commenthandler` (new)** — `ListComments`, `CreateComment`; invisible posts answer **404** (never 403) so the API does not reveal hidden posts, same convention as the reaction endpoints.
- **`posthandler`** — `GetPost`.

### Repository
- **`comment.go` (new)** — `CreateComment`, `GetComment`, `ListPostComments` (with author fields + comment images), `countPostComments`/`CountPostComments` (batched, no N+1), `CanViewComment` (reuses `postVisibleCondition`).
- **`event.go` (new)** — `CreateEvent`, `ListGroupEvents` (single grouped query: events + going/not-going counts + viewer's own `choice`), `SetEventResponse` (`INSERT … ON CONFLICT (event_id, user_id) DO UPDATE` — one row per user+event guaranteed by the migration-6 UNIQUE constraint, so a "duplicate response" is impossible and changing the answer replaces it), `GetEventResponse`, `EventResponseCounts`.
- **`group.go`** — `IsGroupPost`, `GetGroupIDForPost`, `IsGroupEvent`, `GetGroupIDForEvent`, `ListGroupPostIDs` (authorization helpers).
- **`post.go`** — group-aware `postVisibleCondition`, `ListGroupPosts`, shared `enrichPosts` (images + reactions + comment counts, used by feed and group lists), `GetPost` now returns `comment_count`.
- **`file.go`** — group-post and comment-aware `CanViewFile`.
- **`repository.go`** — tiny `placeholders`/`int64sToAny` SQL helpers.

### Models
- **`comment.go` (new)** — `Comment` (+ author fields, images) and `NotificationCommentPost = "comment_post"`.
- **`event.go` (new)** — `GroupEvent`, `EventResponse`, `EventListItem` (with `going_count`, `not_going_count`, `my_choice`, creator name) and `NotificationEventCreated = "event_created"`.
- **`post.go`** — added `CommentCount`.

### Database
**No migration was needed.** Verified against the live schema: `posts(id, author_id, content, privacy, group_id, created_at)` (image column dropped in 000014, type added 000016 / dropped 000018), `comments(id, post_id, author_id, content, created_at)` with `idx_comments_post`, `group_events(id, group_id, creator_id, title, description, date_time, created_at)`, `event_responses(id, event_id, user_id, choice, created_at, UNIQUE(event_id, user_id))` with the right FK cascades and indexes — everything this feature needs already exists. `files.comment_id` (migration 10) was reserved but unused; it is now wired up.

## 3. Frontend Changes

- **`app/(main)/groups/[id]/page.jsx`** — rebuilt around the requested layout:
  - **Header bar**: group avatar (initials tile), group name + member count, `Group Info` button on the right.
  - **Group Info modal**: title, description, creator, member count, created date, your status, and the full member list — all from the existing `GET /groups/{id}` payload; "Invite people" reopens the existing invite modal.
  - **Members** see Group posts (reusing `PostCard`, now with comments) and Events; **outsiders** keep the description / request-to-join / invited states. Posts and events return 403 for outsiders; the page simply shows the empty states.
  - **Bottom composer**: the real `PostForm` (group mode — this is the existing post feature, not a fake chat) with a `+ Create Event` button under it, exactly the composer-like area requested. No fake messaging was introduced; group chat remains its own subject feature.
- **`components/EventCard.jsx` (new)** — title, description, weekday/date, time, `Going: X · Not going: Y`, your answer, and Going / Not going buttons. Optimistic UI: counts and the active button flip instantly and are reconciled with the server response; errors roll back and are shown.
- **`components/EventFormModal.jsx` (new)** — title, optional description, native date + time pickers; client-side validation mirrors the API (required, max lengths, no past dates); on 201 the modal closes and the event is appended to the list without a page reload, with a success notice.
- **`components/PostForm.jsx`** — new `groupId` prop: group posts go to `POST /groups/{id}/posts` and the privacy selector is hidden (group posts are member-only by design); feed behavior unchanged.
- **`components/PostCard.jsx`** — new expandable comments section (count badge, `GET /posts/{id}/comments`, comment form with optional image upload via `comment_id`, error states). The privacy label is hidden for group posts (it is meaningless there). Works unchanged in the feed.
- **`lib/validate.js`** — `comment: 2000` limit (mirrors `commentsvc`).
- **`app/globals.css`** — new sections for comments, events, the group header bar and the group composer; the existing `@media (max-width: 860px)` block was extended (header bar wraps, full-width Group Info button, event actions become a full-width row) so the group page adapts to phones and tablets without hardcoded widths.

## 4. Testing

`go build ./...` and `go vet ./...` pass; `next build` passes. A live server test (`curl` with 4 cookie-jar users, fresh DB) exercised the full matrix:

**Group posts** — member creates (201); non-member / pending-join-request user creates (403); member lists (200, correct body); non-member lists (403); member of another group lists (403); `GET /posts/{id}` member (200) / non-member (404); member reacts (200) / non-member reacts (404).

**Comments** — members comment on a group post (201, 201); non-member comments (404); non-member lists comments (404); member lists (200, count 2); `comment_count` correct in group and feed payloads; cross-group comment (404); **normal-post comments unaffected** (201 + list 200).

**Events** — member creates (201); non-members of both groups create (403); bad date (400); empty title (400); member lists (200 with counts + creator name); non-member lists (403); respond going (counts `{1,0}`), second member not_going (`{1,1}`), **change to going** (`{2,0}` — one row per user+event upserted, no duplicates possible); invalid choice (400); non-member respond (403); unknown event (404); `GET /events/{id}/response` returns the stored choice; non-member query (403).

**Isolation / regression** — non-member's feed contains the group post **0** times; normal public posts still flow; invitation + reaction flows still work. Direct DB inspection confirmed persisted notifications: `event_created` → every member **except the creator**, `comment_post` → the post author (never self), plus all pre-existing group notification types. Frontend smoke test: `/home` 200, `/groups/1` 200, `/api/v1/*` proxy wired.

## 5. Problems Found (not fixed — unrelated to this task)

- `GET /users` does not exist, so "people" lists are derived from feed authors (BUG-025 in PROJECT_AUDIT.md); the invite modal inherits this limitation.
- WebSocket URL is hardcoded to `ws://hostname:8080` (BUG-009) and the Next rewrite targets `localhost:8080` (BUG-010) — compose deployment issue, untouched here.
- The notifications page is live-only; it now will display `event_created`/`comment_post` events received in real time, but there is still no `GET /notifications` endpoint (BUG-016).
- Declined group invitations/join requests permanently block re-inviting (BUG-015).
- Follow requests never produce notifications (BUG-007).
- `websocket/session.go` remains dead code; logout does not close live sockets (BUG-011).
