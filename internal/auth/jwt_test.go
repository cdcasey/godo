package auth

import (
	"testing"
	"time"
)

const testSecret = "test-secret-key"

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken("user-123", "test@example.com", "user", testSecret, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Validate token
	claims, err := ValidateToken(token, testSecret)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("expected UserID 'user-123', got '%s'", claims.UserID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("expected Email 'test@example.com', got '%s'", claims.Email)
	}
	if claims.Role != "user" {
		t.Errorf("expected Role 'user', got '%s'", claims.Role)
	}
}

func TestValidateToken(t *testing.T) {
	validToken, _ := GenerateToken("user-123", "test@example.com", "user", testSecret, time.Hour)
	expiredToken, _ := GenerateToken("user-123", "test@example.com", "user", testSecret, -time.Hour)

	tests := []struct {
		name      string
		token     string
		secret    string
		wantErr   error
		wantClaim string // UserID to check on success
	}{
		{
			name:      "valid token",
			token:     validToken,
			secret:    testSecret,
			wantErr:   nil,
			wantClaim: "user-123",
		},
		{
			name:    "wrong secret",
			token:   validToken,
			secret:  "wrong-secret",
			wantErr: ErrInvalidToken,
		},
		{
			name:    "expired token",
			token:   expiredToken,
			secret:  testSecret,
			wantErr: ErrExpiredToken,
		},
		{
			name:    "malformed token",
			token:   "not.valid.token",
			secret:  testSecret,
			wantErr: ErrInvalidToken,
		},
		{
			name:    "empty token",
			token:   "",
			secret:  testSecret,
			wantErr: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateToken(tt.token, tt.secret)
			if err != tt.wantErr {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && claims.UserID != tt.wantClaim {
				t.Errorf("expected UserID %s, got %s", tt.wantClaim, claims.UserID)
			}
		})
	}
}
