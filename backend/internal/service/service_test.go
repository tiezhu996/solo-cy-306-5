package service

import (
	"testing"

	"gbevent/internal/constants"
)

func TestFmtUint(t *testing.T) {
	cases := []struct {
		in   uint64
		want string
	}{
		{0, "0"}, {1, "1"}, {42, "42"}, {123456789, "123456789"},
	}
	for _, tc := range cases {
		if got := fmtUint(tc.in); got != tc.want {
			t.Errorf("fmtUint(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestRound2(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{0, 0}, {50.0, 50}, {33.3333, 33.33}, {66.6666, 66.67}, {12.345, 12.35},
	}
	for _, tc := range cases {
		if got := round2(tc.in); got != tc.want {
			t.Errorf("round2(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestIsOrganizer(t *testing.T) {
	cases := []struct {
		name         string
		operatorID   uint64
		operatorRole string
		organizerID  uint64
		want         bool
	}{
		{"admin can manage any", 1, constants.RoleAdmin, 99, true},
		{"organizer owner", 5, constants.RoleOrganizer, 5, true},
		{"organizer not owner", 5, constants.RoleOrganizer, 6, false},
		{"normal user not owner", 3, constants.RoleUser, 5, false},
	}
	for _, tc := range cases {
		if got := IsOrganizer(tc.operatorID, tc.operatorRole, tc.organizerID); got != tc.want {
			t.Errorf("%s: IsOrganizer = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestStatusValidators(t *testing.T) {
	if !constants.IsValidActivityStatus(constants.ActivityStatusPublished) {
		t.Error("published should be valid")
	}
	if constants.IsValidActivityStatus("unknown") {
		t.Error("unknown status should be invalid")
	}
	if !constants.IsValidActivityType(constants.ActivityTypeParty) {
		t.Error("party should be valid type")
	}
	if constants.IsValidReviewStatus("bogus") {
		t.Error("bogus review status should be invalid")
	}
}
