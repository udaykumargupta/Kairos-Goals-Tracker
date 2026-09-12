package kairos

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

// Published PBKDF2-HMAC-SHA256 vectors. If these pass, the derivation matches every
// other correct implementation (including golang.org/x/crypto/pbkdf2).
func TestPBKDF2KnownVectors(t *testing.T) {
	cases := []struct {
		password, salt string
		iter, keyLen   int
		want           string
	}{
		{"password", "salt", 1, 32, "120fb6cffcf8b32c43e7225256c4f837a86548c92ccc35480805987cb70be17b"},
		{"password", "salt", 2, 32, "ae4d0c95af6b46d32d0adff928f06dd02a303f8ef3c251dfd6e2d85a95474c43"},
		{"password", "salt", 4096, 32, "c5e478d59288c841aa530db6845c4c8d962893a001ce4e11a4963873aa98134a"},
		{"passwd", "salt", 1, 64, "55ac046e56e3089fec1691c22544b605f94185216dde0465e68b9d57c20dacbc" +
			"49ca9cccf179b645991664b39d77ef317c71b845b1e30bd509112041d3a19783"},
	}
	for _, c := range cases {
		got := hex.EncodeToString(pbkdf2Key([]byte(c.password), []byte(c.salt), c.iter, c.keyLen, sha256.New))
		if got != c.want {
			t.Errorf("pbkdf2(%q,%q,%d,%d)\n got  %s\n want %s", c.password, c.salt, c.iter, c.keyLen, got, c.want)
		}
	}
}

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword(hash, "correct horse battery staple") {
		t.Error("correct password did not verify")
	}
	if VerifyPassword(hash, "wrong password") {
		t.Error("wrong password verified")
	}
	// Two hashes of the same password must differ (random salt).
	other, _ := HashPassword("correct horse battery staple")
	if other == hash {
		t.Error("salt is not random")
	}
}

func TestLegacyBCryptDetected(t *testing.T) {
	// A real Spring BCryptPasswordEncoder output shape.
	legacy := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	if !IsLegacyBCrypt(legacy) {
		t.Error("legacy bcrypt hash not detected")
	}
	if VerifyPassword(legacy, "anything") {
		t.Error("bcrypt hash must not verify through the pbkdf2 path")
	}
	newHash, _ := HashPassword("x")
	if IsLegacyBCrypt(newHash) {
		t.Error("new hash misdetected as legacy")
	}
}

func TestPasswordHashingCost(t *testing.T) {
	start := time.Now()
	if _, err := HashPassword("benchmark"); err != nil {
		t.Fatal(err)
	}
	t.Logf("PBKDF2 at %d iterations took %v per hash", pbkdf2Iterations, time.Since(start))
}
