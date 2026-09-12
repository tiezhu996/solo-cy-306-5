package util

import (
	"strings"
	"testing"
	"time"

	"gbevent/internal/constants"
)

func TestGenerateTokenAndParse(t *testing.T) {
	secret := "test-secret-123456"
	token, err := GenerateToken(secret, time.Hour, 42, "alice", constants.RoleUser)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}
	claims, err := ParseToken(secret, token)
	if err != nil {
		t.Fatalf("ParseToken error: %v", err)
	}
	if claims.UserID != 42 || claims.Username != "alice" || claims.Role != constants.RoleUser {
		t.Fatalf("claims mismatch: %+v", claims)
	}
}

func TestParseTokenInvalid(t *testing.T) {
	cases := []string{"", "not-a-jwt", "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.signature"}
	for _, tc := range cases {
		if _, err := ParseToken("secret", tc); err == nil {
			t.Errorf("expected error for token %q", tc)
		}
	}
}

func TestGenerateVoucherNo(t *testing.T) {
	for i := 0; i < 10; i++ {
		v := GenerateVoucherNo()
		if !ValidateVoucherFormat(v) {
			t.Fatalf("voucher %q not valid", v)
		}
		if !strings.HasPrefix(v, "GB") {
			t.Fatalf("voucher %q missing GB prefix", v)
		}
	}
}

func TestValidateVoucherFormatTable(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"GB20260816000001", true},
		{"GB123", false},
		{"AB20260816000001", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := ValidateVoucherFormat(tc.in); got != tc.want {
			t.Errorf("ValidateVoucherFormat(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestFormatters(t *testing.T) {
	if got := ActivityStatusText(constants.ActivityStatusPublished); got != "已发布" {
		t.Errorf("ActivityStatusText = %q", got)
	}
	if got := RegistrationStatusText(constants.RegistrationStatusCheckedIn); got != "已签到" {
		t.Errorf("RegistrationStatusText = %q", got)
	}
	if got := FormatCapacity(3, 10); got != "3/10" {
		t.Errorf("FormatCapacity = %q", got)
	}
	if got := ActivityTypeText(constants.ActivityTypeTraining); got != "培训" {
		t.Errorf("ActivityTypeText = %q", got)
	}
}

func TestSscanUint64(t *testing.T) {
	cases := []struct {
		in      string
		want    uint64
		wantErr bool
	}{
		{"3", 3, false},
		{"0", 0, false},
		{"abc", 0, true},
		{"", 0, true},
	}
	for _, tc := range cases {
		got, err := SscanUint64(tc.in)
		if (err != nil) != tc.wantErr {
			t.Errorf("SscanUint64(%q) err=%v", tc.in, err)
		}
		if got != tc.want {
			t.Errorf("SscanUint64(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}
