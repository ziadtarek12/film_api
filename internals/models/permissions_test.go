package models

import (
	"testing"
)

// TestPermissionsInclude tests the Include method of Permissions
func TestPermissionsInclude(t *testing.T) {
	tests := []struct {
		name        string
		permissions Permissions
		code        string
		want        bool
	}{
		{
			name:        "Permission included",
			permissions: Permissions{"films:read", "films:write"},
			code:        "films:read",
			want:        true,
		},
		{
			name:        "Permission not included",
			permissions: Permissions{"films:read"},
			code:        "films:write",
			want:        false,
		},
		{
			name:        "Empty permissions",
			permissions: Permissions{},
			code:        "films:read",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.permissions.Include(tt.code); got != tt.want {
				t.Errorf("Permissions.Include() = %v, want %v", got, tt.want)
			}
		})
	}
}
