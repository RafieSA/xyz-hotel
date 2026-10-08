package service

import (
	"strings"
	"testing"
	"time"

	"xyz-hotel/backend/internal/model"
)

func TestValidateVoucher_Expiry(t *testing.T) {
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)

	tests := []struct {
		name    string
		voucher model.Voucher
		nights  int
		wantErr string
	}{
		{"not expired nil", model.Voucher{Code: "FREE", Discount: 10, MinNights: 0, ExpiresAt: nil, UsedCount: 0}, 1, ""},
		{"future expiry", model.Voucher{Code: "FUT", Discount: 10, ExpiresAt: &future}, 1, ""},
		{"past expiry", model.Voucher{Code: "OLD", Discount: 10, ExpiresAt: &past}, 1, "expired"},
		{"exactly now+1ns not expired", model.Voucher{Code: "NOW", Discount: 10, ExpiresAt: func() *time.Time { x := now.Add(time.Hour); return &x }()}, 1, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateVoucher(&tc.voucher, tc.nights, now)
			if tc.wantErr == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr != "" && err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if tc.wantErr != "" && err != nil && !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

func TestValidateVoucher_Quota(t *testing.T) {
	now := time.Now()
	quota1 := 1
	quota5 := 5
	quota0 := 0

	tests := []struct {
		name    string
		voucher model.Voucher
		nights  int
		wantErr string
	}{
		{"quota nil unlimited", model.Voucher{Code: "UNLIM", Discount: 10, Quota: nil, UsedCount: 100}, 1, ""},
		{"quota not exceeded", model.Voucher{Code: "OK", Discount: 10, Quota: &quota5, UsedCount: 2}, 1, ""},
		{"quota exactly at limit", model.Voucher{Code: "FULL", Discount: 10, Quota: &quota1, UsedCount: 1}, 1, "quota exceeded"},
		{"quota exceeded beyond", model.Voucher{Code: "OVER", Discount: 10, Quota: &quota5, UsedCount: 6}, 1, "quota exceeded"},
		{"quota zero immediately exceeded", model.Voucher{Code: "ZERO", Discount: 10, Quota: &quota0, UsedCount: 0}, 1, "quota exceeded"},
		{"quota one used zero ok", model.Voucher{Code: "ONE", Discount: 10, Quota: &quota1, UsedCount: 0}, 1, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateVoucher(&tc.voucher, tc.nights, now)
			if tc.wantErr == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr != "" && err == nil {
				t.Fatalf("expected error %q, got nil", tc.wantErr)
			}
			if tc.wantErr != "" && err != nil && !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

func TestValidateVoucher_MinNights(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		voucher model.Voucher
		nights  int
		wantErr string
	}{
		{"min 0 nights 1 ok", model.Voucher{Code: "A", Discount: 10, MinNights: 0}, 1, ""},
		{"min 3 nights 3 ok", model.Voucher{Code: "B", Discount: 10, MinNights: 3}, 3, ""},
		{"min 3 nights 5 ok", model.Voucher{Code: "C", Discount: 10, MinNights: 3}, 5, ""},
		{"min 3 nights 2 fail", model.Voucher{Code: "D", Discount: 10, MinNights: 3}, 2, "minimum 3 nights"},
		{"min 3 nights 1 fail", model.Voucher{Code: "E", Discount: 10, MinNights: 3}, 1, "minimum"},
		{"min 5 nights 4 fail", model.Voucher{Code: "F", Discount: 20, MinNights: 5}, 4, "minimum"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateVoucher(&tc.voucher, tc.nights, now)
			if tc.wantErr == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr != "" && err == nil {
				t.Fatalf("expected error %q", tc.wantErr)
			}
			if tc.wantErr != "" && err != nil && !strings.Contains(err.Error(), "minimum") {
				t.Fatalf("expected min_nights error, got %q", err.Error())
			}
		})
	}
}

func TestCalculateDiscountedPrice(t *testing.T) {
	tests := []struct {
		name      string
		total     int64
		discount  float64
		want      int64
	}{
		{"no discount", 300000, 0, 300000},
		{"10 percent", 200000, 10, 180000},
		{"50 percent", 100000, 50, 50000},
		{"100 percent free", 50000, 100, 0},
		{"over 100 clamped", 50000, 150, 0},
		{"negative clamped", 50000, -10, 50000},
		{"25 percent", 400000, 25, 300000},
		{"33.33 percent rounding", 30000, 33.33, 20001}, // 30000*66.67/100=20001
		{"12.5 percent", 80000, 12.5, 70000},
		{"zero total", 0, 50, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CalculateDiscountedPrice(tc.total, tc.discount)
			if got != tc.want {
				t.Fatalf("CalculateDiscountedPrice(%d, %.2f) = %d, want %d", tc.total, tc.discount, got, tc.want)
			}
		})
	}
}

func TestDiscountedTotal(t *testing.T) {
	tests := []struct {
		name           string
		pricePerNight  int64
		nights         int
		discount       float64
		want           int64
	}{
		{"2 nights 10% 100k", 100000, 2, 10, 180000},
		{"3 nights no discount", 100000, 3, 0, 300000},
		{"1 night 100% discount", 200000, 1, 100, 0},
		{"5 nights 20% 50000", 50000, 5, 20, 200000},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := DiscountedTotal(tc.pricePerNight, tc.nights, tc.discount)
			if got != tc.want {
				t.Fatalf("DiscountedTotal(%d,%d,%.2f)=%d want %d", tc.pricePerNight, tc.nights, tc.discount, got, tc.want)
			}
		})
	}
}

func TestValidateVoucher_TableCombined(t *testing.T) {
	now := time.Now()
	future := now.Add(48 * time.Hour)
	past := now.Add(-48 * time.Hour)
	quota2 := 2
	tests := []struct {
		name    string
		v       model.Voucher
		nights  int
		wantErr bool
	}{
		{"valid all", model.Voucher{Code: "GOOD", Discount: 15, MinNights: 2, Quota: &quota2, UsedCount: 0, ExpiresAt: &future}, 2, false},
		{"expired despite quota ok", model.Voucher{Code: "EXP", Discount: 15, MinNights: 1, Quota: &quota2, UsedCount: 0, ExpiresAt: &past}, 1, true},
		{"quota exceeded despite not expired", model.Voucher{Code: "QUO", Discount: 15, MinNights: 1, Quota: &quota2, UsedCount: 2, ExpiresAt: &future}, 2, true},
		{"min nights fail despite valid", model.Voucher{Code: "MIN", Discount: 15, MinNights: 3, Quota: &quota2, UsedCount: 0, ExpiresAt: &future}, 2, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateVoucher(&tc.v, tc.nights, now)
			if tc.wantErr && err == nil {
				t.Fatal("expected error but got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
