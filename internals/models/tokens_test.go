package models

import (
	"testing"
	"time"

	"filmapi.zeyadtarek.net/internals/validator"
)

// TestGenerateToken tests the token generation function
func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name    string
		userID  int64
		ttl     time.Duration
		scope   string
		wantErr bool
	}{
		{
			name:    "Valid token",
			userID:  1,
			ttl:     24 * time.Hour,
			scope:   ScopeActivation,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := generateToken(tt.userID, tt.ttl, tt.scope)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if token == nil {
				t.Fatal("generateToken() returned nil token")
			}
			if token.UserId != tt.userID {
				t.Errorf("generateToken() token.UserID = %v, want %v", token.UserId, tt.userID)
			}
			if token.Scope != tt.scope {
				t.Errorf("generateToken() token.Scope = %v, want %v", token.Scope, tt.scope)
			}
			if token.Plaintext == "" {
				t.Error("generateToken() token.Plaintext is empty")
			}
			if len(token.Hash) == 0 {
				t.Error("generateToken() token.Hash is empty")
			}
			
			// Check that expiry is in the future and approximately matches the TTL
			now := time.Now()
			if token.Expiry.Before(now) {
				t.Error("generateToken() token.Expiry is in the past")
			}
			
			expectedExpiry := now.Add(tt.ttl)
			diff := token.Expiry.Sub(expectedExpiry)
			if diff < -time.Second || diff > time.Second {
				t.Errorf("generateToken() token.Expiry differs from expected by %v", diff)
			}
		})
	}
}

// TestValidateTokenPlaintext tests the token validation function
func TestValidateTokenPlaintext(t *testing.T) {
	tests := []struct {
		name          string
		tokenPlaintext string
		wantValid     bool
	}{
		{
			name:          "Valid token",
			tokenPlaintext: "ABCDEFGHIJKLMNOPQRSTUVWXYZ", // 26 characters
			wantValid:     true,
		},
		{
			name:          "Empty token",
			tokenPlaintext: "",
			wantValid:     false,
		},
		{
			name:          "Token too short",
			tokenPlaintext: "ABCDEFGHIJ",
			wantValid:     false,
		},
		{
			name:          "Token too long",
			tokenPlaintext: "ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJ",
			wantValid:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidateTokenPlaintext(v, tt.tokenPlaintext)
			if v.Valid() != tt.wantValid {
				t.Errorf("ValidateTokenPlaintext() got valid = %v, want %v, errors: %v", v.Valid(), tt.wantValid, v.Errors)
			}
		})
	}
}
