# Kairos — Goals Tracker

> *Kairos* (καιρός): the ancient Greek word for the **right, opportune moment** to act.

A clean, private, single-file web app to plan and track your goals across every time horizon — **today, this month, this year, and any custom date range you define**. It lives entirely in your browser: no accounts, no server, no tracking.

## Features

- **📅 Calendar** — month view with today highlighted; click any day to manage that day's goals.
- **Goals at four scopes** — separate, checkable goal lists for:
  - **Day** — goals for a specific date
  - **Month** — goals for the whole month
  - **Year** — goals for the whole year
  - **⏳ Custom timelines** — your own start→end window with times (e.g. *12 Jun 2026, 12:00 PM → 11 Apr 2027, 1:00 PM*), each with its own goals, progress, and live **Upcoming / Active / Ended** status.
- **✓ / ✗ day completion marks** — reach your daily threshold (default **70%**, adjustable) and the day gets a green **✓**; a past day that fell short gets a red **✗**. Today and future days are never crossed.
- **Progress bars** for every scope and timeline.
- **Quick editing** — double-click a goal to rename, hover to delete.
- **Light / dark theme** — follows your system, with a manual toggle.
- **Private by design** — everything is stored in your browser's `localStorage`.
- **Backup & restore** — export all your goals to a JSON file and import it back anytime.

## Usage

No build step, no dependencies. Just open the file:

```bash
open index.html          # macOS
# or double-click index.html in your file manager
# or serve it: python3 -m http.server 4599  →  http://localhost:4599
```

Your data is saved automatically as you go.

> **Note:** Data is tied to the specific browser and profile you use. To move your goals to another browser or machine, use **Export backup** → **Import**.

## Tech

A single self-contained `index.html` — vanilla HTML, CSS, and JavaScript. No frameworks, no external requests. Works fully offline.

## Roadmap ideas

- Carry-over of unfinished goals to the next day
- Streak counter for consecutive successful days
- Timeline spans drawn onto the calendar

## License

Personal project — free to use and adapt.
