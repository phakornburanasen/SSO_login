// Package auth implements JWT (HS256) sign/verify using only the Go standard library.
// The goal is to avoid adding a third-party dependency for a small, well-defined
// spec (RFC 7519) when the project policy is to keep dependencies minimal.
//
// Token format: base64url(header) + "." + base64url(claims) + "." + base64url(HMAC-SHA256)
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// Claims is the JWT payload for the admin login session.
//   Sub  = AD username (subject)
//   Name = display name
//   Exp  = expiration time (unix seconds)
//   Iat  = issued-at time (unix seconds)
type Claims struct {
	Sub  string `json:"sub"`
	Name string `json:"name,omitempty"`
	Exp  int64  `json:"exp"`
	Iat  int64  `json:"iat"`
}

// JWT signs and verifies HS256 tokens.
type JWT struct {
	secret []byte
}

// NewJWT returns a JWT signer. The secret must be non-empty.
func NewJWT(secret string) *JWT {
	if strings.TrimSpace(secret) == "" {
		panic("auth: empty JWT secret")
	}
	return &JWT{secret: []byte(secret)}
}

// Sign issues a new JWT for the given subject (username) and display name.
// Returns the encoded token and its expiration time.
func (j *JWT) Sign(sub, name string, ttl time.Duration) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(ttl)
	claims := Claims{
		Sub:  sub,
		Name: name,
		Exp:  exp.Unix(),
		Iat:  now.Unix(),
	}

	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", time.Time{}, err
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", time.Time{}, err
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signingInput := headerB64 + "." + claimsB64
	mac := hmac.New(sha256.New, j.secret)
	mac.Write([]byte(signingInput))
	sigB64 := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + sigB64, exp, nil
}

// Verify checks the token's signature and expiration, and returns the claims
// on success. The signature is compared with hmac.Equal (constant-time).
func (j *JWT) Verify(tokenStr string) (Claims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalidToken
	}

	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, j.secret)
	mac.Write([]byte(signingInput))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return Claims{}, ErrInvalidToken
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return Claims{}, ErrInvalidToken
	}

	if time.Now().Unix() > claims.Exp {
		return Claims{}, ErrExpiredToken
	}

	return claims, nil
}
