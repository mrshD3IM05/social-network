# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/).

## [Unreleased]

### Added

- **Standalone frontend server** (`servefrontend.go`) — simple Go file server for `frontend/` directory, listens on `:5500` by default, configurable via `-addr` flag; replaces Live Server dependency
- **Chat image attachments** (`message_id` on `POST /files`) — images can be attached to a chat message; upload is only allowed for the message sender, the private recipient, or a group member (`CanAttachToMessage`); existing `files.message_id` column is now used

### Removed

- **Thumbnail feature** — removed pre-generated 300x300 thumbnail generation, the `GET /fs/{id}/thumb` route, and the `Thumbnail` handler. Uploads now store and serve the original image only (`GET /fs/{id}`).
- **Image dimension validation** — the 8000x8000 `DecodeConfig` check existed only to protect thumbnail decoding; with thumbnails gone the backend never decodes images, so the limit was removed with it (10 MB byte cap still applies)
- **Dead `GET /users` endpoint** — removed the registered route and `ListUsers` handler (returned `501 Not Implemented`); users are found by profile ID via `GET /user/{id}`.
- **Post `type` column** — removed the `type` column added in migration `000016` (dropped in `000018`). It only existed to distinguish auto-generated posts (e.g. an `avatar_update` post), which is not in the subject, so posts are now created solely through the post composer and setting an avatar does **not** create a post.

### Changed

- **Image serving is now cacheable** — `GET /fs/{id}` sets `Cache-Control: private, max-age=31536000, immutable`, so browsers cache images privately (shared caches never store them).
- **Image upload cap lowered to 3** — `POST /files` accepts at most 3 images per request instead of 5 (posts and chat messages); frontend composer mirrors the cap.

---

## [1.0.0] - 2026-08-26

### Added

- **Avatar upload** (`POST /avatar`) — authorized endpoint accepting a single image file, stored via the existing file pipeline, sets `users.avatar` to the file ID
- **Lightbox image preview** — clicking any image (post attachment or avatar) opens a fullscreen overlay displaying the original full-resolution version
- **Frontend avatar UI** — profile card shows avatar image with a "Change avatar" file picker; uploads POST to `/avatar` and refresh the view
- **Feed author info** — posts now include `author_first_name`, `author_last_name`, `author_nickname`, and `author_avatar`; feed renders real names and round author avatars
- **`schema.sql`** — consolidated DDL of all 17 SQLite migrations for reference
- **Caddy config** (`caddy/Caddyfile`) — `handle_path /api/v1/*` strips the prefix before proxying to the Go backend on `:8080`; static frontend served from `:5500`
- **Caddy support files** — `start.ps1`, `stop.ps1`, `LICENSE`, `README.md`
- **Thumbnail visibility fix** — `CanViewFile` now allows any authenticated user to serve files referenced as a user's avatar
- **`POST /files` endpoint documentation** for post image attachments (multipart, up to 5 images, 10 MB each, jpeg/png/gif)

### Changed

- **Renamed `internals/` to `internal/`** — follows Go convention for compiler-enforced package privacy; all import paths updated from `sn-backend/internals` to `sn-backend/internal`
- **Frontend `API_BASE`** set to `/api/v1` for caddy-routed same-origin requests (fixes CORS and `SameSite=Lax` cookie issues)
- **`ListVisiblePosts` and `GetPost`** now `JOIN users` to return author profile data alongside each post
- **`filesvc.Repository` interface** extended with `GetUserByID` and `UpdateUser` to support avatar updates

### Removed

- Legacy URL-string avatar input from registration form (avatar is now set via file upload after login)

### Fixed

- Caddyfile prefix handling — old `reverse_proxy` forwarded `/api/v1/login` verbatim (Go mux has `/login`), causing 404s; replaced with `handle_path` to strip the prefix

---

## [0.2.0] - 2026-08-24

### Added

- **Caddy reverse proxy config** — `Caddyfile` and `Caddyfile.docker` for routing `/api/v1/*` to the Go backend and serving the frontend separately
- **Frontend placeholder** (`frontend/readme.md`)

---

## [0.1.0] - 2026-08-24

### Added

- **Authentication** — register (`POST /register`), login (`POST /login`), logout (`POST /logout`), session-based with HttpOnly cookies, bcrypt password hashing, 30-day expiry
- **User profiles** — public/private profiles, follow requests, accept/decline follow requests, unfollow
- **Posts** — create, read, update, delete; three privacy levels (`public`, `almost_private`, `private`), per-user visibility control
- **File uploads** (`POST /files`) — multipart, up to 5 images per request, 10 MB each, jpeg/png/gif validation, stored on disk with metadata in `files` table
- **File serving** (`GET /fs/{id}`) — authenticated, per-user visibility check
- **Groups** — create groups, manage members, invitations, join requests
- **Group events** — create events within groups, RSVP responses
- **Notifications** — system-generated notifications for follows, group invites, events
- **Messages** — private (1:1) and group messaging via WebSocket (`GET /ws`)
- **WebSocket** — gorilla/websocket hub for real-time chat and notifications
- **Rate limiting** — per-IP sliding window: 100 requests/minute, 429 with `Retry-After`
- **SQLite database** — WAL mode, foreign keys, auto-migration via golang-migrate with 17 embedded migrations
- **Layered architecture** — model, middleware, handler, service, repository, db; clear request flow
- **Middleware** — session-based auth (`Authorized`/`Guest`), rate limiting
- **Tests** — auth middleware, auth handler, repository integration, migration integrity

### Database Schema

- `users`, `sessions`, `posts`, `post_visibility`, `comments`, `follow_requests`
- `groups`, `group_members`, `group_invitations`, `group_join_requests`
- `group_events`, `event_responses`
- `notifications`, `messages`, `reactions`, `files`
