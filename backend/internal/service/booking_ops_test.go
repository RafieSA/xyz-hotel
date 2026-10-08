package service

import "testing"

func TestIsValidRoomStatusTransition(t *testing.T) {
	tests := []struct {
		from, to string
		want     bool
	}{
		{"available", "occupied", true},
		{"available", "dirty", true},
		{"available", "maintenance", true},
		{"available", "available", false},
		{"occupied", "dirty", true},
		{"occupied", "available", false},
		{"occupied", "maintenance", false},
		{"dirty", "available", true},
		{"dirty", "maintenance", true},
		{"dirty", "occupied", false},
		{"maintenance", "available", true},
		{"maintenance", "dirty", false},
		{"maintenance", "occupied", false},
	}
	for _, tc := range tests {
		got := IsValidRoomStatusTransition(tc.from, tc.to)
		if got != tc.want {
			t.Errorf("transition %s->%s = %v, want %v", tc.from, tc.to, got, tc.want)
		}
	}
}

func TestValidRoomStatuses(t *testing.T) {
	for _, s := range []string{"available", "occupied", "dirty", "maintenance"} {
		if !ValidRoomStatuses[s] {
			t.Errorf("expected %s valid", s)
		}
	}
	if ValidRoomStatuses["invalid"] {
		t.Error("invalid should not be valid")
	}
}
