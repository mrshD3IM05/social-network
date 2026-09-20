# API Contract — HTTP Endpoints

Base path: `/api/v1` (Caddy and the Next.js dev rewrite both strip it before
forwarding to the Go mux, so the handlers see `/posts`, `/me`, and so on).

Every endpoint needs the session cookie except `POST /register`, `POST /login`
and `GET /health`.

Request bodies are form encoded (`application/x-www-form-urlencoded`), except
where a multipart upload is noted. Responses are JSON.

---

## Conventions

### Pagination

List endpoints take `?limit=` and `?offset=`.

- `limit` defaults to `20` and is capped at `100`
- `offset` defaults to `0`
- a page shorter than `limit` means there is nothing more to fetch

`GET /notifications` and `GET /messages/{id}` take `limit` only; the message
history pages backwards with `before` instead.

### Errors

Errors are plain text with the matching status code:

| Status | Meaning |
|--------|---------|
| 400 | malformed input |
| 401 | no session, or an expired one |
| 403 | authenticated but not allowed (private profile, not the owner) |
| 404 | missing, or hidden from this viewer |
| 409 | conflict (email or nickname taken, duplicate follow) |
| 429 | rate limited, with `Retry-After` |

Rate limiting is per IP: 1000 requests per minute, reported in
`X-RateLimit-Limit` and `X-RateLimit-Remaining`.

---

## Health

```
GET /health
```

`200 {"status":"ok"}`. No session required; the container healthcheck uses it.

---

## Authentication

| Method | Path | Purpose |
|--------|------|---------|
| POST | /register | Create an account, and sign in straight away |
| POST | /login | Sign in with an email **or** a nickname |
| POST | /logout | Sign out, and close that session's websockets |
| GET | /me | The signed-in user, with email and date of birth |
| PATCH | /me | Update your own profile |

### Register

```
POST /register

email, password, first_name, last_name, date_of_birth   (required)
nickname, about_me                                       (optional)
```

`date_of_birth` is `YYYY-MM-DD` and the user must be at least 13.
A nickname is optional, but when given it must be 4–15 lowercase letters or
digits with at least one letter, and no other account may already use it (409).

Avatars are **not** part of this call. Registering signs the user in, so the
client uploads the picture to `POST /avatar` right afterwards.

Response `201`: the user, email and date of birth included.

### Update your profile

```
PATCH /me

first_name, last_name, nickname, about_me, private   (all optional)
```

Only the fields present in the body change. `private` accepts
`true`/`false`/`1`/`0`/`on`/`off`, and is what turns the profile public or
private. Response `200`: the updated user.

---

## Users and profiles

| Method | Path | Purpose |
|--------|------|---------|
| GET | /users | Everyone else on the network (paginated) |
| GET | /user/{id} | One profile |
| GET | /users/{id}/posts | That user's posts, respecting visibility |
| GET | /users/{id}/followers | Who follows them |
| GET | /users/{id}/following | Who they follow |
| GET | /users/{id}/relationship | How you relate to them |

### Get a profile

```
GET /user/{id}
```

Always returns `id`, `first_name`, `last_name`, `avatar`, `nickname`,
`about_me`, `private`, `created_at`.

`email` and `date_of_birth` are added only when the viewer is the owner or an
accepted follower. The password is never part of any response.

A private profile the viewer does not follow answers `403 profile is private`,
which is what the client turns into the locked card. The same rule guards the
follower, following and post lists of that user.

### Relationship

```
GET /users/{id}/relationship
```

`200 {"status": "..."}` where status is `""` (nothing), `pending`, `accepted`,
`declined` or `self`.

---

## Followers

| Method | Path | Purpose |
|--------|------|---------|
| POST | /users/{id}/follow | Follow, or ask to follow |
| DELETE | /users/{id}/follow | Unfollow, or cancel a request |
| GET | /follow-requests | Requests waiting for your answer |
| POST | /follow-requests/{id}/accept | Accept one |
| POST | /follow-requests/{id}/decline | Decline one |

Following a public profile is accepted immediately; a private profile gets a
`pending` request and a `follow_request` notification. Accepting sends a
`follow_accepted` notification back to the requester.

`GET /follow-requests` is what makes accept and decline reachable: it is the
only place a client learns a request id.

```json
[
  {
    "id": 4,
    "status": "pending",
    "created_at": "2026-09-20T17:30:00Z",
    "from_user": { "id": 7, "first_name": "Dina", "last_name": "Doe", "avatar": "" }
  }
]
```

---

## Posts

| Method | Path | Purpose |
|--------|------|---------|
| GET | /posts | The feed (paginated) |
| POST | /posts | Write a post |
| PUT | /posts/{id} | Edit your post |
| DELETE | /posts/{id} | Delete your post |
| GET | /posts/{id}/audience | Who may read a private post |
| PUT | /posts/{id}/audience | Choose who may read it |
| POST | /posts/{id}/reactions | Like or dislike |
| DELETE | /posts/{id}/reactions | Remove your reaction |

### Create

```
POST /posts

content    up to 1000 characters
privacy    public | almost_private | private
```

- `public` — every signed-in user
- `almost_private` — accepted followers of the author
- `private` — only the followers the author picked, see below

Images are attached afterwards with `POST /files` and the new `post_id`.

### Choose who reads a private post

```
PUT /posts/{id}/audience

user_ids=3&user_ids=8       (repeat the field for each user)
```

Replaces the list. Only the author may call it, only on a post whose privacy is
`private`, and ids that do not belong to an accepted follower of the author are
dropped. Response `200 {"user_ids":[3]}` — the list as it was actually stored.

Without this call a `private` post reaches nobody but its author.

### Reactions

`reaction` must be `like` or `dislike`. Sending the same one again removes it.
Response `200 {"likes":2,"dislikes":1,"my_reaction":"like"}`.

---

## Comments

| Method | Path | Purpose |
|--------|------|---------|
| GET | /posts/{id}/comments | Comments on a post (paginated) |
| POST | /posts/{id}/comments | Write one, with images |
| DELETE | /comments/{id} | Delete one |
| POST | /comments/{id}/reactions | Like or dislike |
| DELETE | /comments/{id}/reactions | Remove your reaction |

### Write a comment

```
POST /posts/{id}/comments
Content-Type: multipart/form-data

content    up to 1000 characters
files      up to 3 JPEG/PNG/GIF images, 10 MB each (optional)
```

A plain form without images works too. You can only comment on a post you are
allowed to see. A comment can be deleted by its author **or** by the owner of
the post; deleting it also removes its images.

Response `201`: the comment, with `images` as file ids.

---

## Files

| Method | Path | Purpose |
|--------|------|---------|
| POST | /files | Upload up to 3 images, multipart `files` |
| POST | /avatar | Set your avatar, multipart `avatar` |
| GET | /fs/{id} | Download a file |

JPEG, PNG and GIF only — the type is sniffed from the content, the file name is
never trusted. 10 MB per image. `post_id` **or** `message_id` may be sent to
attach the upload; only the owner of the post may attach to it.

`GET /fs/{id}` checks per viewer that the file is reachable (own upload, an
avatar, a visible post, a comment on one, a conversation you took part in, or a
group you belong to) and answers 404 otherwise.

---

## Notifications

| Method | Path | Purpose |
|--------|------|---------|
| GET | /notifications | History, newest first |
| GET | /notifications/unread | Unread count |
| POST | /notifications/read | Mark them all read |
| POST | /notifications/{id}/read | Mark one read |

```
GET /notifications?limit=50
```

```json
{
  "notifications": [
    {
      "id": 10,
      "user_id": 2,
      "type": "follow_request",
      "actor_id": 7,
      "content": "Dina Doe wants to follow you",
      "group_id": null,
      "read": false,
      "created_at": "2026-09-20T17:30:00Z"
    }
  ],
  "unread": 3
}
```

Types: `follow_request`, `follow_accepted`, `group_invitation`,
`group_join_request`, `group_invite_response`, `group_join_response`.

Every notification is also pushed live over the websocket, so the client keeps
the count in the sidebar up to date without polling.

---

## Chat

| Method | Path | Purpose |
|--------|------|---------|
| GET | /conversations | People you already talked to, newest first |
| GET | /messages/{id} | History with one user |
| GET | /ws | Websocket, for sending and receiving |

```
GET /messages/{userId}?limit=50&before=2026-09-20T17:30:00Z
```

`before` is an RFC3339 timestamp and pages backwards. Messages come back
oldest first, ready to render.

Both endpoints apply the same rule the websocket applies before accepting a
message: the recipient has a public profile, or one of the two users follows
the other. Otherwise `403`.

`GET /conversations`:

```json
[
  {
    "user": { "id": 5, "first_name": "Alice", "last_name": "Smith", "avatar": "abc123" },
    "last_message": {
      "id": 42,
      "from_user_id": 3,
      "to_user_id": 5,
      "content": "See you tomorrow!",
      "created_at": "2026-09-20T14:30:00Z"
    }
  }
]
```

Messages themselves are sent over the websocket, not over HTTP. See
`docs/WEBSOCKET-PROTOCOL.md`.

---

## Groups

Owned by the groups work; see `GROUPS_BACKEND_IMPLEMENTATION.md` for the
details of each payload.

| Method | Path | Purpose |
|--------|------|---------|
| POST | /groups | Create a group |
| GET | /groups | Browse all groups |
| GET | /groups/{id} | Group detail |
| GET | /groups/{id}/members | Members |
| POST | /groups/{id}/invitations | Invite a user |
| GET | /group-invitations | Invitations waiting for you |
| POST | /group-invitations/{id}/accept | Accept one |
| POST | /group-invitations/{id}/decline | Decline one |
| POST | /groups/{id}/join-requests | Ask to join |
| GET | /groups/{id}/join-requests | Requests waiting for the creator |
| POST | /group-join-requests/{id}/accept | Accept one |
| POST | /group-join-requests/{id}/decline | Decline one |
