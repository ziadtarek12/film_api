package validator

import (
	"regexp"
	"testing"
)

// TestValidatorValid tests the Valid method of Validator
func TestValidatorValid(t *testing.T) {
	tests := []struct {
		name      string
		validator *Validator
		want      bool
	}{
		{
			name:      "Empty errors",
			validator: New(),
			want:      true,
		},
		{
			name: "With errors",
			validator: func() *Validator {
				v := New()
				v.AddError("test", "error")
				return v
			}(),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.validator.Valid(); got != tt.want {
				t.Errorf("Validator.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestValidatorAddError tests the AddError method of Validator
func TestValidatorAddError(t *testing.T) {
	t.Run("Add new error", func(t *testing.T) {
		v := New()
		v.AddError("test", "error")
		if len(v.Errors) != 1 {
			t.Errorf("Validator.AddError() did not add error, got %d errors", len(v.Errors))
		}
		if v.Errors["test"] != "error" {
			t.Errorf("Validator.AddError() added wrong error, got %s", v.Errors["test"])
		}
	})

	t.Run("Don't overwrite existing error", func(t *testing.T) {
		v := New()
		v.AddError("test", "error1")
		v.AddError("test", "error2")
		if v.Errors["test"] != "error1" {
			t.Errorf("Validator.AddError() overwrote existing error, got %s", v.Errors["test"])
		}
	})
}

// TestValidatorCheck tests the Check method of Validator
func TestValidatorCheck(t *testing.T) {
	tests := []struct {
		name      string
		ok        bool
		key       string
		message   string
		wantError bool
	}{
		{
			name:      "Check passes",
			ok:        true,
			key:       "test",
			message:   "error",
			wantError: false,
		},
		{
			name:      "Check fails",
			ok:        false,
			key:       "test",
			message:   "error",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			v.Check(tt.ok, tt.key, tt.message)
			_, hasError := v.Errors[tt.key]
			if hasError != tt.wantError {
				t.Errorf("Validator.Check() error = %v, wantError %v", hasError, tt.wantError)
			}
		})
	}
}

// TestIn tests the In function
func TestIn(t *testing.T) {
	tests := []struct {
		name  string
		value string
		list  []string
		want  bool
	}{
		{
			name:  "Value in list",
			value: "test",
			list:  []string{"test", "other"},
			want:  true,
		},
		{
			name:  "Value not in list",
			value: "test",
			list:  []string{"other", "another"},
			want:  false,
		},
		{
			name:  "Empty list",
			value: "test",
			list:  []string{},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := In(tt.value, tt.list...); got != tt.want {
				t.Errorf("In() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMatches tests the Matches function
func TestMatches(t *testing.T) {
	tests := []struct {
		name  string
		value string
		rx    *regexp.Regexp
		want  bool
	}{
		{
			name:  "Value matches regex",
			value: "test@example.com",
			rx:    EmailRX,
			want:  true,
		},
		{
			name:  "Value doesn't match regex",
			value: "not-an-email",
			rx:    EmailRX,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Matches(tt.value, tt.rx); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMatchesEmail tests the MatchesEmail function
func TestMatchesEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{
			name:  "Valid email",
			email: "test@example.com",
			want:  true,
		},
		{
			name:  "Invalid email",
			email: "not-an-email",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchesEmail(tt.email); got != tt.want {
				t.Errorf("MatchesEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestUnique tests the Unique function
func TestUnique(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   bool
	}{
		{
			name:   "All unique values",
			values: []string{"test", "other", "another"},
			want:   true,
		},
		{
			name:   "Duplicate values",
			values: []string{"test", "test", "other"},
			want:   false,
		},
		{
			name:   "Empty slice",
			values: []string{},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Unique(tt.values); got != tt.want {
				t.Errorf("Unique() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMatchesURL tests the MatchesURL function
func TestMatchesURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "Valid image URL (jpg)",
			url:  "https://example.com/image.jpg",
			want: true,
		},
		{
			name: "Valid image URL (png)",
			url:  "https://example.com/image.png",
			want: true,
		},
		{
			name: "Valid image URL with query params",
			url:  "https://example.com/image.jpg?size=large",
			want: true,
		},
		{
			name: "Invalid URL (not an image)",
			url:  "https://example.com/page.html",
			want: false,
		},
		{
			name: "Invalid URL (not a URL)",
			url:  "not-a-url",
			want: false,
		},
		{
			name: "Empty URL",
			url:  "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchesURL(tt.url); got != tt.want {
				t.Errorf("MatchesURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
