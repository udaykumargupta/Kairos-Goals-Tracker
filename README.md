# Kairos — Goals Tracker

> *Kairos* (καιρός): the ancient Greek word for the **right, opportune moment** to act.

**🔗 Live: https://kairos-goals-tracker.vercel.app/**

Track goals across every time horizon — **today, this month, this year, and any custom date range** — with a calendar, journal, streaks, photos, voice notes and shareable progress. Sign in with Google or email, or use it as a guest with everything saved locally in your browser.

**Stack:** Go (standard library only) · PostgreSQL (Neon) · vanilla JS — deployed on Vercel as a serverless function with the frontend on the CDN.

---

## Features

**Goals & tracking**
- **📅 Calendar** — click a day to manage its goals, **⭐ bookmark** dates, and get **✓ / ✗ completion marks** when a day clears (or misses) your threshold.
- **Four scopes** — Day, Month, Year, and **⏳ custom timelines** (your own dated window with a live *Upcoming / Active / Ended* status).
- **⏱ Time meters** — how much of each period has elapsed vs. remaining.
- **→ Next day** — copy a day's unfinished goals forward without losing the originals.
- **📊 Progress** — current and longest streaks, weekly/monthly completion rates, and a 26-week completion heatmap.
- **🎊 Celebration** — a full-screen confetti and fireworks moment when a day clears its threshold.

**Capture**
- **📓 Journal** for every day, month, year and timeline, with a full-page writing mode.
- **🎙 Voice notes** recorded straight into the journal as inline chips you can drag and reposition in the text.
- **🖼 Photo scrapbook** — post images to a day, scattered as a tilted collage; click to enlarge.
- **🗒️ Notes & Reminders** — a always-available personal scratchpad.
- **🎵 Song of the day** — queue up to five YouTube tracks per day that play in a loop.

**Accounts & sharing**
- **Google sign-in, email + password, or guest mode** — per-user cloud sync with an offline cache, or fully local.
- **Forgot password** — emailed, single-use, one-hour reset links.
- **🔗 Read-only share links** — anyone with the link sees your progress (including your streaks) but can't edit.
- Custom display name and avatar, quote of the day, light / dark theme.

**Installable**
- A **PWA** — installable on desktop and mobile, works offline via a service worker.
- Ships as an **Android app** through a Trusted Web Activity, verified by `/.well-known/assetlinks.json`.

---

## Architecture

```
Browser (frontend/index.html — one self-contained file)
  │  Google Identity Services → Google ID token
  │  POST /api/auth/google  →  verify against Google's JWKS → upsert user → issue app JWT
  │  Authorization: Bearer <jwt>
  ├─ /api/state              load / save the whole goal + journal document (jsonb)
  ├─ /api/profile[/picture]  display name / avatar
  └─ /api/share…             enable a token; /?share=<token> renders a read-only view

Vercel
  ├─ frontend/ → copied to public/ at build time, served from the CDN
  └─ api/index.go → one Go function; every /api/** path is rewritten to it
                    and dispatched internally (see vercel.json)
```

**The backend has no third-party dependencies** — there is no `go.sum`, and nothing is fetched at build time beyond the standard library. That is deliberate, and it shapes a few choices:

| Concern | How it's done | Why |
|---|---|---|
| Postgres | Neon's **SQL-over-HTTP** endpoint | No driver dependency, and a stateless request per query suits serverless — there are no pooled TCP connections to leak as function instances come and go. |
| JWT (HS256) | `crypto/hmac` | ~100 lines; the algorithm is derived from the secret's length so tokens stay interchangeable with the previous Spring implementation. |
| Google ID tokens | `crypto/rsa` + Google's JWKS | Verifies signature, issuer, audience, expiry and `email_verified`. |
| Passwords | **PBKDF2-HMAC-SHA256** (`crypto/hmac`) | BCrypt isn't in the standard library. Verified against published RFC vectors. |
| Email | `net/smtp` | Password-reset links only. |

`cmd/server` runs the identical routes as an ordinary long-lived HTTP server, so the app also works on any container host.

---

## Project structure

```
Kairos/
├── frontend/index.html     # the single-file web app (plus manifest.json, sw.js, icons/)
├── api/index.go            # Vercel serverless entrypoint
├── kairos/                 # the backend: routing, auth, store, jwt, mail
├── cmd/server/             # the same API as a normal HTTP server (local dev / containers)
├── vercel.json             # static build + /api/** routing
├── backend/                # original Spring Boot implementation (still deployable via Dockerfile)
├── Dockerfile              # builds the Java app + bundles the frontend
└── DEPLOY.md               # self-hosting guide
```

---

## Run locally

Prerequisites: **Go 1.22+** and a PostgreSQL database (a free [Neon](https://neon.tech) project is easiest — the HTTP endpoint it exposes is what the app talks to).

```bash
# Point at your database and set a signing secret (min. 32 characters)
export DATABASE_URL="postgresql://user:password@your-endpoint.neon.tech/kairos?sslmode=require"
export JWT_SECRET="$(openssl rand -base64 32)"
export GOOGLE_CLIENT_ID="…apps.googleusercontent.com"   # optional: only for Google sign-in

go run ./cmd/server
```

Open <http://localhost:8080> → sign in, or **Continue as guest** (no Google setup needed).

`DATABASE_URL` also accepts the JDBC-style trio used by the Java build — `DB_URL` (with or without a `jdbc:` prefix) plus `DB_USERNAME` and `DB_PASSWORD`.

Run the tests with:

```bash
go test ./...
```

---

## Deploy your own (free)

1. Create a free **[Neon](https://neon.tech)** Postgres project and copy its connection string.
2. Import this repo into **[Vercel](https://vercel.com)** (framework preset: **Other**). `vercel.json` handles the rest.
3. Set the environment variables below.
4. Add your deployed URL to your Google OAuth client's **Authorized JavaScript origins**, or Google sign-in will be rejected. Email/password and guest mode work without this.

### Environment variables

| Variable | Required | Purpose |
|---|---|---|
| `DATABASE_URL` | ✅ | Postgres connection string |
| `JWT_SECRET` | ✅ | Token signing key, **32+ characters** |
| `GOOGLE_CLIENT_ID` | — | Enables Google sign-in |
| `JWT_EXPIRATION_MS` | — | Session lifetime (default 7 days) |
| `MAIL_HOST` · `MAIL_PORT` | — | SMTP for password resets (e.g. `smtp.gmail.com` · `587`) |
| `MAIL_USERNAME` · `MAIL_PASSWORD` | — | SMTP credentials (for Gmail, an App Password) |
| `MAIL_FROM_NAME` | — | Sender name (default `Kairos`) |
| `APP_BASE_URL` | — | Base for reset links; derived from the request when unset |
| `CORS_ALLOWED_ORIGINS` | — | Extra allowed origins, comma-separated |

Without the `MAIL_*` variables the reset link is written to the logs instead of emailed, which is convenient in development.

Tables are created by the schema in `backend/` on first run of the Java app; if you start fresh with the Go backend, create `users`, `user_state` and `password_reset_tokens` first.

---

## API

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| `GET` | `/api/public/config` | public | `{ googleClientId }` |
| `POST` | `/api/auth/google` | public | `{ idToken }` → `{ token, user }` |
| `POST` | `/api/auth/register` | public | `{ email, password, name? }` → `{ token, user }` |
| `POST` | `/api/auth/login` | public | `{ email, password }` → `{ token, user }` |
| `POST` | `/api/auth/forgot-password` | public | `{ email }` → always the same neutral message |
| `POST` | `/api/auth/reset-password` | public | `{ token, newPassword }` → `{ token, user }` |
| `GET` · `PUT` | `/api/state` | Bearer | load / save the goal + journal document |
| `GET` · `PUT` | `/api/profile` | Bearer | read profile / set display name |
| `PUT` | `/api/profile/picture` | Bearer | set a custom avatar |
| `GET` | `/api/share/status` | Bearer | current share state |
| `POST` | `/api/share/enable` · `/api/share/disable` | Bearer | toggle a share link |
| `GET` | `/api/public/share/{token}` | public | read-only snapshot |
| `GET` | `/.well-known/assetlinks.json` | public | Android TWA verification |

`PUT /api/state` takes the **bare** JSON document; both verbs respond with `{ data, updatedAt }`, where `updatedAt` is epoch milliseconds.

---

## License

Personal project — free to use and adapt.
