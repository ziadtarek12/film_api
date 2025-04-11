package models

import (
	"testing"

	"filmapi.zeyadtarek.net/internals/validator"
)

// TestValidateFilters tests the filter validation function
func TestValidateFilters(t *testing.T) {
	tests := []struct {
		name      string
		filters   Filters
		wantValid bool
	}{
		{
			name: "Valid filters",
			filters: Filters{
				Page:         1,
				PageSize:     20,
				SortValues:   []string{"id", "-title"},
				SortSafelist: []string{"id", "title", "-id", "-title"},
			},
			wantValid: true,
		},
		{
			name: "Invalid page (zero)",
			filters: Filters{
				Page:         0,
				PageSize:     20,
				SortValues:   []string{"id"},
				SortSafelist: []string{"id", "title", "-id", "-title"},
			},
			wantValid: false,
		},
		{
			name: "Invalid page (too large)",
			filters: Filters{
				Page:         20_000_000,
				PageSize:     20,
				SortValues:   []string{"id"},
				SortSafelist: []string{"id", "title", "-id", "-title"},
			},
			wantValid: false,
		},
		{
			name: "Invalid page size (zero)",
			filters: Filters{
				Page:         1,
				PageSize:     0,
				SortValues:   []string{"id"},
				SortSafelist: []string{"id", "title", "-id", "-title"},
			},
			wantValid: false,
		},
		{
			name: "Invalid page size (too large)",
			filters: Filters{
				Page:         1,
				PageSize:     200,
				SortValues:   []string{"id"},
				SortSafelist: []string{"id", "title", "-id", "-title"},
			},
			wantValid: false,
		},
		{
			name: "Invalid sort value",
			filters: Filters{
				Page:         1,
				PageSize:     20,
				SortValues:   []string{"invalid"},
				SortSafelist: []string{"id", "title", "-id", "-title"},
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidateFilters(v, tt.filters)
			if v.Valid() != tt.wantValid {
				t.Errorf("ValidateFilters() got valid = %v, want %v, errors: %v", v.Valid(), tt.wantValid, v.Errors)
			}
		})
	}
}

// TestFiltersMethods tests the methods of Filters
func TestFiltersMethods(t *testing.T) {
	// Test sortColumn method
	t.Run("sortColumn with no sort values", func(t *testing.T) {
		f := Filters{
			SortValues:   []string{},
			SortSafelist: []string{"id", "title", "-id", "-title"},
		}
		got := f.sortColumn()
		want := ""
		if got != want {
			t.Errorf("Filters.sortColumn() = %v, want %v", got, want)
		}
	})

	t.Run("sortColumn with one sort value", func(t *testing.T) {
		f := Filters{
			SortValues:   []string{"id"},
			SortSafelist: []string{"id", "title", "-id", "-title"},
		}
		got := f.sortColumn()
		want := "id ASC,"
		if got != want {
			t.Errorf("Filters.sortColumn() = %v, want %v", got, want)
		}
	})

	t.Run("sortColumn with multiple sort values", func(t *testing.T) {
		f := Filters{
			SortValues:   []string{"id", "title"},
			SortSafelist: []string{"id", "title", "-id", "-title"},
		}
		got := f.sortColumn()
		if got != "id ASC,title ASC," {
			t.Errorf("Filters.sortColumn() = %v, want %v", got, "id ASC,title ASC,")
		}
	})

	// Test limit method
	t.Run("limit", func(t *testing.T) {
		f := Filters{
			PageSize: 20,
		}
		got := f.limit()
		want := 20
		if got != want {
			t.Errorf("Filters.limit() = %v, want %v", got, want)
		}
	})

	// Test offset method
	t.Run("offset with page 1", func(t *testing.T) {
		f := Filters{
			Page:     1,
			PageSize: 20,
		}
		got := f.offset()
		want := 0
		if got != want {
			t.Errorf("Filters.offset() = %v, want %v", got, want)
		}
	})

	t.Run("offset with page > 1", func(t *testing.T) {
		f := Filters{
			Page:     3,
			PageSize: 20,
		}
		got := f.offset()
		want := 40
		if got != want {
			t.Errorf("Filters.offset() = %v, want %v", got, want)
		}
	})
}
