# HANDOFF: social-network

> **Next agent:** read this file, check it against the repo with `git status`, then continue from "Next steps". The section below is updated automatically from the latest coding session in this folder; notes added outside it are kept.

<!-- handoff:auto:start -->
## Auto handoff

_Last update: 2026-09-17 21:34 +0100, from a **Claude Code** session (claude-opus-5). Kept current by the [handoff app](https://github.com/chentaymane/handoff): this section is rewritten automatically, anything outside it is kept._

### Goal
> just in front we need add some checkers for inputs in all profile name post and all inputs max text and other and max 3 image per post

### Where we stopped
> All frontend checks are in, and the build compiles clean.
>
> ## What I added
>
> **[frontend/lib/validate.js](frontend/lib/validate.js)** — one module holding every rule, so no form invents its own. The limits mirror the Go API (`internal/service/authsvc` and `internal/service/filesvc`) so the client never rejects what the server accepts, or vice-versa:
>
> | Input | Rule |
> |---|---|
> | First / last name | required, ≤ 50 |
> | Email | required, ≤ 254, same regex as the API |
> | Nickname | 4–15 lowercase letters/digits, ≥ 1 letter |
> | Password | 8–72 (72 = bcrypt's cut-off) |
> | Date of birth | real date, not future, 13+ years, ≤ 120 |
> | About me | optional, ≤ 500 |
> | Post | required, ≤ 1000 |
> | Chat message | required, ≤ 1000 |
> | People search | ≤ 50 |
> | **Images per post** | **max 3**, JPEG/PNG/GIF, 10 MB each |
>
> **[PostForm.jsx](frontend/components/PostForm.jsx)** — picking more than 3 images (or a wrong format/oversized file) is refused right in the picker with a message, the input resets so the user can pick again, and a **Remove** button clears the selection. Textarea capped at 1000 with a live counter; Publish is disabled while the post is empty or over the limit; content is trimmed before sending.
>
> **[register/page.jsx](frontend/app/(auth)/register/page.jsx)** — every field validated on submit with its own message under it, a red border on the offending input, `maxLength` on each field, `max` on the date picker set to today − 13 years, and a field's message clears as soon as it's e ...

### Next steps
1. Read "Where we stopped" and finish anything it left open.
2. Run `git status` to find uncommitted or half-finished edits before starting new work.

### Recent commands
- `python - <<'PY' import io p = "app/globals.css" s = io.open(p, encoding="utf-8").read() s = s.replace(""" color: var(--ink-soft); font-size: 13px; font-weigh...`
- `npx next build 2>&1 | tail -30`
- `cp lib/validate.js "$TMPDIR/validate.mjs" 2>/dev/null || cp lib/validate.js /c/Users/chent/AppData/Local/Temp/claude/c--Users-chent-Desktop-social-network/f9...`
- `SP=/c/Users/chent/AppData/Local/Temp/claude/c--Users-chent-Desktop-social-network/f99f6ee7-283a-442a-abdf-be9ee3496717/scratchpad && mkdir -p "$SP" && cd /c/...`
- `npx eslint 2>&1 | tail -20`
- `cd /c/Users/chent/Desktop/social-network && git status --short && echo "=== DIFF STAT ===" && git diff --stat`
- `ls -la HANDOFF.md && head -20 HANDOFF.md`
- `cd /c/Users/chent/Desktop/social-network/frontend && python - <<'PY' import io p = "app/(auth)/register/page.jsx" s = io.open(p, encoding="utf-8").read() s =...`

### Errors seen
- Exit code 1 /usr/bin/bash: line 26: /c/Users/chent/AppData/Local/Temp/claude/c--Users-chent-Desktop-social-network/f99f6ee7-283a-442a-abdf-be9ee3496717/scratchpad/t.mjs: No such file or directory node:internal/modules/cjs/loader:1479 thr...

### Session
- **Tool:** Claude Code (claude-opus-5), 35 tool calls
- **Active:** 2026-09-17 21:26 to 2026-09-17 21:34
- **Context:** ~83K of 200K tokens (42%, window size assumed)
- **Full log:** `C:\Users\chent\.claude\projects\c--Users-chent-Desktop-social-network\f99f6ee7-283a-442a-abdf-be9ee3496717.jsonl`

### Repo
- **Branch:** `main` @ `a794e5d` - change the font
- **Upstream:** `origin/main` - ahead 0, behind 0
- **Uncommitted:** 7 changed, 2 untracked

Recent commits:

```
a794e5d 2026-09-15 change the font
84d798c 2026-09-15 move the front to the main branch
b4c2272 2026-09-12 Update compose.yml
c2c8dea 2026-09-11 swap frontend for Next.js and add post reactions
b478599 2026-09-10 remove post type column and auto-post-on-avatar concept
```

Working tree:

```
 M frontend/app/(auth)/login/page.jsx
 M frontend/app/(auth)/register/page.jsx
 M frontend/app/(main)/chat/[id]/page.jsx
 M frontend/app/(main)/people/page.jsx
 M frontend/app/(main)/settings/page.jsx
 M frontend/app/globals.css
 M frontend/components/PostForm.jsx
?? frontend/components/CharCount.jsx
?? frontend/lib/validate.js
```
<!-- handoff:auto:end -->
