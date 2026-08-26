## INTRODUCTION
this is an api backend for our social network

## how to run
go run ./cmd/server
the server listens on :8080
it creates a sqlite database (sn.db) in the working directory and runs the embedded sql migrations automatically

## endpoints (/api/v*/ prefix shall be added using caddy)
/register and /login are guest only (logged in users get rejected), everything else needs a session cookie

### auth
| method | path | request | response |
|---|---|---|---|
| POST | /register | form: email, password, first_name, last_name, date_of_birth, nickname, about_me, avatar | 201 + private user json, logs you in right away |
| POST | /login | form: email, password | 200 + private user json |
| POST | /logout | - | deletes the session, clears the cookie and force closes that session's websockets |
| GET | /me | - | current user (private shape: adds email and date_of_birth to the public one) |

### users & follows
| method | path | request | response |
|---|---|---|---|
| GET | /users | - | not implemented yet (501) |
| GET | /user/{id} | - | public profile, 403 if the profile is private and you don't follow them |
| POST | /users/{id}/follow | - | follows the user, or creates a follow request if their profile is private |
| DELETE | /users/{id}/follow | - | unfollows |
| POST | /follow-requests/{id}/accept | - | 204 |
| POST | /follow-requests/{id}/decline | - | 204 |

### posts
| method | path | request | response |
|---|---|---|---|
| GET | /posts | - | posts visible to you (will add cursor pagination to it) |
| POST | /posts | form: content, privacy = public \| almost_private \| private | 201 + post json |
| PUT | /posts/{id} | form: content, privacy | 200 + post json, only the owner can update |
| DELETE | /posts/{id} | - | 204, only the owner can delete |

### files
| method | path | request | response |
|---|---|---|---|
| POST | /files | multipart: files[] or file (max 5 files, 10 MB each, jpeg/png/gif only), optional post_id to attach them to a post | 201 + stored file json |
| POST | /avatar | multipart: avatar (single image, same type/size limits) | 200 + private user json, sets your avatar |
| GET | /fs/{id} | - | serves the original file after a per user visibility check, 404 if you can't see it |
| GET | /fs/{id}/thumb | - | serves the pre-generated thumbnail (max 300x300) with long lived cache headers, 404 if you can't see the file or no thumbnail exists |

thumbnails are generated once during upload and stored next to the originals under uploads/thumbnails/, never resized per request
images are validated for dimensions (max 8000x8000) before full decode to prevent memory exhaustion; the existing 10 MB upload size limit is unchanged
animated gifs get a static first frame thumbnail; images smaller than 300x300 are kept as-is; uploads made before thumbnails existed return 404

### websocket
| method | path | request | response |
|---|---|---|---|
| GET | /ws | upgrade | chat + notifications over one socket (gorilla/websocket) |

client sends {"type":"message", "to_user_id" or "group_id", "content"} (exactly one target)
server sends back messages (echoed to the sender too), {"type":"notification", ...} events and {"type":"error", ...} for rejected input

## auth
cookie sessions, no jwt
a random 32 byte token stored in the sessions table with a 30 day expiry
cookie name is "session" (HttpOnly, SameSite=Lax)
passwords hashed with bcrypt

## rate limiting
every request goes through a per ip limiter: 100 requests per minute (sliding window), 429 with Retry-After when exceeded

## database
sqlite (WAL mode, foreign keys on)
migrations are embedded with go:embed and run at startup using golang-migrate (internal/db/migrations/sqlite)
tables: users, sessions, posts, follow_requests, groups, group_members, group_invitations, group_join_requests, group events, notifications, messages (private + group), reactions, files

## architecture
we are using a layered architecture
(model, middleware, handler, websocket, service, repository, db)

request flow: middleware -> handler -> service -> repository -> sqlite
the websocket hub sits next to that stack and talks to the repository directly

```mermaid
block-beta
    columns 1

    block:layers
        columns 5
        MW["middleware\nrate limit + auth"]
        H["handler\nrequest parsing + response"]
        S["service\nbusiness logic"]
        R["repository\nquery building"]
        DB["sqlite\nWAL mode + migrations"]
    end

    block:ws
        columns 3
        space
        WS["websocket hub\nchat + notifications"]
        space
    end

    MW --> H --> S --> R --> DB
    WS --> R
```

```mermaid
sequenceDiagram
    participant C as client
    participant MW as middleware
    participant H as handler
    participant S as service
    participant R as repository
    participant DB as sqlite

    C->>MW: HTTP request
    MW->>MW: rate limit check (per IP, 100/min)
    MW->>MW: session cookie lookup
    alt guest route
        MW->>H: forward (guest allowed)
    else authorized route
        MW->>MW: reject if no session
        MW->>H: forward (userID injected)
    end
    H->>H: parse request body / params
    H->>S: call service method
    S->>S: business logic + validation
    S->>R: call repository method
    R->>R: build SQL query
    R->>DB: execute query
    DB-->>R: rows / result
    R-->>S: model structs
    S-->>H: response data
    H-->>C: JSON + status code
```

```mermaid
sequenceDiagram
    participant C as client
    participant WS as websocket hub
    participant R as repository
    participant DB as sqlite
    participant R2 as repository
    participant H as handler (via hub)

    C->>WS: upgrade GET /ws (session cookie)
    WS->>WS: register session in hub
    loop realtime chat
        C->>WS: {"type":"message","to_user_id":2,"content":"hi"}
        WS->>WS: validate sender session
        WS->>R: CreateMessage(from, to, content)
        R->>DB: INSERT INTO messages
        WS->>WS: find recipient session
        WS->>C: {"type":"message",...} (echo to sender)
        WS->>C: {"type":"message",...} (deliver to recipient)
    end
    Note over WS,R: hub talks directly to repository, bypassing handler/service layers
```

## sequence diagrams

### login

Client authenticates with email/password. The middleware checks rate limits, the handler looks up the user, the service verifies the bcrypt hash, and a session cookie is set on success.

```mermaid
sequenceDiagram
    participant C as client
    participant MW as middleware
    participant H as handler
    participant S as service
    participant R as repository
    participant DB as sqlite

    C->>MW: POST /login (email, password)
    MW->>MW: rate limit check (per IP)
    MW->>H: forward request
    H->>S: FindUserByEmail(email)
    S->>R: FindUserByEmail(email)
    R->>DB: SELECT * FROM users WHERE email = ?
    DB-->>R: user row
    R-->>S: user
    S->>S: bcrypt.CompareHashAndPassword
    S-->>H: user, err (auth success)
    H->>S: CreateSession(user.ID)
    S->>R: CreateSession(token, userID, expiry)
    R->>DB: INSERT INTO sessions
    H->>C: 200 + Set-Cookie: session=token, HttpOnly, SameSite=Lax
```

### file upload + thumbnail generation

Client uploads up to 5 images (max 10 MB each). The service writes the original to disk, validates dimensions via `image.DecodeConfig` before full decode, generates a 300x300 thumbnail using an area-average downscaler, and stores both on disk. File metadata is written to the database after the thumbnail is saved; if thumbnail generation fails the original is rolled back.

```mermaid
sequenceDiagram
    participant C as client
    participant H as filehandler
    participant S as filesvc
    participant F as file system
    participant R as repository
    participant DB as sqlite

    C->>H: POST /files (multipart, up to 5 images)
    loop for each file
        H->>H: parse multipart, validate size (10 MB) + type (jpeg/png/gif)
        H->>S: Upload(file, postID)
        S->>F: write original to uploads/<id>
        S->>S: detectContentType (512 byte header read)
        S->>S: image.DecodeConfig — validate dimensions (≤8000×8000)
        S->>S: jpeg/png/gif.Decode — full decode
        S->>S: fitThumbnail (scale down to ≤300×300, never upscale)
        S->>S: encode thumbnail (JPEG q82 / PNG / GIF first frame)
        S->>F: write thumbnail to uploads/thumbnails/<id>
        S->>R: store file metadata
        R->>DB: INSERT INTO files
    end
    H-->>C: 201 + stored file JSON
```

### follow request flow

A user follows another user. If the target's profile is public, the follow is accepted immediately. If private, a pending follow request is created and the target user receives a real-time WebSocket notification.

```mermaid
sequenceDiagram
    participant C1 as follower
    participant H as handler
    participant S as service
    participant R as repository
    participant DB as sqlite
    participant WS as websocket hub
    participant C2 as target user

    C1->>H: POST /users/{id}/follow
    H->>S: Follow(userID, targetID)
    S->>R: FindUserByID(targetID)
    S->>S: if target profile is private?
    alt public profile
        S->>R: CreateFollow(followerID, targetID)
        R->>DB: INSERT INTO follow_requests (status=accepted)
    else private profile
        S->>R: CreateFollow(followerID, targetID, status=pending)
        R->>DB: INSERT INTO follow_requests (status=pending)
    end
    S-->>H: follow created
    H->>WS: notify(targetID, follow event)
    WS->>C2: {"type":"notification", "event":"follow_request", ...}
    H-->>C1: 201
```

### accept / decline follow request

The target user reviews a pending follow request and accepts or declines it. The repository updates the status in the `follow_requests` table.

```mermaid
sequenceDiagram
    participant C as target user
    participant H as handler
    participant S as service
    participant R as repository
    participant DB as sqlite

    C->>H: POST /follow-requests/{id}/accept
    H->>S: AcceptFollowRequest(requestID)
    S->>R: FindFollowRequest(requestID)
    S->>R: UpdateFollowRequestStatus(requestID, accepted)
    R->>DB: UPDATE follow_requests SET status = 'accepted'
    H-->>C: 204
```

### websocket messaging

Clients connect via WebSocket with a session cookie. The hub registers each session and routes messages directly through the repository, bypassing the handler/service layers. Messages are persisted and delivered to both the sender and recipient in real time.

```mermaid
sequenceDiagram
    participant C1 as sender
    participant WS as websocket hub
    participant H as handler
    participant R as repository
    participant DB as sqlite
    participant C2 as recipient

    C1->>WS: connect GET /ws (session cookie)
    WS->>WS: register session in hub
    loop chat
        C1->>WS: {"type":"message","to_user_id":2,"content":"hi"}
        WS->>H: handleMessage(senderID, payload)
        H->>R: CreateMessage(from, to, content)
        R->>DB: INSERT INTO messages
        H-->>WS: message stored
        WS->>WS: find recipient session in hub
        WS->>C2: {"type":"message","from_user_id":1,"content":"hi"}
        WS->>C1: {"type":"message","from_user_id":1,"content":"hi"}
    end
```

### avatar upload

Client uploads a single image for their avatar. The service follows the same pipeline as post image uploads — writing the original, validating dimensions, generating a thumbnail — then updates the `users.avatar` field to point to the new file. Thumbnails are served via `GET /fs/{id}/thumb`.

```mermaid
sequenceDiagram
    participant C as client
    participant H as filehandler
    participant S as filesvc
    participant F as file system
    participant R as repository
    participant DB as sqlite

    C->>H: POST /avatar (single image)
    H->>H: parse multipart, validate size + type
    H->>S: SetAvatar(userID, header)
    S->>S: Upload(ownerID, header, nil)
    S->>F: write original to uploads/<id>
    S->>S: detectContentType (512 byte header read)
    S->>S: image.DecodeConfig — validate dimensions (≤8000×8000)
    S->>S: jpeg/png/gif.Decode — full decode
    S->>S: fitThumbnail (scale down to ≤300×300, never upscale)
    S->>S: encode thumbnail (JPEG q82 / PNG / GIF first frame)
    S->>F: write thumbnail to uploads/thumbnails/<id>
    S->>R: store file metadata
    R->>DB: INSERT INTO files
    S->>R: GetUserByID(ownerID)
    R->>DB: SELECT * FROM users WHERE id = ?
    S->>S: user.Avatar = file.ID
    S->>R: UpdateUser(user)
    R->>DB: UPDATE users SET avatar = fileID
    H-->>C: 200 + updated private user JSON
```
