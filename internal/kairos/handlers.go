package kairos

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"
)

// resetTokenTTLMinutes matches RESET_TOKEN_TTL_MINUTES in the Java AuthServiceImpl.
const resetTokenTTLMinutes = 60

// maxPictureChars caps the uploaded avatar data URL, as the Java controller did.
const maxPictureChars = 700_000

// ---- shared helpers ------------------------------------------------------

func normalizeEmail(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func validEmail(s string) bool {
	addr, err := mail.ParseAddress(s)
	return err == nil && addr.Address == s && strings.Contains(s, ".")
}

// validatePassword mirrors @Size(min = 8, max = 100) with the field-prefixed
// message format Spring's validation handler produced.
func validatePassword(field, pw string) string {
	switch {
	case strings.TrimSpace(pw) == "":
		return field + " must not be blank"
	case len(pw) < 8:
		return field + " must be at least 8 characters"
	case len(pw) > 100:
		return field + " must be at most 100 characters"
	}
	return ""
}

// issueFor mints a token and returns the standard { token, user } payload.
func (a *App) issueFor(w http.ResponseWriter, u *User) {
	token, err := a.IssueJWT(u)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, AuthResponse{Token: token, User: toUserDto(u)})
}

// currentUser reloads the caller from the database; the JWT carries no live state.
func (a *App) currentUser(w http.ResponseWriter, r *http.Request, p Principal) (*User, bool) {
	user, err := a.UserByID(r.Context(), p.ID)
	if err != nil {
		serverError(w, err)
		return nil, false
	}
	if user == nil {
		authFailed(w, "User not found")
		return nil, false
	}
	return user, true
}

// ---- auth ----------------------------------------------------------------

func (a *App) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDToken string `json:"idToken"`
	}
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "idToken must not be blank")
		return
	}
	if strings.TrimSpace(req.IDToken) == "" {
		authFailed(w, "Missing Google ID token")
		return
	}

	gu, err := a.VerifyGoogleIDToken(r.Context(), req.IDToken)
	if err != nil {
		authFailed(w, err.Error())
		return
	}
	user, err := a.UpsertGoogleUser(r.Context(), gu.Sub, gu.Email, gu.Name, gu.Picture)
	if err != nil {
		serverError(w, err)
		return
	}
	a.issueFor(w, user)
}

func (a *App) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "email must not be blank")
		return
	}
	email := normalizeEmail(req.Email)
	if email == "" {
		badRequest(w, "email must not be blank")
		return
	}
	if !validEmail(email) {
		badRequest(w, "email must be a well-formed email address")
		return
	}
	if msg := validatePassword("password", req.Password); msg != "" {
		badRequest(w, msg)
		return
	}

	exists, err := a.EmailExists(r.Context(), email)
	if err != nil {
		serverError(w, err)
		return
	}
	if exists {
		badRequest(w, "That email is already registered. Try logging in instead.")
		return
	}

	// Fall back to the local-part of the address when no name is supplied.
	name := strings.TrimSpace(req.Name)
	if name == "" {
		if at := strings.Index(email, "@"); at > 0 {
			name = email[:at]
		} else {
			name = email
		}
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		serverError(w, err)
		return
	}
	user, err := a.InsertLocalUser(r.Context(), email, name, hash)
	if err != nil {
		serverError(w, err)
		return
	}
	a.issueFor(w, user)
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "email must not be blank")
		return
	}
	email := normalizeEmail(req.Email)
	if email == "" {
		badRequest(w, "email must not be blank")
		return
	}
	if strings.TrimSpace(req.Password) == "" {
		badRequest(w, "password must not be blank")
		return
	}

	user, err := a.UserByEmail(r.Context(), email)
	if err != nil {
		serverError(w, err)
		return
	}
	// Same message for unknown address and wrong password, so the endpoint doesn't
	// reveal which emails are registered.
	if user == nil {
		authFailed(w, "Invalid email or password")
		return
	}
	if !user.HasPassword() {
		authFailed(w, "This email is registered with Google — use “Sign in with Google”.")
		return
	}
	// Hashes written by the previous Spring backend are BCrypt, which this build
	// cannot verify; send those users through the reset flow once.
	if IsLegacyBCrypt(user.PasswordHash) {
		authFailed(w, "Please reset your password using “Forgot password” to finish a security upgrade.")
		return
	}
	if !VerifyPassword(user.PasswordHash, req.Password) {
		authFailed(w, "Invalid email or password")
		return
	}
	a.issueFor(w, user)
}

func (a *App) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	// The reply never varies, so a malformed body is treated like an unknown address.
	const neutral = "If that email is registered, a reset link is on its way."
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"message": neutral})
		return
	}

	email := normalizeEmail(req.Email)
	if email != "" {
		user, err := a.UserByEmail(r.Context(), email)
		if err != nil {
			serverError(w, err)
			return
		}
		if user != nil {
			raw, err := randomURLToken(32) // 43 chars of base64url
			if err != nil {
				serverError(w, err)
				return
			}
			if err := a.ReplaceResetToken(r.Context(), user.ID, sha256Hex(raw), resetTokenTTLMinutes); err != nil {
				serverError(w, err)
				return
			}
			a.SendPasswordReset(user.Email, a.baseURL(r)+"/?reset="+raw)
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": neutral})
}

func (a *App) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"newPassword"`
	}
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "token must not be blank")
		return
	}
	if strings.TrimSpace(req.Token) == "" {
		authFailed(w, "This reset link is invalid or has expired.")
		return
	}
	if msg := validatePassword("newPassword", req.NewPassword); msg != "" {
		badRequest(w, msg)
		return
	}

	tokenID, userID, err := a.UsableResetToken(r.Context(), sha256Hex(req.Token))
	if err != nil {
		serverError(w, err)
		return
	}
	if tokenID == 0 {
		authFailed(w, "This reset link is invalid or has expired.")
		return
	}

	hash, err := HashPassword(req.NewPassword)
	if err != nil {
		serverError(w, err)
		return
	}
	if err := a.MarkResetTokenUsed(r.Context(), tokenID); err != nil {
		serverError(w, err)
		return
	}
	user, err := a.SetPasswordHash(r.Context(), userID, hash)
	if err != nil {
		serverError(w, err)
		return
	}
	if user == nil {
		authFailed(w, "User not found")
		return
	}
	a.issueFor(w, user) // signs the user straight in, like the Java flow
}

// ---- profile -------------------------------------------------------------

func (a *App) handleGetProfile(w http.ResponseWriter, r *http.Request, p Principal) {
	user, ok := a.currentUser(w, r, p)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, toUserDto(user))
}

func (a *App) handleUpdateProfile(w http.ResponseWriter, r *http.Request, p Principal) {
	var req struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "Invalid request body")
		return
	}
	user, err := a.SetDisplayName(r.Context(), p.ID, strings.TrimSpace(req.Name))
	if err != nil {
		serverError(w, err)
		return
	}
	if user == nil {
		authFailed(w, "User not found")
		return
	}
	writeJSON(w, http.StatusOK, toUserDto(user))
}

func (a *App) handleUpdatePicture(w http.ResponseWriter, r *http.Request, p Principal) {
	var req struct {
		Picture string `json:"picture"`
	}
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "Invalid request body")
		return
	}
	if len(req.Picture) > maxPictureChars {
		badRequest(w, "Image is too large; please choose a smaller picture")
		return
	}
	user, err := a.SetCustomPicture(r.Context(), p.ID, strings.TrimSpace(req.Picture))
	if err != nil {
		serverError(w, err)
		return
	}
	if user == nil {
		authFailed(w, "User not found")
		return
	}
	writeJSON(w, http.StatusOK, toUserDto(user))
}

// ---- state ---------------------------------------------------------------

// StateResponse is { data, updatedAt } with epoch-millis timestamps; both are null
// for a user who has never saved.
type StateResponse struct {
	Data      json.RawMessage `json:"data"`
	UpdatedAt *int64          `json:"updatedAt"`
}

func (a *App) handleGetState(w http.ResponseWriter, r *http.Request, p Principal) {
	data, updated, err := a.State(r.Context(), p.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	res := StateResponse{Data: data}
	if data != nil {
		res.UpdatedAt = &updated
	}
	writeJSON(w, http.StatusOK, res)
}

func (a *App) handleSaveState(w http.ResponseWriter, r *http.Request, p Principal) {
	// The body is the bare document, not a wrapper object.
	var data json.RawMessage
	if err := decodeJSON(r, &data); err != nil {
		badRequest(w, "State data must not be null")
		return
	}
	if len(data) == 0 || string(data) == "null" {
		badRequest(w, "State data must not be null")
		return
	}

	updated, err := a.SaveState(r.Context(), p.ID, data)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, StateResponse{Data: data, UpdatedAt: &updated})
}

// ---- share ---------------------------------------------------------------

func (a *App) handleShareStatus(w http.ResponseWriter, r *http.Request, p Principal) {
	user, ok := a.currentUser(w, r, p)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, ShareStatus{Enabled: user.ShareToken != "", Token: user.ShareToken})
}

// handleShareEnable is idempotent: an already-shared account keeps its existing link.
func (a *App) handleShareEnable(w http.ResponseWriter, r *http.Request, p Principal) {
	user, ok := a.currentUser(w, r, p)
	if !ok {
		return
	}
	if user.ShareToken == "" {
		token, err := randomURLToken(16) // 22 chars of base64url
		if err != nil {
			serverError(w, err)
			return
		}
		if user, err = a.SetShareToken(r.Context(), p.ID, token); err != nil {
			serverError(w, err)
			return
		}
	}
	writeJSON(w, http.StatusOK, ShareStatus{Enabled: true, Token: user.ShareToken})
}

func (a *App) handleShareDisable(w http.ResponseWriter, r *http.Request, p Principal) {
	if _, err := a.SetShareToken(r.Context(), p.ID, ""); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ShareStatus{Enabled: false})
}

// SharedProfile is the read-only snapshot behind a share link.
type SharedProfile struct {
	Name      *string         `json:"name"`
	Picture   *string         `json:"picture"`
	UpdatedAt *int64          `json:"updatedAt"`
	Data      json.RawMessage `json:"data"`
}

func (a *App) handlePublicShare(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	user, err := a.UserByShareToken(r.Context(), token)
	if err != nil {
		serverError(w, err)
		return
	}
	if user == nil || token == "" {
		notFound(w, "Share link not found or disabled")
		return
	}

	data, updated, err := a.State(r.Context(), user.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	res := SharedProfile{Name: orNull(user.EffectiveName()), Picture: orNull(user.EffectivePicture()), Data: data}
	if data != nil {
		res.UpdatedAt = &updated
	}
	writeJSON(w, http.StatusOK, res)
}

// ---- config / well-known -------------------------------------------------

func (a *App) handlePublicConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"googleClientId": a.Cfg.GoogleClientID})
}

// assetLinks proves to Android that this site and the TWA share an owner, so the
// installed app opens without a browser address bar.
const assetLinks = `[{
  "relation": ["delegate_permission/common.handle_all_urls"],
  "target": {
    "namespace": "android_app",
    "package_name": "app.railway.up.urkairos.twa",
    "sha256_cert_fingerprints": ["54:85:9F:CB:CB:87:32:3A:BA:72:EB:7D:DB:F7:5D:F6:62:33:97:B8:26:96:DD:D3:79:FD:F1:CA:0C:59:2F:9E"]
  }
}]
`

func (a *App) handleAssetLinks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(assetLinks))
}
