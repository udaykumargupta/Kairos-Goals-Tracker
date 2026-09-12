package kairos

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

// GoogleUser is the subset of the ID token we care about.
type GoogleUser struct {
	Sub     string
	Email   string
	Name    string
	Picture string
}

const googleCertsURL = "https://www.googleapis.com/oauth2/v3/certs"

// Google publishes its signing keys as a JWKS and rotates them. We cache the set in
// process (survives warm serverless invocations) and refetch when it goes stale or
// when a token references a key id we haven't seen.
var (
	jwksMu      sync.Mutex
	jwksKeys    map[string]*rsa.PublicKey
	jwksFetched time.Time
)

const jwksTTL = time.Hour

func googleKeys(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	jwksMu.Lock()
	defer jwksMu.Unlock()

	fresh := time.Since(jwksFetched) < jwksTTL && jwksKeys != nil
	if key, ok := jwksKeys[kid]; ok && fresh {
		return key, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleCertsURL, nil)
	if err != nil {
		return nil, err
	}
	res, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach Google's key endpoint: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google key endpoint returned %d", res.StatusCode)
	}

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(res.Body).Decode(&jwks); err != nil {
		return nil, err
	}

	keys := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		// The exponent is a big-endian byte string; left-pad to 8 bytes to read it.
		padded := make([]byte, 8)
		copy(padded[8-len(eBytes):], eBytes)
		keys[k.Kid] = &rsa.PublicKey{
			N: new(big.Int).SetBytes(nBytes),
			E: int(binary.BigEndian.Uint64(padded)),
		}
	}
	jwksKeys, jwksFetched = keys, time.Now()

	key, ok := keys[kid]
	if !ok {
		return nil, errors.New("token signed by an unknown Google key")
	}
	return key, nil
}

// VerifyGoogleIDToken validates the signature, issuer, audience and expiry of a
// Google ID token, then requires a verified email — mirroring the Java
// GoogleIdTokenVerifier plus its explicit email_verified check.
func (a *App) VerifyGoogleIDToken(ctx context.Context, idToken string) (*GoogleUser, error) {
	if strings.TrimSpace(idToken) == "" {
		return nil, errors.New("Missing Google ID token")
	}
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, errors.New("Invalid Google ID token")
	}

	headerRaw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("Invalid Google ID token")
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(headerRaw, &header); err != nil || header.Alg != "RS256" {
		return nil, errors.New("Invalid Google ID token")
	}

	key, err := googleKeys(ctx, header.Kid)
	if err != nil {
		return nil, errors.New("Could not verify Google ID token")
	}

	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errors.New("Invalid Google ID token")
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], sig); err != nil {
		return nil, errors.New("Invalid Google ID token")
	}

	payloadRaw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("Invalid Google ID token")
	}
	var claims struct {
		Iss           string `json:"iss"`
		Aud           string `json:"aud"`
		Sub           string `json:"sub"`
		Exp           int64  `json:"exp"`
		Email         string `json:"email"`
		EmailVerified any    `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := json.Unmarshal(payloadRaw, &claims); err != nil {
		return nil, errors.New("Invalid Google ID token")
	}

	if claims.Iss != "accounts.google.com" && claims.Iss != "https://accounts.google.com" {
		return nil, errors.New("Invalid Google ID token")
	}
	if a.Cfg.GoogleClientID == "" || claims.Aud != a.Cfg.GoogleClientID {
		return nil, errors.New("Invalid Google ID token")
	}
	if claims.Exp <= time.Now().Unix() {
		return nil, errors.New("Invalid Google ID token")
	}
	if claims.Sub == "" {
		return nil, errors.New("Invalid Google ID token")
	}

	// email_verified is normally a bool but has historically appeared as "true".
	verified := false
	switch v := claims.EmailVerified.(type) {
	case bool:
		verified = v
	case string:
		verified = v == "true"
	}
	if !verified {
		return nil, errors.New("Google account email is not verified")
	}

	return &GoogleUser{Sub: claims.Sub, Email: claims.Email, Name: claims.Name, Picture: claims.Picture}, nil
}
