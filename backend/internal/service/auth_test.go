package service

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"xyz-hotel/backend/internal/model"
	"xyz-hotel/backend/internal/repo"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantFail bool
	}{
		{"simple", "password123", false},
		{"with special", "P@ssw0rd!2024", false},
		{"long", "this-is-a-very-long-password-that-should-still-hash-ok-123", false},
		{"min length 8", "12345678", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := HashPassword(tc.password)
			if err != nil {
				t.Fatalf("HashPassword failed: %v", err)
			}
			if hash == tc.password {
				t.Fatal("hash equals plain password")
			}
			if !CheckPassword(hash, tc.password) {
				t.Fatal("CheckPassword should succeed for correct password")
			}
			if CheckPassword(hash, tc.password+"x") {
				t.Fatal("CheckPassword should fail for wrong password")
			}
		})
	}
}

func TestCheckPassword_Wrong(t *testing.T) {
	hash, _ := HashPassword("correct-horse")
	cases := []struct {
		name     string
		password string
		want     bool
	}{
		{"correct", "correct-horse", true},
		{"wrong", "wrong-horse", false},
		{"empty", "", false},
		{"case sensitive", "Correct-horse", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CheckPassword(hash, tc.password)
			if got != tc.want {
				t.Fatalf("CheckPassword(%q) = %v, want %v", tc.password, got, tc.want)
			}
		})
	}
}

func TestGenerateTokens(t *testing.T) {
	svc := NewAuthService(&repo.UserRepo{DB: nil}, "test-secret-for-unit-tests-12345")
	user := &model.User{ID: 42, Name: "Test User", Email: "test@example.com", Role: model.RoleCustomer}

	access, refresh, err := svc.IssueTokensForTest(user)
	if err != nil {
		t.Fatalf("IssueTokensForTest failed: %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatal("tokens should not be empty")
	}
	if access == refresh {
		t.Fatal("access and refresh should differ")
	}

	// Verify access claims
	parseAndCheck := func(tokenStr, expectedTyp string, expectedExp time.Duration) {
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte("test-secret-for-unit-tests-12345"), nil
		})
		if err != nil || !token.Valid {
			t.Fatalf("parse %s token failed: %v", expectedTyp, err)
		}
		if claims.UserID != 42 {
			t.Fatalf("claims UserID = %d, want 42", claims.UserID)
		}
		if claims.Role != model.RoleCustomer {
			t.Fatalf("claims Role = %s, want customer", claims.Role)
		}
		if claims.TokenType != expectedTyp {
			t.Fatalf("TokenType = %s, want %s", claims.TokenType, expectedTyp)
		}
		if claims.ExpiresAt == nil || claims.IssuedAt == nil {
			t.Fatal("missing exp/iat")
		}
		diff := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time)
		// Allow 2s leeway
		if diff < expectedExp-2*time.Second || diff > expectedExp+2*time.Second {
			t.Fatalf("expiry diff for %s = %v, want ~%v", expectedTyp, diff, expectedExp)
		}
	}
	parseAndCheck(access, "access", 15*time.Minute)
	parseAndCheck(refresh, "refresh", 7*24*time.Hour)

	// Access token should not be usable as refresh (middleware blocks refresh typ)
	// Simulate Refresh rejecting access token (without DB, we check typ check before DB)
	// Create a scenario where we try to parse access token as refresh: typ mismatch should be caught
	claims := &Claims{}
	tok, _ := jwt.ParseWithClaims(access, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte("test-secret-for-unit-tests-12345"), nil
	})
	if !tok.Valid {
		t.Fatal("access token should be valid")
	}
	if claims.TokenType == "refresh" {
		t.Fatal("access token typ should not be refresh")
	}
}

func TestGenerateTokens_CustomerRoleEnforced(t *testing.T) {
	// Ensure token embeds role correctly for RBAC
	svc := NewAuthService(&repo.UserRepo{DB: nil}, "another-secret")
	cases := []struct {
		role string
	}{
		{model.RoleOwner},
		{model.RoleManager},
		{model.RoleReceptionist},
		{model.RoleCustomer},
	}
	for _, tc := range cases {
		t.Run(tc.role, func(t *testing.T) {
			u := &model.User{ID: 1, Email: "a@b.com", Role: tc.role}
			access, _, err := svc.IssueTokensForTest(u)
			if err != nil {
				t.Fatalf("issue failed: %v", err)
			}
			claims := &Claims{}
			_, err = jwt.ParseWithClaims(access, claims, func(t *jwt.Token) (interface{}, error) {
				return []byte("another-secret"), nil
			})
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}
			if claims.Role != tc.role {
				t.Fatalf("role %s != %s", claims.Role, tc.role)
			}
		})
	}
}
