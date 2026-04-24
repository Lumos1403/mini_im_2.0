package jwt

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

type Claims struct {
	UserID    int64  `json:"user_id"`
	DeviceID  string `json:"device_id"`
	JTI       string `json:"jti"`
	TokenType string `json:"token_type"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
}

type Token struct {
	Value  string
	Claims Claims
}

type Pair struct {
	AccessToken  Token
	RefreshToken Token
	ExpiresIn    int64
}

type Manager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewManager(accessSecret string, refreshSecret string, accessTTL time.Duration, refreshTTL time.Duration) (*Manager, error) {
	if accessSecret == "" || refreshSecret == "" {
		return nil, errors.New("jwt secrets are required")
	}
	if accessTTL <= 0 || refreshTTL <= 0 {
		return nil, errors.New("jwt ttl must be positive")
	}

	return &Manager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}, nil
}

func (m *Manager) GeneratePair(userID int64, deviceID string) (*Pair, error) {
	accessToken, err := m.generate(userID, deviceID, TokenTypeAccess, m.accessTTL, m.accessSecret)
	if err != nil {
		return nil, err
	}

	refreshToken, err := m.generate(userID, deviceID, TokenTypeRefresh, m.refreshTTL, m.refreshSecret)
	if err != nil {
		return nil, err
	}

	return &Pair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(m.accessTTL.Seconds()),
	}, nil
}

func (m *Manager) ParseAccessToken(value string) (*Claims, error) {
	return parse(value, TokenTypeAccess, m.accessSecret)
}

func (m *Manager) ParseRefreshToken(value string) (*Claims, error) {
	return parse(value, TokenTypeRefresh, m.refreshSecret)
}

func (m *Manager) AccessTTL() time.Duration {
	return m.accessTTL
}

func (m *Manager) RefreshTTL() time.Duration {
	return m.refreshTTL
}

func (m *Manager) generate(userID int64, deviceID string, tokenType string, ttl time.Duration, secret []byte) (Token, error) {
	now := time.Now()
	jti, err := RandomHex(16)
	if err != nil {
		return Token{}, err
	}

	claims := Claims{
		UserID:    userID,
		DeviceID:  deviceID,
		JTI:       jti,
		TokenType: tokenType,
		ExpiresAt: now.Add(ttl).Unix(),
		IssuedAt:  now.Unix(),
	}

	value, err := sign(claims, secret)
	if err != nil {
		return Token{}, err
	}

	return Token{Value: value, Claims: claims}, nil
}

func sign(claims Claims, secret []byte) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)
	encodedPayload := base64.RawURLEncoding.EncodeToString(claimsJSON)
	unsigned := encodedHeader + "." + encodedPayload
	signature := hmacSHA256(unsigned, secret)

	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func parse(value string, expectedType string, secret []byte) (*Claims, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	unsigned := parts[0] + "." + parts[1]
	expectedSignature := hmacSHA256(unsigned, secret)
	actualSignature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrInvalidToken
	}
	if subtle.ConstantTimeCompare(expectedSignature, actualSignature) != 1 {
		return nil, ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrInvalidToken
	}
	if claims.TokenType != expectedType || claims.UserID <= 0 || claims.DeviceID == "" || claims.JTI == "" {
		return nil, ErrInvalidToken
	}
	if time.Now().Unix() >= claims.ExpiresAt {
		return nil, ErrExpiredToken
	}

	return &claims, nil
}

func hmacSHA256(value string, secret []byte) []byte {
	hash := hmac.New(sha256.New, secret)
	hash.Write([]byte(value))
	return hash.Sum(nil)
}

func RandomHex(size int) (string, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
