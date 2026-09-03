# social-network

A social network built on a Go API and a static vanilla-JS frontend, both
served by a single Go process.

## how to run

```powershell
cd backend
go run ./cmd/server
```

Then open <http://localhost:8080>. Ctrl+C to stop.

Flags: `-addr` (default `:8080`) and `-frontend` (default `frontend`).

Run it from `backend/` - that is where it reads `frontend/` from and where it
creates `sn.db` and `uploads/`. Migrations are embedded and run at startup.

### requirements

| what | why | install |
|---|---|---|
| Go 1.25+ | builds and runs the server | <https://go.dev/dl> |
| a gcc toolchain | the sqlite driver (`mattn/go-sqlite3`) is cgo | `winget install BrechtSanders.WinLibs.POSIX.UCRT` |

After installing gcc, reopen the terminal so `PATH` picks it up.

## the pieces

One process, [backend/cmd/server](backend/cmd/server), listening on `:8080`:

| path | served by |
|---|---|
| `/api/v1/*` | the API - the prefix is stripped, so the routes in [backend/internal/server](backend/internal/server) stay prefix-free |
| everything else | static files from [backend/frontend](backend/frontend) |

Because the API and the frontend share an origin, the frontend calls it through
the relative path `/api/v1` - no CORS, and `SameSite=Lax` cookies just work.

## docs

- [backend/readme.md](backend/readme.md) - endpoints, auth, architecture diagrams
- [CHANGELOG.md](CHANGELOG.md)
