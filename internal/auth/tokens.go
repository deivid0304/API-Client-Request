package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

type AccessClaims struct {
	UserID int64     `json:"user_id"`
	Email  string    `json:"email"`
	Exp    time.Time `json:"exp"`
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), ttl: ttl}
}

func (m *TokenManager) NewAccessToken(userID int64, email string) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(m.ttl)
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	payload := map[string]any{
		"sub":   strconv.FormatInt(userID, 10),
		"email": email,
		"exp":   expiresAt.Unix(),
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", time.Time{}, err
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", time.Time{}, err
	}

	unsigned := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(payloadJSON)
	signature := sign(unsigned, m.secret)
	return unsigned + "." + signature, expiresAt, nil
}

func (m *TokenManager) ValidateAccessToken(token string) (AccessClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return AccessClaims{}, errors.New("formato invalido")
	}

	unsigned := parts[0] + "." + parts[1]
	expected := sign(unsigned, m.secret)
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return AccessClaims{}, errors.New("assinatura invalida")
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return AccessClaims{}, err
	}

	var payload struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Exp   int64  `json:"exp"`
	}
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return AccessClaims{}, err
	}

	expiresAt := time.Unix(payload.Exp, 0).UTC()
	if time.Now().UTC().After(expiresAt) {
		return AccessClaims{}, errors.New("token expirado")
	}

	userID, err := strconv.ParseInt(payload.Sub, 10, 64)
	if err != nil {
		return AccessClaims{}, err
	}
	return AccessClaims{UserID: userID, Email: payload.Email, Exp: expiresAt}, nil
}

func NewRefreshToken() (string, string, error) {
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", "", err
	}
	token := "rt_" + base64.RawURLEncoding.EncodeToString(random)
	return token, HashRefreshToken(token), nil
}

func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func NewFamilyID() (string, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return hex.EncodeToString(random), nil
}

func sign(unsigned string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(unsigned))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func FormatBearer(token string) string {
	return fmt.Sprintf("Bearer %s", token)
}
