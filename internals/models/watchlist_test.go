package models

import (
	"testing"
	"time"

	"filmapi.zeyadtarek.net/internals/validator"
)

// TestValidateWatchlistEntry tests the watchlist entry validation function
func TestValidateWatchlistEntry(t *testing.T) {
	tests := []struct {
		name      string
		entry     *Watchlist
		wantValid bool
	}{
		{
			name: "Valid watchlist entry",
			entry: &Watchlist{
				FilmID:   1,
				Priority: 5,
				Notes:    "Great movie to watch",
				Watched:  false,
			},
			wantValid: true,
		},
		{
			name: "Valid watchlist entry with rating",
			entry: &Watchlist{
				FilmID:   1,
				Priority: 8,
				Notes:    "Already watched",
				Watched:  true,
				Rating:   func() *int { r := 9; return &r }(),
			},
			wantValid: true,
		},
		{
			name: "Invalid film ID",
			entry: &Watchlist{
				FilmID:   0,
				Priority: 5,
				Notes:    "Test",
				Watched:  false,
			},
			wantValid: false,
		},
		{
			name: "Invalid priority - too low",
			entry: &Watchlist{
				FilmID:   1,
				Priority: 0,
				Notes:    "Test",
				Watched:  false,
			},
			wantValid: false,
		},
		{
			name: "Invalid priority - too high",
			entry: &Watchlist{
				FilmID:   1,
				Priority: 11,
				Notes:    "Test",
				Watched:  false,
			},
			wantValid: false,
		},
		{
			name: "Notes too long",
			entry: &Watchlist{
				FilmID:   1,
				Priority: 5,
				Notes:    string(make([]byte, 1001)), // 1001 characters
				Watched:  false,
			},
			wantValid: false,
		},
		{
			name: "Invalid rating - too low",
			entry: &Watchlist{
				FilmID:   1,
				Priority: 5,
				Notes:    "Test",
				Watched:  true,
				Rating:   func() *int { r := 0; return &r }(),
			},
			wantValid: false,
		},
		{
			name: "Invalid rating - too high",
			entry: &Watchlist{
				FilmID:   1,
				Priority: 5,
				Notes:    "Test",
				Watched:  true,
				Rating:   func() *int { r := 11; return &r }(),
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidateWatchlistEntry(v, tt.entry)

			if v.Valid() != tt.wantValid {
				t.Errorf("ValidateWatchlistEntry() = %v, want %v", v.Valid(), tt.wantValid)
				if !v.Valid() {
					t.Logf("Validation errors: %v", v.Errors)
				}
			}
		})
	}
}

// TestWatchlistWatchedAtAutoSet tests that WatchedAt is automatically set when Watched is true
func TestWatchlistWatchedAtAutoSet(t *testing.T) {
	entry := &Watchlist{
		FilmID:   1,
		Priority: 5,
		Notes:    "Test",
		Watched:  true,
		// WatchedAt is nil initially
	}

	v := validator.New()
	ValidateWatchlistEntry(v, entry)

	if !v.Valid() {
		t.Errorf("Expected validation to pass, got errors: %v", v.Errors)
	}

	if entry.WatchedAt == nil {
		t.Error("Expected WatchedAt to be set when Watched is true")
	}

	// Check that the time is recent (within last minute)
	if time.Since(*entry.WatchedAt) > time.Minute {
		t.Error("Expected WatchedAt to be set to current time")
	}
}

// TestWatchlistModel tests would require database setup, so we'll skip them for now
// but provide the structure for future implementation

func TestWatchlistInsert(t *testing.T) {
	t.Skip("Skipping test that requires database setup")
}

func TestWatchlistGet(t *testing.T) {
	t.Skip("Skipping test that requires database setup")
}

func TestWatchlistGetAll(t *testing.T) {
	t.Skip("Skipping test that requires database setup")
}

func TestWatchlistUpdate(t *testing.T) {
	t.Skip("Skipping test that requires database setup")
}

func TestWatchlistDelete(t *testing.T) {
	t.Skip("Skipping test that requires database setup")
}

func TestWatchlistCheckExists(t *testing.T) {
	t.Skip("Skipping test that requires database setup")
}
