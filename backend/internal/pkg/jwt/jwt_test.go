package jwt

import (
	"errors"
	"testing"
	"time"
)

func TestGenerateAndParsePair(t *testing.T) {
	manager, err := NewManager("access-secret", "refresh-secret", time.Minute, time.Hour)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	pair, err := manager.GeneratePair(123, "device-a")
	if err != nil {
		t.Fatalf("GeneratePair() error = %v", err)
	}

	accessClaims, err := manager.ParseAccessToken(pair.AccessToken.Value)
	if err != nil {
		t.Fatalf("ParseAccessToken() error = %v", err)
	}
	if accessClaims.UserID != 123 || accessClaims.DeviceID != "device-a" || accessClaims.TokenType != TokenTypeAccess {
		t.Fatalf("unexpected access claims: %#v", accessClaims)
	}

	refreshClaims, err := manager.ParseRefreshToken(pair.RefreshToken.Value)
	if err != nil {
		t.Fatalf("ParseRefreshToken() error = %v", err)
	}
	if refreshClaims.UserID != 123 || refreshClaims.DeviceID != "device-a" || refreshClaims.TokenType != TokenTypeRefresh {
		t.Fatalf("unexpected refresh claims: %#v", refreshClaims)
	}
}

func TestParseRejectsWrongTokenType(t *testing.T) {
	manager, err := NewManager("access-secret", "refresh-secret", time.Minute, time.Hour)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	pair, err := manager.GeneratePair(123, "device-a")
	if err != nil {
		t.Fatalf("GeneratePair() error = %v", err)
	}

	if _, err := manager.ParseAccessToken(pair.RefreshToken.Value); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("ParseAccessToken(refresh) error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestParseExpiredToken(t *testing.T) {
	manager, err := NewManager("access-secret", "refresh-secret", time.Nanosecond, time.Hour)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	pair, err := manager.GeneratePair(123, "device-a")
	if err != nil {
		t.Fatalf("GeneratePair() error = %v", err)
	}

	time.Sleep(2 * time.Second)
	if _, err := manager.ParseAccessToken(pair.AccessToken.Value); !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("ParseAccessToken(expired) error = %v, want %v", err, ErrExpiredToken)
	}
}
