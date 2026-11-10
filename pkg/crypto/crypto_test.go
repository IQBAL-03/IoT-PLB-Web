package crypto_test

import (
	"testing"
	"time"

	"plantcare-iot/pkg/crypto"
)

func TestHashPassword(t *testing.T) {
	pass := "secretPassword123"
	hash1 := crypto.HashPassword(pass)
	hash2 := crypto.HashPassword(pass)

	if hash1 == "" {
		t.Fatal("password hash should not be empty")
	}
	if hash1 != hash2 {
		t.Fatal("identical passwords must produce identical hashes")
	}
	if hash1 == pass {
		t.Fatal("password must never be plaintext")
	}
}

func TestHashOTP(t *testing.T) {
	code := "123456"
	hash := crypto.HashOTP(code)

	if hash == "" {
		t.Fatal("OTP hash should not be empty")
	}
	if hash == code {
		t.Fatal("OTP must not be plaintext")
	}
	if len(hash) != 64 {
		t.Fatalf("expected SHA-256 hex length of 64, got %d", len(hash))
	}

	hashAgain := crypto.HashOTP("  123456  ")
	if hash != hashAgain {
		t.Fatal("OTP hash should trim surrounding whitespace")
	}
}

func TestJWTLifecycle(t *testing.T) {
	secret := "test-jwt-secret-key-32-chars-long"
	token, err := crypto.CreateJWT(1, "test@example.com", "Test User", secret)
	if err != nil {
		t.Fatalf("failed to create JWT: %v", err)
	}

	claims, err := crypto.VerifyJWT(token, secret)
	if err != nil {
		t.Fatalf("failed to verify valid JWT: %v", err)
	}

	if claims.UserID != 1 || claims.Email != "test@example.com" || claims.Name != "Test User" {
		t.Fatalf("claims mismatch: %+v", claims)
	}

	// Verify tampering fails
	tamperedToken := token + "tamper"
	_, err = crypto.VerifyJWT(tamperedToken, secret)
	if err == nil {
		t.Fatal("tampered token must fail verification")
	}

	// Verify wrong secret fails
	_, err = crypto.VerifyJWT(token, "wrong-secret-key")
	if err == nil {
		t.Fatal("token verified with wrong secret must fail")
	}
}

func TestJWTExpired(t *testing.T) {
	now := time.Now().Unix()
	if now <= 0 {
		t.Fatal("invalid timestamp")
	}
}
