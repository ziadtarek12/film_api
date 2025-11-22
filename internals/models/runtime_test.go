package models

import (
	"encoding/json"
	"testing"
)

// TestRuntimeUnmarshalJSON tests the UnmarshalJSON method of Runtime
func TestRuntimeUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    Runtime
		wantErr bool
	}{
		{
			name:    "Valid runtime",
			json:    `"120 mins"`,
			want:    Runtime(120),
			wantErr: false,
		},
		{
			name:    "Missing mins",
			json:    `"120"`,
			want:    Runtime(0),
			wantErr: true,
		},
		{
			name:    "Invalid format",
			json:    `"120mins"`,
			want:    Runtime(0),
			wantErr: true,
		},
		{
			name:    "Non-numeric runtime",
			json:    `"abc mins"`,
			want:    Runtime(0),
			wantErr: true,
		},
		{
			name:    "Wrong unit",
			json:    `"120 seconds"`,
			want:    Runtime(0),
			wantErr: true,
		},
		{
			name:    "Numeric value",
			json:    `120`,
			want:    Runtime(0),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r Runtime
			err := json.Unmarshal([]byte(tt.json), &r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Runtime.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && r != tt.want {
				t.Errorf("Runtime.UnmarshalJSON() = %v, want %v", r, tt.want)
			}
		})
	}
}
