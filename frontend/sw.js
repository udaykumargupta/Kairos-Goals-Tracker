/* Kairos service worker — makes the app installable + offline-capable.
   Deliberately conservative:
     • only touches same-origin GET requests
     • never touches /api/** (auth + sync must always hit the network)
     • never touches cross-origin (Google Sign-In, fonts, YouTube)
     • documents use network-first, so an online launch always gets the latest deploy
   Bump CACHE when the shell changes to evict old caches. */
const CACHE = "kairos-v1";
const SHELL = ["/", "/manifest.json", "/icons/icon-192.png", "/icons/icon-512.png"];

self.addEventListener("install", (e) => {
  e.waitUntil(
    caches.open(CACHE).then((c) => c.addAll(SHELL)).catch(() => {}).then(() => self.skipWaiting())
  );
});

self.addEventListener("activate", (e) => {
  e.waitUntil(
    caches.keys()
      .then((keys) => Promise.all(keys.filter((k) => k !== CACHE).map((k) => caches.delete(k))))
      .then(() => self.clients.claim())
  );
});

self.addEventListener("fetch", (e) => {
  const req = e.request;
  if (req.method !== "GET") return;                       // leave API writes (POST/PUT) alone
  const url = new URL(req.url);
  if (url.origin !== self.location.origin) return;        // Google/YouTube/fonts → straight to network
  if (url.pathname.startsWith("/api/")) return;           // dynamic + authed → always network

  // Documents: network-first (fresh on every online launch), fall back to cache offline.
  if (req.mode === "navigate") {
    e.respondWith(
      fetch(req)
        .then((res) => { const copy = res.clone(); caches.open(CACHE).then((c) => c.put("/", copy)); return res; })
        .catch(() => caches.match("/").then((r) => r || caches.match("/index.html")))
    );
    return;
  }

  // Other same-origin assets (icons, manifest): cache-first, then fill the cache.
  e.respondWith(
    caches.match(req).then((cached) =>
      cached || fetch(req).then((res) => { const copy = res.clone(); caches.open(CACHE).then((c) => c.put(req, copy)); return res; })
    )
  );
});
