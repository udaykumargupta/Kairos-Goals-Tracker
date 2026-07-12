# Kairos — Goals Tracker

> *Kairos* (καιρός): the ancient Greek word for the **right, opportune moment** to act.

**🔗 Live: https://kairos-7yt7.onrender.com**

Track goals across every time horizon — **today, this month, this year, and any custom date range** — with a calendar, journal, and shareable progress. Sign in with Google to sync across devices, or use it as a guest with everything saved locally in your browser.

**Stack:** Java 17 · Spring Boot · Spring Security (JWT) · PostgreSQL · vanilla JS · Docker — deployed on Render + Neon.

## Features

- **📅 Calendar** — click a day to manage its goals, **⭐ bookmark** dates, and get **✓ / ✗ completion marks** when a day clears (or misses) your threshold.
- **Goals at four scopes** — Day, Month, Year, and **⏳ custom timelines** (your own dated window with a live *Upcoming / Active / Ended* status).
- **📓 Journal** for each day, month, year, and timeline — with a full-page writing mode.
- **⏱ Time meters** — how much of each period (and timeline) has elapsed vs. remaining.
- **Google sign-in or guest mode** — per-user cloud sync with an offline cache, or fully local.
- **🔗 Read-only share links**, a **custom display name & avatar**, quote of the day, and light / dark theme.

## Architecture

```
Browser (frontend/index.html)
  │  Google Identity Services → Google ID token
  │  POST /api/auth/google  →  verify token → upsert user → issue app JWT
  │  Authorization: Bearer <jwt>
  ├─ /api/state           load / save goal + journal JSON  (PostgreSQL jsonb)
  ├─ /api/profile[/picture]  display name / avatar
  └─ /api/share…          enable a token; /?share=<token> renders a read-only view
```

Thin controllers → service interfaces → Spring Data repositories; each class has one responsibility and depends on abstractions, so the DB or token verifier can be swapped without touching callers. Maven bundles `frontend/` into the app's static content, so the UI is served same-origin (no CORS, and a valid origin for Google Sign-In).

## Project structure

```
Kairos/
├── frontend/index.html     # the single-file web app
├── backend/                # Spring Boot (auth · user · state · share · security · config)
├── Dockerfile              # builds app + bundles frontend (used by Render)
├── render.yaml             # Render Blueprint
├── docker-compose.yml      # local PostgreSQL
└── DEPLOY.md               # free deploy guide (Render + Neon)
```

## Run locally

Prerequisites: JDK 17+ and Docker (the `./mvnw` wrapper means no global Maven).

```bash
# 1. Secrets — set GOOGLE_CLIENT_ID and JWT_SECRET (see .env.example)
cp .env.example .env

# 2. PostgreSQL
docker run -d --name kairos-postgres \
  -e POSTGRES_DB=kairos -e POSTGRES_USER=kairos -e POSTGRES_PASSWORD=kairos \
  -p 5432:5432 -v kairos_pgdata:/var/lib/postgresql/data postgres:16

# 3. Backend (also serves the frontend)
cd backend
export $(grep -v '^#' ../.env | xargs)
./mvnw spring-boot:run
```

Open <http://localhost:8080> → **Sign in with Google**, or **Continue as guest** (no Google setup needed). Tables are auto-created on first boot.

For a Google Client ID and full self-host steps, see **[DEPLOY.md](DEPLOY.md)**.

## Deploy your own (free)

**[DEPLOY.md](DEPLOY.md)** walks through **Render** (free web service via `Dockerfile`) + **Neon** (free PostgreSQL): push to GitHub → Render **New → Blueprint** → set env vars → add your live URL to Google's Authorized JavaScript origins and publish the consent screen.

## API

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| `GET` | `/api/public/config` | public | `{ googleClientId }` |
| `POST` | `/api/auth/google` | public | `{ idToken }` → `{ token, user }` |
| `GET` · `PUT` | `/api/state` | Bearer | load / save goal + journal JSON |
| `GET` · `PUT` | `/api/profile` | Bearer | read profile / set display name |
| `PUT` | `/api/profile/picture` | Bearer | set custom avatar |
| `GET` | `/api/share/status` | Bearer | current share state |
| `POST` | `/api/share/enable` · `/api/share/disable` | Bearer | toggle a share link |
| `GET` | `/api/public/share/{token}` | public | read-only snapshot |

## License

Personal project — free to use and adapt.
