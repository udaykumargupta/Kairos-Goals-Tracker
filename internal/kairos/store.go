package kairos

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// App carries configuration plus the database client.
type App struct {
	Cfg Config
	DB  *Neon
}

func New(cfg Config) (*App, error) {
	db, err := NewNeon(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	return &App{Cfg: cfg, DB: db}, nil
}

// Ids are cast to text and NULLable text is flattened with COALESCE, so every
// column has one predictable JSON representation coming back over HTTP.
const userCols = `id::text AS id, google_sub, email,
	COALESCE(name,'') AS name, COALESCE(display_name,'') AS display_name,
	COALESCE(picture,'') AS picture, COALESCE(custom_picture,'') AS custom_picture,
	COALESCE(password_hash,'') AS password_hash, COALESCE(share_token,'') AS share_token`

func scanUser(r Row) *User {
	if r == nil {
		return nil
	}
	return &User{
		ID:            r.Int64("id"),
		GoogleSub:     r.Str("google_sub"),
		Email:         r.Str("email"),
		Name:          r.Str("name"),
		DisplayName:   r.Str("display_name"),
		Picture:       r.Str("picture"),
		CustomPicture: r.Str("custom_picture"),
		PasswordHash:  r.Str("password_hash"),
		ShareToken:    r.Str("share_token"),
	}
}

// userOne runs a single-user lookup; a nil *User means "no such row".
func (a *App) userOne(ctx context.Context, where string, args ...any) (*User, error) {
	row, err := a.DB.QueryOne(ctx, `SELECT `+userCols+` FROM users WHERE `+where, args...)
	if err != nil {
		return nil, err
	}
	return scanUser(row), nil
}

func (a *App) UserByID(ctx context.Context, id int64) (*User, error) {
	return a.userOne(ctx, `id = $1`, id)
}

// UserByEmail mirrors findFirstByEmailIgnoreCaseOrderByIdAsc: email is NOT unique
// (registering locally and later signing in with Google creates a second row), so
// the lowest id deterministically wins.
func (a *App) UserByEmail(ctx context.Context, email string) (*User, error) {
	return a.userOne(ctx, `lower(email) = lower($1) ORDER BY id ASC LIMIT 1`, email)
}

func (a *App) UserByGoogleSub(ctx context.Context, sub string) (*User, error) {
	return a.userOne(ctx, `google_sub = $1`, sub)
}

func (a *App) UserByShareToken(ctx context.Context, token string) (*User, error) {
	return a.userOne(ctx, `share_token = $1`, token)
}

func (a *App) EmailExists(ctx context.Context, email string) (bool, error) {
	row, err := a.DB.QueryOne(ctx,
		`SELECT EXISTS (SELECT 1 FROM users WHERE lower(email) = lower($1)) AS found`, email)
	if err != nil {
		return false, err
	}
	return row != nil && row.Bool("found"), nil
}

// InsertLocalUser creates an email/password account. google_sub is NOT NULL and
// UNIQUE, so local accounts get the same synthetic "local:<uuid>" value Java used.
func (a *App) InsertLocalUser(ctx context.Context, email, name, passwordHash string) (*User, error) {
	sub, err := localSub()
	if err != nil {
		return nil, err
	}
	row, err := a.DB.QueryOne(ctx, `
		INSERT INTO users (google_sub, email, name, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, now(), now())
		RETURNING `+userCols,
		sub, email, nullIfEmpty(name), passwordHash)
	if err != nil {
		return nil, err
	}
	return scanUser(row), nil
}

// UpsertGoogleUser matches on google_sub only (never email), refreshing the cached
// Google profile while leaving display_name, custom_picture, password_hash and
// share_token untouched — exactly like the Java upsertFromGoogle.
func (a *App) UpsertGoogleUser(ctx context.Context, sub, email, name, picture string) (*User, error) {
	row, err := a.DB.QueryOne(ctx, `
		INSERT INTO users (google_sub, email, name, picture, created_at, updated_at)
		VALUES ($1, $2, $3, $4, now(), now())
		ON CONFLICT (google_sub) DO UPDATE
		   SET email = EXCLUDED.email, name = EXCLUDED.name,
		       picture = EXCLUDED.picture, updated_at = now()
		RETURNING `+userCols,
		sub, email, nullIfEmpty(name), nullIfEmpty(picture))
	if err != nil {
		return nil, err
	}
	return scanUser(row), nil
}

// updateUserCol sets one column and bumps updated_at, like Java's @PreUpdate.
func (a *App) updateUserCol(ctx context.Context, id int64, column string, value any) (*User, error) {
	row, err := a.DB.QueryOne(ctx,
		`UPDATE users SET `+column+` = $2, updated_at = now() WHERE id = $1 RETURNING `+userCols,
		id, value)
	if err != nil {
		return nil, err
	}
	return scanUser(row), nil
}

func (a *App) SetDisplayName(ctx context.Context, id int64, v string) (*User, error) {
	return a.updateUserCol(ctx, id, "display_name", nullIfEmpty(v))
}

func (a *App) SetCustomPicture(ctx context.Context, id int64, v string) (*User, error) {
	return a.updateUserCol(ctx, id, "custom_picture", nullIfEmpty(v))
}

func (a *App) SetPasswordHash(ctx context.Context, id int64, v string) (*User, error) {
	return a.updateUserCol(ctx, id, "password_hash", nullIfEmpty(v))
}

func (a *App) SetShareToken(ctx context.Context, id int64, v string) (*User, error) {
	return a.updateUserCol(ctx, id, "share_token", nullIfEmpty(v))
}

// ---- user_state ----------------------------------------------------------

// State returns the stored document and its updated_at as epoch millis. A user who
// has never saved yields (nil, 0, nil) so the API can answer {"data":null,...}.
func (a *App) State(ctx context.Context, userID int64) (json.RawMessage, int64, error) {
	row, err := a.DB.QueryOne(ctx, `
		SELECT data::text AS data,
		       (EXTRACT(epoch FROM updated_at) * 1000)::bigint::text AS updated_at
		FROM user_state WHERE user_id = $1`, userID)
	if err != nil || row == nil {
		return nil, 0, err
	}
	return row.JSON("data"), row.Int64("updated_at"), nil
}

// SaveState upserts the document and returns the new updated_at in epoch millis.
func (a *App) SaveState(ctx context.Context, userID int64, data json.RawMessage) (int64, error) {
	row, err := a.DB.QueryOne(ctx, `
		INSERT INTO user_state (user_id, data, updated_at) VALUES ($1, $2::jsonb, now())
		ON CONFLICT (user_id) DO UPDATE SET data = EXCLUDED.data, updated_at = now()
		RETURNING (EXTRACT(epoch FROM updated_at) * 1000)::bigint::text AS updated_at`,
		userID, string(data))
	if err != nil {
		return 0, err
	}
	if row == nil {
		return 0, fmt.Errorf("state was not saved")
	}
	return row.Int64("updated_at"), nil
}

// ---- password_reset_tokens ----------------------------------------------

// ReplaceResetToken invalidates any outstanding links for the user, then stores the
// SHA-256 hash of a freshly issued one (the raw token is never persisted).
func (a *App) ReplaceResetToken(ctx context.Context, userID int64, tokenHash string, ttlMinutes int) error {
	if err := a.DB.Exec(ctx, `DELETE FROM password_reset_tokens WHERE user_id = $1`, userID); err != nil {
		return err
	}
	return a.DB.Exec(ctx, `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at, used)
		VALUES ($1, $2, now() + ($3 || ' minutes')::interval, false)`,
		userID, tokenHash, ttlMinutes)
}

// UsableResetToken returns the user id behind an unused, unexpired token, or 0.
func (a *App) UsableResetToken(ctx context.Context, tokenHash string) (tokenID, userID int64, err error) {
	row, err := a.DB.QueryOne(ctx, `
		SELECT id::text AS id, user_id::text AS user_id
		FROM password_reset_tokens
		WHERE token_hash = $1 AND used = false AND expires_at > now()`, tokenHash)
	if err != nil || row == nil {
		return 0, 0, err
	}
	return row.Int64("id"), row.Int64("user_id"), nil
}

func (a *App) MarkResetTokenUsed(ctx context.Context, id int64) error {
	return a.DB.Exec(ctx, `UPDATE password_reset_tokens SET used = true WHERE id = $1`, id)
}

// ---- helpers -------------------------------------------------------------

// localSub builds the synthetic google_sub for password accounts: "local:<uuid v4>".
func localSub() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // RFC 4122 variant
	return fmt.Sprintf("local:%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// randomURLToken returns n cryptographically random bytes as unpadded base64url —
// 16 bytes for share links (22 chars), 32 for reset tokens (43 chars), matching Java.
func randomURLToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
