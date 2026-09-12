package kairos

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Config is every runtime setting, read from the same environment variables the
// Spring backend used — so an existing Railway/Render/Neon setup moves across unchanged.
type Config struct {
	DatabaseURL    string
	JWTSecret      []byte
	JWTExpiryMS    int64
	GoogleClientID string

	MailHost     string
	MailPort     int
	MailUsername string
	MailPassword string
	MailFrom     string
	MailFromName string

	AppBaseURL     string   // public base URL for links in emails; blank = derive from request
	AllowedOrigins []string // CORS patterns; "*" allowed once inside the host
}

func env(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

// LoadConfig mirrors application.yml's defaults.
func LoadConfig() (Config, error) {
	c := Config{
		// NOTE: the Java side used secret.getBytes(UTF_8) directly (no base64 decode),
		// so we must use the raw bytes of the same string to stay token-compatible.
		JWTSecret:      []byte(env("JWT_SECRET", "kairos-dev-secret-change-me-please-0123456789")),
		GoogleClientID: env("GOOGLE_CLIENT_ID", ""),
		MailHost:       env("MAIL_HOST", ""),
		MailUsername:   env("MAIL_USERNAME", ""),
		MailPassword:   env("MAIL_PASSWORD", ""),
		MailFromName:   env("MAIL_FROM_NAME", "Kairos"),
		AppBaseURL:     strings.TrimSuffix(env("APP_BASE_URL", ""), "/"),
	}
	c.MailFrom = env("MAIL_FROM", c.MailUsername)
	c.MailPort, _ = strconv.Atoi(env("MAIL_PORT", "587"))

	// Same defaults the Spring CORS config used, plus Vercel. The app is served
	// same-origin in production, so these mostly matter for local development.
	c.AllowedOrigins = []string{
		"http://localhost:8080", "http://127.0.0.1:8080",
		"http://localhost:5500", "http://127.0.0.1:5500",
		"http://localhost:3000",
		"https://*.onrender.com", "https://*.up.railway.app", "https://*.vercel.app",
	}
	for _, extra := range strings.Split(env("CORS_ALLOWED_ORIGINS", ""), ",") {
		if v := strings.TrimSpace(extra); v != "" {
			c.AllowedOrigins = append(c.AllowedOrigins, v)
		}
	}

	ms, err := strconv.ParseInt(env("JWT_EXPIRATION_MS", "604800000"), 10, 64) // 7 days
	if err != nil {
		return c, fmt.Errorf("invalid JWT_EXPIRATION_MS: %w", err)
	}
	c.JWTExpiryMS = ms

	dsn, err := databaseDSN()
	if err != nil {
		return c, err
	}
	c.DatabaseURL = dsn
	return c, nil
}

// databaseDSN accepts either a ready-made Postgres URL (DATABASE_URL / POSTGRES_URL,
// e.g. straight from the Neon dashboard) or the Java-style DB_URL + DB_USERNAME +
// DB_PASSWORD trio, and returns a DSN pgx understands.
func databaseDSN() (string, error) {
	if raw := env("DATABASE_URL", env("POSTGRES_URL", "")); raw != "" {
		return sanitizeDSN(raw)
	}

	raw := env("DB_URL", "")
	if raw == "" {
		return "", fmt.Errorf("no database configured: set DATABASE_URL, or DB_URL + DB_USERNAME + DB_PASSWORD")
	}
	// "jdbc:postgresql://host/db?x" -> "postgresql://host/db?x"
	raw = strings.TrimPrefix(raw, "jdbc:")

	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid DB_URL: %w", err)
	}
	// The JDBC form keeps credentials in separate variables; fold them into the URL.
	if user := env("DB_USERNAME", ""); user != "" {
		if pw := env("DB_PASSWORD", ""); pw != "" {
			u.User = url.UserPassword(user, pw)
		} else {
			u.User = url.User(user)
		}
	}
	return sanitizeDSN(u.String())
}

// sanitizeDSN drops query parameters pgx does not understand. Neon's copy-paste
// string carries channel_binding, which is a libpq-only flag; sslmode already
// guarantees the connection is encrypted.
func sanitizeDSN(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid database URL: %w", err)
	}
	if u.Scheme == "" {
		return "", fmt.Errorf("database URL must start with postgres:// or postgresql://")
	}
	q := u.Query()
	q.Del("channel_binding")
	u.RawQuery = q.Encode()
	return u.String(), nil
}
