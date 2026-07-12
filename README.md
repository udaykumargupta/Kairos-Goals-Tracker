# Kairos — Goals Tracker

> *Kairos* (καιρός): the ancient Greek word for the **right, opportune moment** to act.

Plan and track goals across every time horizon — **today, this month, this year, and any custom date range** — with your data synced privately to your Google account.

Kairos is a **Spring Boot** backend (Google sign-in → JWT → per-user JSON store in PostgreSQL) that also serves a **single-file vanilla-JS frontend**. Everything runs same-origin at `http://localhost:8080`.

## Features

- **📅 Calendar** with today highlighted; click any day to manage its goals.
- **Goals at four scopes** — **Day**, **Month**, **Year**, and **⏳ Custom timelines** (your own start→end window with times, each with goals, progress, and a live *Upcoming / Active / Ended* status).
- **✓ / ✗ day completion marks** — hit your daily threshold (default 70%, adjustable) → green ✓; a past day below it → red ✗.
- **Google (Gmail) sign-in** — your goals are stored against your Google account and follow you across browsers/devices.
- **Offline-friendly** — localStorage caches your goals; changes sync to the server when you're online (with a Saved / Saving / Offline indicator).
- **Light / dark theme**, JSON backup/restore.

## Architecture

```
Browser (frontend/index.html)
  │  Google Identity Services → Google ID token
  │  POST /api/auth/google  ──►  verify ID token ─► upsert User ─► issue app JWT
  │  Authorization: Bearer <jwt>
  ├─ GET  /api/state   ──► load this user's goal JSON
  └─ PUT  /api/state   ──► save this user's goal JSON   (PostgreSQL jsonb)
```

SOLID layering: thin controllers → service **interfaces** (`AuthService`, `UserService`, `UserStateService`, `GoogleTokenVerifier`, `JwtService`) → Spring Data repositories. Each class has one responsibility; collaborators are injected via constructors and depend on abstractions, so the DB or token-verification mechanism can be swapped without touching callers.

## Project structure

```
Kairos/
├── frontend/                     # the single-file web app (source of truth)
│   └── index.html
├── backend/                      # standard Spring Boot project
│   ├── mvnw, mvnw.cmd, .mvn/     # Maven wrapper — no global Maven needed
│   ├── pom.xml
│   └── src/main/java/com/kairos/
│       ├── auth/                 # Google verification + login (controller, service, dto)
│       ├── user/                 # User entity, repository, service
│       ├── state/                # per-user goal document (jsonb) API
│       ├── security/             # JWT service + filter + Spring Security config
│       ├── config/               # typed properties + public config endpoint
│       └── common/               # exception handling
├── docker-compose.yml            # PostgreSQL
├── .env.example
└── README.md
```

At build time Maven copies `frontend/` into the app's static content, so Spring serves the UI same-origin at `http://localhost:8080` (no CORS, and a valid origin for Google Sign-In). Edit the UI in `frontend/index.html`; rebuild/restart to pick up changes.

## Prerequisites

- **JDK 17+** — `brew install openjdk@21` (Maven is **not** required; the `./mvnw` wrapper handles it)
- **Docker** (for PostgreSQL)
- A **Google OAuth Client ID** (see below)

## 1) Create a Google OAuth Client ID

1. Go to <https://console.cloud.google.com/> → create/select a project.
2. **APIs & Services → OAuth consent screen**: choose **External**, fill app name + your email, add yourself as a **Test user**, save.
3. **APIs & Services → Credentials → Create credentials → OAuth client ID**.
4. **Application type: Web application**.
5. Under **Authorized JavaScript origins**, add:
   - `http://localhost:8080`
6. Create → copy the **Client ID** (looks like `1234-abcd.apps.googleusercontent.com`).
   *(No client secret is needed — we verify Google's ID token, we don't do a server-side code exchange.)*

## 2) Configure secrets

```bash
cp .env.example .env
# edit .env:
#   GOOGLE_CLIENT_ID=<your client id>
#   JWT_SECRET=<random 32+ char string>   e.g. openssl rand -base64 48
```

## 3) Start PostgreSQL

With Docker Compose:

```bash
docker compose up -d          # or: docker-compose up -d
```

No compose plugin? Use plain Docker:

```bash
docker run -d --name kairos-postgres \
  -e POSTGRES_DB=kairos -e POSTGRES_USER=kairos -e POSTGRES_PASSWORD=kairos \
  -p 5432:5432 -v kairos_pgdata:/var/lib/postgresql/data postgres:16
```

## 4) Run the backend (serves the frontend too)

```bash
cd backend
export $(grep -v '^#' ../.env | xargs)     # load GOOGLE_CLIENT_ID / JWT_SECRET
./mvnw spring-boot:run                      # first run downloads Maven + deps
```

Then open **<http://localhost:8080>**, click **Sign in with Google**, and your goals will sync.

> First run auto-creates the `users` and `user_state` tables (`ddl-auto: update`).

## Configuration reference (`application.yml`)

| Key / env | Default | Purpose |
|-----------|---------|---------|
| `GOOGLE_CLIENT_ID` | — | OAuth Client ID; token audience |
| `JWT_SECRET` | dev secret | HMAC signing key (**≥32 chars**) |
| `JWT_EXPIRATION_MS` | `604800000` (7d) | JWT lifetime |
| `DB_URL` / `DB_USERNAME` / `DB_PASSWORD` | localhost / kairos / kairos | PostgreSQL connection |

## API

| Method | Path | Auth | Body / Result |
|--------|------|------|---------------|
| `GET`  | `/api/public/config` | public | `{ googleClientId }` |
| `POST` | `/api/auth/google`   | public | `{ idToken }` → `{ token, user }` |
| `GET`  | `/api/state`         | Bearer JWT | `{ data, updatedAt }` |
| `PUT`  | `/api/state`         | Bearer JWT | goal JSON → `{ data, updatedAt }` |

## Notes & roadmap

- `ddl-auto: update` is convenient for a personal app; adopt Flyway/Liquibase migrations before any production/multi-user use.
- Ideas: carry over unfinished goals to the next day, streak counter, timeline spans drawn on the calendar.

## License

Personal project — free to use and adapt.
