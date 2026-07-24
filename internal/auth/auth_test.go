package auth

import (
	"errors"
	"testing"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "" || hash == "correct horse battery staple" {
		t.Fatalf("HashPassword() returned unusable hash %q", hash)
	}

	if err := VerifyPassword(hash, "correct horse battery staple"); err != nil {
		t.Fatalf("VerifyPassword() with correct password error = %v, want nil", err)
	}

	err = VerifyPassword(hash, "wrong password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("VerifyPassword() with wrong password error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestVerifyPasswordNeverLeaksHashDetail(t *testing.T) {
	err := VerifyPassword("not-a-real-bcrypt-hash", "anything")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("VerifyPassword() with malformed hash error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestNewSessionTokenIsRandomAndHashable(t *testing.T) {
	rawA, hashA, err := NewSessionToken()
	if err != nil {
		t.Fatalf("NewSessionToken() error = %v", err)
	}
	rawB, hashB, err := NewSessionToken()
	if err != nil {
		t.Fatalf("NewSessionToken() error = %v", err)
	}

	if rawA == rawB {
		t.Fatalf("NewSessionToken() produced identical raw tokens across calls")
	}
	if len(rawA) != 64 {
		t.Fatalf("len(raw) = %d, want 64 (32 bytes hex-encoded)", len(rawA))
	}
	if hashA != HashToken(rawA) {
		t.Fatalf("HashToken(raw) = %q, want %q", HashToken(rawA), hashA)
	}
	if hashA == hashB {
		t.Fatalf("NewSessionToken() produced identical hashes for distinct tokens")
	}
	if hashA == rawA {
		t.Fatalf("NewSessionToken() hash must not equal the raw token")
	}
}

func TestConstantTimeEqual(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{"equal", "abc123", "abc123", true},
		{"different length", "abc", "abcd", false},
		{"same length different content", "abc123", "abc124", false},
		{"both empty", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConstantTimeEqual(tt.a, tt.b); got != tt.want {
				t.Fatalf("ConstantTimeEqual(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestHasAtLeast(t *testing.T) {
	tests := []struct {
		role    types.Role
		minimum types.Role
		want    bool
	}{
		{types.RoleAdmin, types.RoleReadOnly, true},
		{types.RoleAdmin, types.RoleOperator, true},
		{types.RoleAdmin, types.RoleAdmin, true},
		{types.RoleOperator, types.RoleAdmin, false},
		{types.RoleOperator, types.RoleOperator, true},
		{types.RoleReadOnly, types.RoleOperator, false},
		{types.RoleReadOnly, types.RoleReadOnly, true},
		{types.Role("bogus"), types.RoleReadOnly, false},
		{types.RoleAdmin, types.Role("bogus"), false},
	}
	for _, tt := range tests {
		if got := HasAtLeast(tt.role, tt.minimum); got != tt.want {
			t.Fatalf("HasAtLeast(%q, %q) = %v, want %v", tt.role, tt.minimum, got, tt.want)
		}
	}
}
