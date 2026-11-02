// Package crypto handles password hashing and JSON Web Token (JWT) processing.
package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"plantcare-iot/internal/model"
)

const salt = "plb_salt_2026"

// HashPassword produces a salted SHA-256 hash string for user authentication.
func HashPassword(pass string) string {
	h := sha256.New()
	h.Write([]byte(salt + pass))
	return hex.EncodeToString(h.Sum(nil))
}

// CreateJWT generates a signed HS256 JWT token with 30-day validity.
func CreateJWT(uid int, email, name string, secret string) (string, error) {
	exp := time.Now().Add(30 * 24 * time.Hour).Unix()
	claims := model.UserClaims{
		UserID: uid,
		Email:  email,
		Name:   name,
		Exp:    exp,
	}

	headerJSON := `{"alg":"HS256","typ":"JWT"}`
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	headerEnc := base64.RawURLEncoding.EncodeToString([]byte(headerJSON))
	claimsEnc := base64.RawURLEncoding.EncodeToString(claimsJSON)
	payload := headerEnc + "." + claimsEnc

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	sigEnc := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return payload + "." + sigEnc, nil
}

// VerifyJWT validates the signature and expiration time of a token string.
func VerifyJWT(tokenStr string, secret string) (*model.UserClaims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	payload := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if parts[2] != expectedSig {
		return nil, errors.New("invalid token signature")
	}

	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims model.UserClaims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, err
	}

	if time.Now().Unix() > claims.Exp {
		return nil, errors.New("token expired")
	}
	return &claims, nil
}

