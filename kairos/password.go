package kairos

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"hash"
	"strconv"
	"strings"
)

// Password hashing uses PBKDF2-HMAC-SHA256, which is built entirely from standard
// library primitives (crypto/hmac + crypto/sha256) and therefore needs no external
// module. Stored form is self-describing so the cost can be raised later without
// invalidating existing hashes:
//
//	pbkdf2_sha256$<iterations>$<base64 salt>$<base64 derived key>
const (
	pbkdf2Iterations = 600_000 // OWASP guidance for PBKDF2-HMAC-SHA256
	pbkdf2SaltLen    = 16
	pbkdf2KeyLen     = 32
	pbkdf2Prefix     = "pbkdf2_sha256"
)

// pbkdf2Key is the standard PBKDF2 construction from RFC 8018 §5.2:
// DK = T1 || T2 || …, where Ti = U1 xor U2 xor … xor Uc and
// U1 = PRF(P, S || INT_32_BE(i)), Uj = PRF(P, U(j-1)).
func pbkdf2Key(password, salt []byte, iter, keyLen int, h func() hash.Hash) []byte {
	prf := hmac.New(h, password)
	hashLen := prf.Size()
	numBlocks := (keyLen + hashLen - 1) / hashLen

	var buf [4]byte
	dk := make([]byte, 0, numBlocks*hashLen)
	u := make([]byte, hashLen)

	for block := 1; block <= numBlocks; block++ {
		prf.Reset()
		prf.Write(salt)
		buf[0], buf[1] = byte(block>>24), byte(block>>16)
		buf[2], buf[3] = byte(block>>8), byte(block)
		prf.Write(buf[:4])
		dk = prf.Sum(dk)

		t := dk[len(dk)-hashLen:]
		copy(u, t)
		for n := 2; n <= iter; n++ {
			prf.Reset()
			prf.Write(u)
			u = u[:0]
			u = prf.Sum(u)
			for i := range u {
				t[i] ^= u[i]
			}
		}
	}
	return dk[:keyLen]
}

// HashPassword derives a new stored hash for a plaintext password.
func HashPassword(plain string) (string, error) {
	salt := make([]byte, pbkdf2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	dk := pbkdf2Key([]byte(plain), salt, pbkdf2Iterations, pbkdf2KeyLen, sha256.New)
	return fmt.Sprintf("%s$%d$%s$%s", pbkdf2Prefix, pbkdf2Iterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(dk)), nil
}

// IsLegacyBCrypt reports whether a stored hash came from the old Spring backend.
// Those are Blowfish-based and cannot be verified without a third-party library,
// so such accounts are asked to reset their password once.
func IsLegacyBCrypt(stored string) bool {
	return strings.HasPrefix(stored, "$2a$") ||
		strings.HasPrefix(stored, "$2b$") ||
		strings.HasPrefix(stored, "$2y$")
}

// VerifyPassword checks a plaintext password against a stored hash.
func VerifyPassword(stored, plain string) bool {
	parts := strings.Split(stored, "$")
	if len(parts) != 4 || parts[0] != pbkdf2Prefix {
		return false
	}
	iter, err := strconv.Atoi(parts[1])
	if err != nil || iter <= 0 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}
	got := pbkdf2Key([]byte(plain), salt, iter, len(want), sha256.New)
	return subtle.ConstantTimeCompare(got, want) == 1
}

// sha256Hex is the lowercase hex SHA-256 used to store password-reset tokens,
// matching the Java HexFormat.of().formatHex(md.digest(...)).
func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%x", sum)
}
