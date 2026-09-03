# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/).

## [Unreleased]

### Changed

- **Single-process serving** — `cmd/server` now serves the static frontend alongside the API on `:8080`; the API keeps the `/api/v1` prefix, stripped before the mux, so route registration is unchanged. New `-addr` and `-frontend` flags

### Removed

- **Caddy reverse proxy** (`caddy/`) and the **standalone frontend server** (`servefrontend.go`) — no longer needed now that one Go process serves both
- **`run.ps1` / `stop.ps1`** — with a single process there is nothing to orchestrate; the app starts with `go run ./cmd/server` from `backend/`

### Added

- **Standalone frontend server** (`servefrontend.go`) — simple Go file server for `frontend/` directory, listens on `:5500` by default, configurable via `-addr` flag; replaces Live Server dependency
- **Image dimension validation** — images exceeding 8000x8000 pixels are rejected with `ErrInvalidImage` before full decode, preventing memory exhaustion from small compressed files with extreme dimensions

---

## [1.0.0] - 2026-08-26

### Added

- **Avatar upload** (`POST /avatar`) — authorized endpoint accepting a single image file, stored via the existing file pipeline, sets `users.avatar` to the file ID
- **Thumbnail generation** — pre-generated on every upload using a stdlib-only area-average downscaler (JPEG q82, PNG, static GIF first-frame), capped at 300x300, never upscaled, mandatory with full rollback on failure
- **Thumbnail serving** (`GET /fs/{id}/thumb`) — serves the pre-generated thumbnail with `Cache-Control: public, max-age=31536000, immutable` headers; returns 404 if missing
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
