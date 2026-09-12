package kairos

import "strings"

// User mirrors the `users` table. Nullable text columns are read via COALESCE and
// written through nullIfEmpty, so "unset" stays a real SQL NULL — important for the
// UNIQUE columns (share_token, google_sub) where '' would collide across rows.
type User struct {
	ID            int64
	GoogleSub     string
	Email         string
	Name          string
	DisplayName   string
	Picture       string
	CustomPicture string
	PasswordHash  string
	ShareToken    string
}

// EffectiveName is Java's effectiveName(): the chosen display name if set, else the Google name.
func (u *User) EffectiveName() string {
	if strings.TrimSpace(u.DisplayName) != "" {
		return u.DisplayName
	}
	return u.Name
}

// EffectivePicture is Java's effectivePicture(): the uploaded avatar if set, else the Google one.
func (u *User) EffectivePicture() string {
	if strings.TrimSpace(u.CustomPicture) != "" {
		return u.CustomPicture
	}
	return u.Picture
}

// HasPassword reports whether this account can sign in with a password (vs Google-only).
func (u *User) HasPassword() bool { return strings.TrimSpace(u.PasswordHash) != "" }

// nullIfEmpty keeps empty strings out of nullable/UNIQUE columns.
func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// ---- API payloads (field names must match the old Jackson output exactly) ----

// UserDto is the user object embedded in auth responses and returned by /api/profile.
// name/picture are pointers so an unset value serialises as null, exactly as the
// Jackson output did — "" and null are not interchangeable for API consumers.
type UserDto struct {
	Email       string  `json:"email"`
	Name        *string `json:"name"`
	Picture     *string `json:"picture"`
	HasPassword bool    `json:"hasPassword"`
}

// orNull maps "" to a JSON null.
func orNull(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func toUserDto(u *User) UserDto {
	return UserDto{
		Email:       u.Email,
		Name:        orNull(u.EffectiveName()),
		Picture:     orNull(u.EffectivePicture()),
		HasPassword: u.HasPassword(),
	}
}

// AuthResponse is { token, user } returned by the three sign-in endpoints.
type AuthResponse struct {
	Token string  `json:"token"`
	User  UserDto `json:"user"`
}

// ShareStatus is { enabled, token } — token omitted when sharing is off.
type ShareStatus struct {
	Enabled bool   `json:"enabled"`
	Token   string `json:"token,omitempty"`
}
