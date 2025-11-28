package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/auth-service/internal/config"
	"github.com/xiiisorate/granula_api/auth-service/internal/repository"
)

func newTestJWTService() *JWTService {
	return NewJWTService(config.JWTConfig{
		Secret:        "test-secret-key-for-testing-12345",
		AccessExpire:  15 * time.Minute,
		RefreshExpire: 7 * 24 * time.Hour,
	})
}

func TestJWTService_GenerateAccessToken(t *testing.T) {
	svc := newTestJWTService()

	user := &repository.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  "user",
	}

	token, expiresAt, err := svc.GenerateAccessToken(user)

	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	if token == "" {
		t.Error("Token should not be empty")
	}

	if expiresAt.Before(time.Now()) {
		t.Error("ExpiresAt should be in the future")
	}

	// Token should expire in ~15 minutes
	expectedExpire := time.Now().Add(15 * time.Minute)
	if expiresAt.After(expectedExpire.Add(time.Second)) || expiresAt.Before(expectedExpire.Add(-time.Second)) {
		t.Errorf("ExpiresAt should be ~15 minutes from now, got %v", expiresAt)
	}
}

func TestJWTService_GenerateRefreshToken(t *testing.T) {
	svc := newTestJWTService()

	user := &repository.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  "user",
	}

	token, expiresAt, err := svc.GenerateRefreshToken(user)

	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	if token == "" {
		t.Error("Token should not be empty")
	}

	// Token should expire in ~7 days
	expectedExpire := time.Now().Add(7 * 24 * time.Hour)
	if expiresAt.After(expectedExpire.Add(time.Second)) || expiresAt.Before(expectedExpire.Add(-time.Second)) {
		t.Errorf("ExpiresAt should be ~7 days from now, got %v", expiresAt)
	}
}

func TestJWTService_ValidateToken_Valid(t *testing.T) {
	svc := newTestJWTService()

	user := &repository.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  "admin",
	}

	token, _, err := svc.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("UserID mismatch: expected %v, got %v", user.ID, claims.UserID)
	}

	if claims.Email != user.Email {
		t.Errorf("Email mismatch: expected %v, got %v", user.Email, claims.Email)
	}

	if claims.Role != user.Role {
		t.Errorf("Role mismatch: expected %v, got %v", user.Role, claims.Role)
	}
}

func TestJWTService_ValidateToken_Invalid(t *testing.T) {
	svc := newTestJWTService()

	_, err := svc.ValidateToken("invalid-token")
	if err == nil {
		t.Error("ValidateToken should fail for invalid token")
	}
}

func TestJWTService_ValidateToken_WrongSecret(t *testing.T) {
	svc1 := NewJWTService(config.JWTConfig{
		Secret:        "secret-one",
		AccessExpire:  15 * time.Minute,
		RefreshExpire: 7 * 24 * time.Hour,
	})

	svc2 := NewJWTService(config.JWTConfig{
		Secret:        "secret-two",
		AccessExpire:  15 * time.Minute,
		RefreshExpire: 7 * 24 * time.Hour,
	})

	user := &repository.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  "user",
	}

	token, _, _ := svc1.GenerateAccessToken(user)

	_, err := svc2.ValidateToken(token)
	if err == nil {
		t.Error("ValidateToken should fail for token signed with different secret")
	}
}

func TestJWTService_GetExpireDurations(t *testing.T) {
	accessExpire := 30 * time.Minute
	refreshExpire := 14 * 24 * time.Hour

	svc := NewJWTService(config.JWTConfig{
		Secret:        "test-secret",
		AccessExpire:  accessExpire,
		RefreshExpire: refreshExpire,
	})

	if svc.GetAccessExpire() != accessExpire {
		t.Errorf("AccessExpire mismatch: expected %v, got %v", accessExpire, svc.GetAccessExpire())
	}

	if svc.GetRefreshExpire() != refreshExpire {
		t.Errorf("RefreshExpire mismatch: expected %v, got %v", refreshExpire, svc.GetRefreshExpire())
	}
}

