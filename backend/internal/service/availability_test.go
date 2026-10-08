package service

import (
	"testing"
	"time"
)

func parseDateMust(t *testing.T, s string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("parse %s: %v", s, err)
	}
	return parsed
}

func TestIsOverlapping(t *testing.T) {
	tests := []struct {
		name     string
		aStart   string
		aEnd     string
		bStart   string
		bEnd     string
		overlap  bool
	}{
		// No overlap cases: touching edges are NOT overlapping (check_out == check_in is free)
		{"no overlap - before", "2026-10-10", "2026-10-12", "2026-10-12", "2026-10-14", false},
		{"no overlap - after", "2026-10-14", "2026-10-16", "2026-10-10", "2026-10-12", false},
		{"no overlap - disjoint far", "2026-10-01", "2026-10-05", "2026-10-10", "2026-10-15", false},
		{"no overlap - adjacent same day check_out==check_in", "2026-10-10", "2026-10-11", "2026-10-11", "2026-10-12", false},
		// Overlap cases
		{"overlap - exact same", "2026-10-10", "2026-10-12", "2026-10-10", "2026-10-12", true},
		{"overlap - partial start", "2026-10-10", "2026-10-14", "2026-10-12", "2026-10-16", true},
		{"overlap - partial end", "2026-10-12", "2026-10-16", "2026-10-10", "2026-10-14", true},
		{"overlap - inside", "2026-10-10", "2026-10-20", "2026-10-12", "2026-10-15", true},
		{"overlap - contains", "2026-10-12", "2026-10-15", "2026-10-10", "2026-10-20", true},
		{"overlap - one night inside", "2026-10-10", "2026-10-15", "2026-10-11", "2026-10-12", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			aStart := parseDateMust(t, tc.aStart)
			aEnd := parseDateMust(t, tc.aEnd)
			bStart := parseDateMust(t, tc.bStart)
			bEnd := parseDateMust(t, tc.bEnd)
			got := IsOverlapping(aStart, aEnd, bStart, bEnd)
			if got != tc.overlap {
				t.Fatalf("IsOverlapping([%s,%s), [%s,%s)) = %v, want %v", tc.aStart, tc.aEnd, tc.bStart, tc.bEnd, got, tc.overlap)
			}
			// Symmetry check
			got2 := IsOverlapping(bStart, bEnd, aStart, aEnd)
			if got2 != tc.overlap {
				t.Fatalf("symmetry failed: IsOverlapping swap = %v, want %v", got2, tc.overlap)
			}
		})
	}
}

func TestCalculateAvailable(t *testing.T) {
	tests := []struct {
		name       string
		total      int
		overlapped int
		want       int
	}{
		{"all free", 8, 0, 8},
		{"partial", 5, 2, 3},
		{"full", 3, 3, 0},
		{"overbook clamp", 2, 5, 0},
		{"single unit free", 1, 0, 1},
		{"single unit occupied", 1, 1, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CalculateAvailable(tc.total, tc.overlapped)
			if got != tc.want {
				t.Fatalf("CalculateAvailable(%d,%d)=%d want %d", tc.total, tc.overlapped, got, tc.want)
			}
		})
	}
}

func TestNightsBetween(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		out     string
		want    int
		wantErr bool
	}{
		{"2 nights", "2026-10-10", "2026-10-12", 2, false},
		{"1 night", "2026-10-10", "2026-10-11", 1, false},
		{"7 nights", "2026-10-01", "2026-10-08", 7, false},
		{"invalid order", "2026-10-12", "2026-10-10", 0, true},
		{"same day", "2026-10-10", "2026-10-10", 0, true},
		{"bad format", "10-10-2026", "2026-10-12", 0, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NightsBetween(tc.in, tc.out)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("NightsBetween(%s,%s)=%d want %d", tc.in, tc.out, got, tc.want)
			}
		})
	}
}
