# Kairos — Goals Tracker

> *Kairos* (καιρός): the ancient Greek word for the **right, opportune moment** to act.

**🔗 Live app: https://kairos-7yt7.onrender.com**

Plan and track goals across every time horizon — **today, this month, this year, and any custom date range** — with a journal, a calendar, and progress you can share. Sign in with Google to sync across devices, or use it as a guest with everything saved locally.

Kairos is a **Spring Boot** backend (Google sign-in → JWT → per-user JSON store in PostgreSQL) that also serves a **single-file vanilla-JS frontend**, deployed on **Render + Neon**.

## Features

- **📅 Calendar** — today highlighted, click any day to manage its goals, **⭐ bookmark** important dates.
- **Goals at four scopes** — **Day**, **Month**, **Year**, and **⏳ Custom timelines** (your own start→end window with times, each with goals, progress, and a live *Upcoming / Active / Ended* status).
- **📓 Journaling** — free-text notes for each day, month, year, and timeline, with a **full-page** writing mode.
- **⏱ Time meters** — how much of the day/month/year (and each timeline) has elapsed vs. remaining.
- **✓ / ✗ completion marks** — reach your daily threshold (default 70%, adjustable) → green ✓; a past day below it → red ✗.
- **Google sign-in *or* guest mode** — sync to your account, or keep everything in this browser (upgradeable later).
- **🔗 Share progress** — generate a read-only link so anyone can view your progress.
- **Custom profile** — set a display name and upload an avatar that override your Google profile.
- **Offline-friendly** — localStorage caches your goals; changes sync when online (Saved / Saving / Offline indicator).
- **Quote of the day**, light / dark theme, and a fixed-height layout where only the goal lists and journal scroll.

## Architecture

```
Browser (frontend/index.html)
  │  Google Identity Services → Google ID token
  │  POST /api/auth/google  ──►  verify ID token ─► upsert User ─► issue app JWT
  │  Authorization: Bearer <jwt>
  ├─ GET/PUT /api/state          load / save this user's goal + journal JSON  (PostgreSQL jsonb)
  ├─ GET/PUT /api/profile        read profile / set custom display name
  ├─ PUT     /api/profile/picture set custom avatar
  └─ share:  enable/disable a token; /?share=<token> renders a read-only view
```

SOLID layering: thin controllers → service **interfaces** (`AuthService`, `UserService`, `UserStateService`, `ShareService`, `GoogleTokenVerifier`, `JwtService`) → Spring Data repositories. Each class has one responsibility and depends on abstractions, so the DB or token-verification mechanism can be swapped without touching callers.

## Project structure

```
Kairos/
├── frontend/index.html           # the single-file web app (source of truth)
├── backend/                      # standard Spring Boot project (Maven wrapper included)
│   ├── pom.xml
│   └── src/main/java/com/kairos/
│       ├── auth/                 # Google verification + login
│       ├── user/                 # User entity, repo, service, profile endpoints
│       ├── state/                # per-user goal/journal document (jsonb) API
│       ├── share/                # read-only progress sharing
│       ├── security/             # JWT service + filter + Spring Security config
│       ├── config/               # typed properties + public config endpoint
│       └── common/               # exception handling
├── Dockerfile                    # builds app + bundles frontend (used by Render)
├── render.yaml                   # Render Blueprint
├── docker-compose.yml            # local PostgreSQL
├── DEPLOY.md                     # free deploy guide (Render + Neon)
└── .env.example
```

At build time Maven copies `frontend/` into the app's static content, so Spring serves the UI same-origin (no CORS, and a valid origin for Google Sign-In). Edit the UI in `frontend/index.html`; rebuild to pick up changes.

## Run locally

**Prerequisites:** JDK 17+ (`brew install openjdk@21`), Docker. The `./mvnw` wrapper means no global Maven is needed.

1. **Google OAuth Client ID** — [console.cloud.google.com](https://console.cloud.google.com/) → APIs & Services → Credentials → **Create OAuth client ID** → **Web application** → add `http://localhost:8080` to *Authorized JavaScript origins* → copy the Client ID. *(No client secret needed — we verify Google's ID token.)*

2. **Secrets:**
   ```bash
   cp .env.example .env
   # set GOOGLE_CLIENT_ID and JWT_SECRET (openssl rand -base64 48)
   ```

3. **PostgreSQL:**
   ```bash
   docker run -d --name kairos-postgres \
     -e POSTGRES_DB=kairos -e POSTGRES_USER=kairos -e POSTGRES_PASSWORD=kairos \
     -p 5432:5432 -v kairos_pgdata:/var/lib/postgresql/data postgres:16
   ```

4. **Backend (serves the frontend too):**
   ```bash
   cd backend
   export $(grep -v '^#' ../.env | xargs)
   ./mvnw spring-boot:run
   ```

Open **<http://localhost:8080>** → sign in with Google, or **Continue as guest**. Tables are auto-created on first boot (`ddl-auto: update`).

> Prefer not to sign in at all locally? Just click **Continue as guest** — no Google setup required.

## Deploy your own (free)

See **[DEPLOY.md](DEPLOY.md)** for a step-by-step guide using **Render** (free web service via `Dockerfile`) + **Neon** (free PostgreSQL). In short: push to GitHub → Render **New → Blueprint** (reads `render.yaml`) → set env vars → add your live URL to Google's Authorized JavaScript origins and publish the consent screen.

## Configuration reference

| Key / env | Default | Purpose |
|-----------|---------|---------|
| `GOOGLE_CLIENT_ID` | — | OAuth Client ID; token audience |
| `JWT_SECRET` | dev secret | HMAC signing key (**≥32 chars**) |
| `JWT_EXPIRATION_MS` | `604800000` (7d) | JWT lifetime |
| `DB_URL` / `DB_USERNAME` / `DB_PASSWORD` | localhost / kairos / kairos | PostgreSQL connection |
| `PORT` | `8080` | HTTP port (injected by Render) |

## API

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| `GET`  | `/api/public/config` | public | `{ googleClientId }` |
| `POST` | `/api/auth/google` | public | `{ idToken }` → `{ token, user }` |
| `GET` / `PUT` | `/api/state` | Bearer | load / save goal + journal JSON |
| `GET` / `PUT` | `/api/profile` | Bearer | read profile / set display name |
| `PUT`  | `/api/profile/picture` | Bearer | set custom avatar |
| `GET`  | `/api/share/status` | Bearer | current share state |
| `POST` | `/api/share/enable` · `/api/share/disable` | Bearer | toggle a share link |
| `GET`  | `/api/public/share/{token}` | public | read-only snapshot for a share link |

## Notes & roadmap

- `ddl-auto: update` is convenient here; adopt Flyway/Liquibase migrations for heavier production use.
- Free hosting note: Render's free instance sleeps when idle, so the first request after a quiet period takes ~30–60s to wake.
- Ideas: carry unfinished goals to the next day, streak counter, timeline spans drawn on the calendar.

## License

Personal project — free to use and adapt.
