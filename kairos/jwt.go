package kairos

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"hash"
	"strconv"
	"strings"
	"time"
)

// Principal is the authenticated caller decoded from a bearer token
// (the Go equivalent of Java's AuthenticatedUser record).
type Principal struct {
	ID      int64
	Email   string
	Name    string
	Picture string
}

// jwtAlg mirrors jjwt's Keys.hmacShaKeyFor(), which chooses the HMAC variant from
// the key's bit length. The Java backend signed with whatever that returned, so we
// must derive the same algorithm from the same secret or old tokens won't verify.
func jwtAlg(secret []byte) (name string, newHash func() hash.Hash) {
	switch bits := len(secret) * 8; {
	case bits >= 512:
		return "HS512", sha512.New
	case bits >= 384:
		return "HS384", sha512.New384
	default:
		return "HS256", sha256.New
	}
}

func b64(b []byte) string        { return base64.RawURLEncoding.EncodeToString(b) }
func unb64(s string) ([]byte, error) { return base64.RawURLEncoding.DecodeString(s) }

func sign(signingInput string, secret []byte, newHash func() hash.Hash) []byte {
	m := hmac.New(newHash, secret)
	m.Write([]byte(signingInput))
	return m.Sum(nil)
}

// IssueJWT mints a token with the same claims the Java backend used:
//
//	sub = String.valueOf(user.getId()), email, name = effectiveName(),
//	picture = the raw Google picture, iat, exp
//
// jjwt drops null claims, so accounts without a Google picture get no "picture" key.
func (a *App) IssueJWT(u *User) (string, error) {
	alg, newHash := jwtAlg(a.Cfg.JWTSecret)
	now := time.Now()

	header, err := json.Marshal(map[string]string{"alg": alg})
	if err != nil {
		return "", err
	}
	claims := map[string]any{
		"sub":   strconv.FormatInt(u.ID, 10),
		"email": u.Email,
		"name":  u.EffectiveName(),
		"iat":   now.Unix(),
		"exp":   now.Add(time.Duration(a.Cfg.JWTExpiryMS) * time.Millisecond).Unix(),
	}
	if u.Picture != "" {
		claims["picture"] = u.Picture
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	input := b64(header) + "." + b64(payload)
	return input + "." + b64(sign(input, a.Cfg.JWTSecret, newHash)), nil
}

// VerifyJWT checks the signature and expiry. Like the Java verify(), every failure
// collapses to "not authenticated" rather than a distinct error.
func (a *App) VerifyJWT(token string) (Principal, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Principal{}, false
	}
	expectedAlg, newHash := jwtAlg(a.Cfg.JWTSecret)

	// Pin the algorithm to the one our key implies, so a token can't talk us into
	// a weaker scheme (alg-confusion).
	headerRaw, err := unb64(parts[0])
	if err != nil {
		return Principal{}, false
	}
	var header struct {
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(headerRaw, &header); err != nil || header.Alg != expectedAlg {
		return Principal{}, false
	}

	got, err := unb64(parts[2])
	if err != nil {
		return Principal{}, false
	}
	if !hmac.Equal(got, sign(parts[0]+"."+parts[1], a.Cfg.JWTSecret, newHash)) {
		return Principal{}, false
	}

	payloadRaw, err := unb64(parts[1])
	if err != nil {
		return Principal{}, false
	}
	var claims map[string]any
	if err := json.Unmarshal(payloadRaw, &claims); err != nil {
		return Principal{}, false
	}

	// jjwt treats a token as expired once exp <= now (no clock skew allowance).
	exp, ok := claims["exp"].(float64)
	if !ok || time.Now().Unix() >= int64(exp) {
		return Principal{}, false
	}

	sub, _ := claims["sub"].(string)
	id, err := strconv.ParseInt(sub, 10, 64)
	if err != nil {
		return Principal{}, false
	}
	str := func(k string) string { s, _ := claims[k].(string); return s }
	return Principal{ID: id, Email: str("email"), Name: str("name"), Picture: str("picture")}, true
}
