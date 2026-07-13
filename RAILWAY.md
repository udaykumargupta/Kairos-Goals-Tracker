# Deploying Kairos on Railway (same database as Render)

This deploys Kairos as a second instance on **[Railway](https://railway.app)** that talks to the **exact same Neon PostgreSQL** your Render instance uses — so existing users sign in and see all their data. Both deployments read and write the same database.

> **The one rule that keeps data safe:** do **not** add a Railway PostgreSQL plugin. Instead, set `DB_URL` / `DB_USERNAME` / `DB_PASSWORD` to the same Neon values already in Render. A Railway database would be a fresh, empty DB — that's how you'd "lose" the existing users.

> **Cost note:** Railway no longer has a permanent free tier — new accounts get a one-time trial credit, then it's the Hobby plan (~$5/mo). Render stays free. This just adds Railway alongside Render; both hit the same Neon DB.

---

## 1. Copy the current secrets from Render

Open your Render service → **Environment** tab and copy these values (you'll paste them into Railway in step 3):

| Variable | Where it comes from |
|----------|---------------------|
| `GOOGLE_CLIENT_ID` | `837234044534-q2h5fnmfplaug34tksav866avfp4stie.apps.googleusercontent.com` |
| `JWT_SECRET` | copy Render's value (reusing it keeps both instances consistent) |
| `DB_URL` | copy Render's value **exactly** (this is the Neon JDBC URL) |
| `DB_USERNAME` | copy Render's value exactly |
| `DB_PASSWORD` | copy Render's value exactly |

`DB_URL` looks like `jdbc:postgresql://ep-xxxx.aws.neon.tech/neondb?sslmode=require`. Copying it verbatim is what makes Railway use the same database.

---

## 2. Create the Railway service from GitHub

1. Sign in at **https://railway.app** and click **New Project → Deploy from GitHub repo**.
2. Pick the **`Kairos-Goals-Tracker`** repo. Railway reads `railway.json` and builds the `Dockerfile` automatically.
3. Let the first build start — it'll fail the healthcheck until the variables are set (next step). That's expected.

---

## 3. Set the environment variables

In the service → **Variables** tab, add the five variables from step 1 (Railway injects `PORT` itself — don't set it):

```
GOOGLE_CLIENT_ID = 837234044534-q2h5fnmfplaug34tksav866avfp4stie.apps.googleusercontent.com
JWT_SECRET       = <same value as Render>
DB_URL           = <same Neon JDBC URL as Render>
DB_USERNAME      = <same as Render>
DB_PASSWORD      = <same as Render>
```

Save — Railway redeploys with the new variables.

---

## 4. Give it a public URL

Service → **Settings → Networking → Generate Domain**. You'll get a URL like:

```
https://kairos-production-xxxx.up.railway.app
```

Once the deploy is healthy (`/api/public/config` returns 200), open it — guest mode should work immediately. Google Sign-In needs one more step.

---

## 5. Authorize the Railway URL for Google Sign-In

In **[Google Cloud Console](https://console.cloud.google.com/apis/credentials)** → your OAuth Client ID → **Authorized JavaScript origins → + Add URI**:

```
https://kairos-production-xxxx.up.railway.app
```

(exact Railway URL — `https`, no trailing slash. Keep the Render URL and `http://localhost:8080` too.) Changes take a minute or two to propagate. After that, sign in with Google on the Railway URL and your existing goals/journal load straight from Neon.

---

## Verify data is shared (not lost)

Sign in with the same Google account on both the Render URL and the Railway URL — you should see identical goals. Add a goal on one, refresh the other: it appears, because both read the same Neon database.

## Notes

- Neon's free tier auto-suspends when idle and resumes on the next query (adds a moment to the first request), same as with Render.
- Because both instances share one DB, schema is already created — Railway won't re-create tables, it just connects.
- To deploy updates later: `git push` — Railway redeploys automatically, just like Render.
