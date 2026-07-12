# Deploying Kairos (free) — Render + Neon

Kairos runs as one Docker web service (Spring Boot serving the frontend) plus a PostgreSQL database. This guide uses **Render** (free web service) for the app and **Neon** (free Postgres) for the database.

> Heads-up: Render's free instance **sleeps after ~15 min idle**, so the first visit after a quiet period takes ~30–60s to wake. That's normal for free Java hosting.

---

## 0. Push the repo to GitHub

Render deploys from GitHub, so make sure your latest commits are pushed:

```bash
cd /Users/uday.g/Desktop/services/Kairos
git push
```

(Username `udaykumargupta`, password = a Personal Access Token.)

---

## 1. Create the database on Neon

1. Sign up at **https://neon.tech** (free).
2. **Create a project** (any name, e.g. `kairos`). It creates a database and a role.
3. Open **Dashboard → Connect** (or "Connection Details"). You'll see a connection string like:
   ```
   postgresql://myuser:mypassword@ep-cool-name-123456.us-east-2.aws.neon.tech/neondb?sslmode=require
   ```
4. From that string, note these three pieces — you'll paste them into Render next:

   | Render env var | Value (from the Neon string) |
   |----------------|------------------------------|
   | `DB_URL` | `jdbc:postgresql://ep-cool-name-123456.us-east-2.aws.neon.tech/neondb?sslmode=require` &nbsp;← the host + `/db?sslmode=require`, prefixed with **`jdbc:`** |
   | `DB_USERNAME` | `myuser` |
   | `DB_PASSWORD` | `mypassword` |

   > Note the `DB_URL` is the Neon URL with the `postgresql://user:pass@` part replaced by **`jdbc:postgresql://`** (host onward), keeping `?sslmode=require`.

---

## 2. Deploy the app on Render

1. Sign up at **https://render.com** (free), and connect your GitHub.
2. **New → Blueprint**, pick the `Kairos-Goals-Tracker` repo. Render reads `render.yaml` and creates the `kairos` web service.
   *(Or: New → Web Service → the repo → Runtime: Docker.)*
3. When prompted, fill the environment variables:
   - `GOOGLE_CLIENT_ID` = `837234044534-q2h5fnmfplaug34tksav866avfp4stie.apps.googleusercontent.com`
   - `JWT_SECRET` = leave as the auto-generated value (or set your own 32+ char string)
   - `DB_URL`, `DB_USERNAME`, `DB_PASSWORD` = the three values from Neon (step 1)
4. Click **Create / Deploy**. First build takes a few minutes (it compiles the app in Docker).
5. When it's live you'll get a URL like **`https://kairos-xxxx.onrender.com`**. Copy it.

The app auto-creates its tables in Neon on first boot.

---

## 3. Let anyone sign in with Google

Two settings in **[Google Cloud Console](https://console.cloud.google.com/apis/credentials)** for the project that owns your Client ID:

**a) Authorize the live origin** (Credentials → your OAuth Client ID → *Authorized JavaScript origins* → **+ Add URI**):
```
https://kairos-xxxx.onrender.com
```
*(exactly your Render URL — `https`, no trailing slash. Keep `http://localhost:8080` too for local dev.)*

**b) Publish the consent screen** so it's not limited to test users (**APIs & Services → OAuth consent screen**):
- If it's in **Testing**, either add each person under **Test users**, or click **Publish app** → **Confirm** to make it available to everyone.
- Kairos only uses basic **email / profile** scopes (non-sensitive), so publishing needs **no Google verification**. New users may briefly see an "unverified app" notice — that's expected and safe to continue.

Changes can take a minute or two to propagate.

---

## 4. Use it

Open **`https://kairos-xxxx.onrender.com`**, sign in with Google (or continue as guest), and share your progress link — it'll now work for anyone on the internet.

---

## Environment variables reference

| Var | Purpose |
|-----|---------|
| `GOOGLE_CLIENT_ID` | Google OAuth Web Client ID (token audience) |
| `JWT_SECRET` | App JWT signing secret (≥32 chars) |
| `DB_URL` | `jdbc:postgresql://<host>/<db>?sslmode=require` |
| `DB_USERNAME` / `DB_PASSWORD` | Neon role credentials |
| `PORT` | Injected by Render automatically |

## Notes
- Free Render instance sleeps when idle → occasional cold starts. Upgrading to a paid instance removes this.
- Neon's free tier also auto-suspends the DB when idle; it resumes on the next query (adds a moment to the first request).
- To deploy updates later: just `git push` — Render redeploys automatically.
