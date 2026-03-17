package grid

import (
	"testing"
)

func TestParseClasses(t *testing.T) {
	tests := []struct {
		name    string
		classes []string
		want    Position
		found   bool
	}{
		{
			name:    "valid grid class",
			classes: []string{"row-2-col-3"},
			want:    Position{Row: 2, Col: 3},
			found:   true,
		},
		{
			name:    "grid class among others",
			classes: []string{"highlight", "row-1-col-1", "primary"},
			want:    Position{Row: 1, Col: 1},
			found:   true,
		},
		{
			name:    "no grid class",
			classes: []string{"highlight", "primary"},
			want:    Position{},
			found:   false,
		},
		{
			name:    "empty classes",
			classes: []string{},
			want:    Position{},
			found:   false,
		},
		{
			name:    "nil classes",
			classes: nil,
			want:    Position{},
			found:   false,
		},
		{
			name:    "invalid format",
			classes: []string{"row-abc-col-def"},
			want:    Position{},
			found:   false,
		},
		{
			name:    "zero row rejected",
			classes: []string{"row-0-col-1"},
			want:    Position{},
			found:   false,
		},
		{
			name:    "large grid position",
			classes: []string{"row-50-col-50"},
			want:    Position{Row: 50, Col: 50},
			found:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found := ParseClasses(tt.classes)
			if found != tt.found {
				t.Errorf("ParseClasses() found = %v, want %v", found, tt.found)
			}
			if got != tt.want {
				t.Errorf("ParseClasses() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPixelPosition(t *testing.T) {
	opts := DefaultOptions() // CellWidth=200, CellHeight=120, Gap=40, Padding=60

	tests := []struct {
		name  string
		pos   Position
		wantX float64
		wantY float64
	}{
		{
			name:  "top left cell",
			pos:   Position{Row: 1, Col: 1},
			wantX: 60,
			wantY: 60,
		},
		{
			name:  "second column",
			pos:   Position{Row: 1, Col: 2},
			wantX: 300, // 60 + (200+40)
			wantY: 60,
		},
		{
			name:  "second row",
			pos:   Position{Row: 2, Col: 1},
			wantX: 60,
			wantY: 220, // 60 + (120+40)
		},
		{
			name:  "row 3 col 3",
			pos:   Position{Row: 3, Col: 3},
			wantX: 540, // 60 + 2*(200+40)
			wantY: 380, // 60 + 2*(120+40)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y := opts.PixelPosition(tt.pos)
			if x != tt.wantX {
				t.Errorf("PixelPosition() x = %v, want %v", x, tt.wantX)
			}
			if y != tt.wantY {
				t.Errorf("PixelPosition() y = %v, want %v", y, tt.wantY)
			}
		})
	}
}

func TestPixelPositionCustomOpts(t *testing.T) {
	opts := Options{CellWidth: 100, CellHeight: 80, Gap: 20, Padding: 10}
	x, y := opts.PixelPosition(Position{Row: 2, Col: 3})
	wantX := float64(10 + 2*(100+20)) // 250
	wantY := float64(10 + 1*(80+20))  // 110
	if x != wantX {
		t.Errorf("x = %v, want %v", x, wantX)
	}
	if y != wantY {
		t.Errorf("y = %v, want %v", y, wantY)
	}
}

func TestValidatePositions(t *testing.T) {
	t.Run("no conflicts", func(t *testing.T) {
		positions := map[string]Position{
			"a": {Row: 1, Col: 1},
			"b": {Row: 1, Col: 2},
			"c": {Row: 2, Col: 1},
		}
		if err := ValidatePositions(positions); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("conflict detected", func(t *testing.T) {
		positions := map[string]Position{
			"a": {Row: 1, Col: 1},
			"b": {Row: 1, Col: 1},
		}
		if err := ValidatePositions(positions); err == nil {
			t.Error("expected conflict error, got nil")
		}
	})

	t.Run("empty map", func(t *testing.T) {
		if err := ValidatePositions(map[string]Position{}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
