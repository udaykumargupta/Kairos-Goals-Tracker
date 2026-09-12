package kairos

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// maxBodyBytes caps request bodies. The state document carries embedded photos and
// voice notes as data URLs, so it needs generous headroom.
const maxBodyBytes = 24 << 20 // 24 MB

type ctxKey int

const principalKey ctxKey = iota

// Routes wires every endpoint. Path/method patterns mirror the Spring controllers.
func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()

	// --- public auth ---
	mux.HandleFunc("POST /api/auth/google", a.handleGoogleLogin)
	mux.HandleFunc("POST /api/auth/register", a.handleRegister)
	mux.HandleFunc("POST /api/auth/login", a.handleLogin)
	mux.HandleFunc("POST /api/auth/forgot-password", a.handleForgotPassword)
	mux.HandleFunc("POST /api/auth/reset-password", a.handleResetPassword)

	// --- public config / share ---
	mux.HandleFunc("GET /api/public/config", a.handlePublicConfig)
	mux.HandleFunc("GET /api/public/share/{token}", a.handlePublicShare)
	mux.HandleFunc("GET /.well-known/assetlinks.json", a.handleAssetLinks)

	// --- authenticated ---
	mux.HandleFunc("GET /api/profile", a.authed(a.handleGetProfile))
	mux.HandleFunc("PUT /api/profile", a.authed(a.handleUpdateProfile))
	mux.HandleFunc("PUT /api/profile/picture", a.authed(a.handleUpdatePicture))
	mux.HandleFunc("GET /api/state", a.authed(a.handleGetState))
	mux.HandleFunc("PUT /api/state", a.authed(a.handleSaveState))
	mux.HandleFunc("GET /api/share/status", a.authed(a.handleShareStatus))
	mux.HandleFunc("POST /api/share/enable", a.authed(a.handleShareEnable))
	mux.HandleFunc("POST /api/share/disable", a.authed(a.handleShareDisable))

	// Anything reaching the API that matched no route above. Logging the path makes
	// a platform routing misconfiguration obvious instead of a silent 404.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("unrouted request: %s %s", r.Method, r.URL.Path)
		notFound(w, "No handler for "+r.URL.Path)
	})

	return a.withCORS(mux)
}

// withCORS applies the same policy the Spring config did: only /api/**, the caller's
// Origin echoed back when it matches an allowed pattern, and no credentials (auth is
// a bearer token, not a cookie).
func (a *App) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			if origin := r.Header.Get("Origin"); origin != "" && a.originAllowed(origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Set("Access-Control-Max-Age", "1800")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (a *App) originAllowed(origin string) bool {
	for _, pattern := range a.Cfg.AllowedOrigins {
		if matchOrigin(pattern, origin) {
			return true
		}
	}
	return false
}

// matchOrigin supports one leading "*." wildcard in the host, e.g. https://*.vercel.app.
func matchOrigin(pattern, origin string) bool {
	if pattern == origin || pattern == "*" {
		return true
	}
	star := strings.Index(pattern, "*")
	if star < 0 {
		return false
	}
	prefix, suffix := pattern[:star], pattern[star+1:]
	return len(origin) >= len(prefix)+len(suffix) &&
		strings.HasPrefix(origin, prefix) && strings.HasSuffix(origin, suffix)
}

// authed requires a valid bearer token, answering with the same 401 body the Spring
// entry point produced when it is missing or invalid.
func (a *App) authed(h func(http.ResponseWriter, *http.Request, Principal)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			unauthorized(w)
			return
		}
		principal, ok := a.VerifyJWT(strings.TrimPrefix(header, "Bearer "))
		if !ok {
			unauthorized(w)
			return
		}
		h(w, r.WithContext(context.WithValue(r.Context(), principalKey, principal)), principal)
	}
}

// ---- request/response helpers -------------------------------------------

func decodeJSON(r *http.Request, dst any) error {
	return json.NewDecoder(io.LimitReader(r.Body, maxBodyBytes)).Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// unauthorized matches SecurityConfig.unauthorizedEntryPoint().
func unauthorized(w http.ResponseWriter) {
	writeJSON(w, http.StatusUnauthorized, map[string]string{
		"error":   "unauthorized",
		"message": "Authentication required",
	})
}

// apiError matches GlobalExceptionHandler's LinkedHashMap body:
// {timestamp, status, error, message}.
func apiError(w http.ResponseWriter, status int, label, message string) {
	writeJSON(w, status, map[string]any{
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"status":    status,
		"error":     label,
		"message":   message,
	})
}

func badRequest(w http.ResponseWriter, msg string)   { apiError(w, http.StatusBadRequest, "Bad Request", msg) }
func authFailed(w http.ResponseWriter, msg string)   { apiError(w, http.StatusUnauthorized, "Unauthorized", msg) }
func notFound(w http.ResponseWriter, msg string)     { apiError(w, http.StatusNotFound, "Not Found", msg) }

// serverError logs the cause and returns a generic message (never leaks internals).
func serverError(w http.ResponseWriter, err error) {
	log.Printf("internal error: %v", err)
	apiError(w, http.StatusInternalServerError, "Internal Server Error", "Something went wrong")
}

// baseURL resolves the public origin for links in emails: the configured value if
// set, else reconstructed from the request, honouring the proxy's forwarded headers
// (the Go equivalent of Spring's forward-headers-strategy: framework).
func (a *App) baseURL(r *http.Request) string {
	if a.Cfg.AppBaseURL != "" {
		return a.Cfg.AppBaseURL
	}
	scheme := "http"
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = strings.TrimSpace(strings.Split(proto, ",")[0])
	} else if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	if fwd := r.Header.Get("X-Forwarded-Host"); fwd != "" {
		host = strings.TrimSpace(strings.Split(fwd, ",")[0])
	}
	return scheme + "://" + host
}
