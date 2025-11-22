package models

import (
	"testing"

	"filmapi.zeyadtarek.net/internals/validator"
)

// TestValidateUser tests the user validation function
func TestValidateUser(t *testing.T) {
	tests := []struct {
		name      string
		user      *User
		wantValid bool
	}{
		{
			name: "Valid user",
			user: func() *User {
				user := &User{
					Name:      "Test User",
					Email:     "test@example.com",
					Activated: true,
				}
				_ = user.Password.Set("password123")
				return user
			}(),
			wantValid: true,
		},
		{
			name: "Missing name",
			user: func() *User {
				user := &User{
					Name:      "",
					Email:     "test@example.com",
					Activated: true,
				}
				_ = user.Password.Set("password123")
				return user
			}(),
			wantValid: false,
		},
		{
			name: "Invalid email",
			user: func() *User {
				user := &User{
					Name:      "Test User",
					Email:     "not-an-email",
					Activated: true,
				}
				_ = user.Password.Set("password123")
				return user
			}(),
			wantValid: false,
		},
		{
			name: "Password too short",
			user: func() *User {
				user := &User{
					Name:      "Test User",
					Email:     "test@example.com",
					Activated: true,
				}
				_ = user.Password.Set("short")
				return user
			}(),
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidateUser(v, tt.user)
			if v.Valid() != tt.wantValid {
				t.Errorf("ValidateUser() got valid = %v, want %v, errors: %v", v.Valid(), tt.wantValid, v.Errors)
			}
		})
	}
}

// TestValidateEmail tests the email validation function
func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		wantValid bool
	}{
		{
			name:      "Valid email",
			email:     "test@example.com",
			wantValid: true,
		},
		{
			name:      "Empty email",
			email:     "",
			wantValid: false,
		},
		{
			name:      "Invalid email (no @)",
			email:     "testexample.com",
			wantValid: false,
		},
		{
			name:      "Invalid email (no domain)",
			email:     "test@",
			wantValid: false,
		},
		{
			name:      "Invalid email (no username)",
			email:     "@example.com",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidateEmail(v, tt.email)
			if v.Valid() != tt.wantValid {
				t.Errorf("ValidateEmail() got valid = %v, want %v, errors: %v", v.Valid(), tt.wantValid, v.Errors)
			}
		})
	}
}

// TestValidatePasswordPlaintext tests the password validation function
func TestValidatePasswordPlaintext(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantValid bool
	}{
		{
			name:      "Valid password",
			password:  "password123",
			wantValid: true,
		},
		{
			name:      "Empty password",
			password:  "",
			wantValid: false,
		},
		{
			name:      "Password too short",
			password:  "short",
			wantValid: false,
		},
		{
			name:      "Password too long",
			password:  string(make([]byte, 73)),
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidatePasswordPlaintext(v, tt.password)
			if v.Valid() != tt.wantValid {
				t.Errorf("ValidatePasswordPlaintext() got valid = %v, want %v, errors: %v", v.Valid(), tt.wantValid, v.Errors)
			}
		})
	}
}

// TestPasswordSet tests the password Set method
func TestPasswordSet(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
		wantErr   bool
	}{
		{
			name:      "Valid password",
			plaintext: "password123",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &password{}
			err := p.Set(tt.plaintext)
			if (err != nil) != tt.wantErr {
				t.Errorf("password.Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if p.plaintext == nil || *p.plaintext != tt.plaintext {
				t.Errorf("password.Set() plaintext = %v, want %v", *p.plaintext, tt.plaintext)
			}
			if p.hash == nil {
				t.Errorf("password.Set() hash is nil")
			}
		})
	}
}

// TestPasswordMatches tests the password Matches method
func TestPasswordMatches(t *testing.T) {
	p := &password{}
	plaintext := "password123"
	err := p.Set(plaintext)
	if err != nil {
		t.Fatalf("Failed to set password: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
		want      bool
		wantErr   bool
	}{
		{
			name:      "Matching password",
			plaintext: plaintext,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "Non-matching password",
			plaintext: "wrongpassword",
			want:      false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.Matches(tt.plaintext)
			if (err != nil) != tt.wantErr {
				t.Errorf("password.Matches() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("password.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestUserIsAnonymous tests the IsAnonymous method
func TestUserIsAnonymous(t *testing.T) {
	tests := []struct {
		name string
		user *User
		want bool
	}{
		{
			name: "Anonymous user",
			user: AnonymousUser,
			want: true,
		},
		{
			name: "Regular user",
			user: &User{
				ID:    1,
				Name:  "Test User",
				Email: "test@example.com",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.user.IsAnonyomous(); got != tt.want {
				t.Errorf("User.IsAnonymous() = %v, want %v", got, tt.want)
			}
		})
	}
}
