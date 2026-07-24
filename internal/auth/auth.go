// Package auth implements password hashing, session tokens, and role-based
// permission checks for the kfleet hub.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidCredentials is returned when a password does not match a stored hash.
var ErrInvalidCredentials = errors.New("invalid credentials")

// sessionTokenBytes is the size of a raw session token before hex encoding.
// 32 bytes (256 bits) matches the agent registration token strength.
const sessionTokenBytes = 32

// dummyPasswordHash is a valid bcrypt hash used only to equalize the work
// performed for an unknown username with the work performed for a known user.
// It is not associated with an account and cannot grant access.
const dummyPasswordHash = "$2y$10$C.uA6NnV8uyP8LuHF4NWB.aMKjPd8hmPh4EsY7zoQUbI.Ch1zhrJS"

// HashPassword returns a bcrypt hash of the supplied plaintext password.
// The plaintext is never logged or persisted by callers of this function.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword reports whether password matches the bcrypt hash. It
// returns ErrInvalidCredentials (never the underlying bcrypt error, which
// can echo hash material) when the password does not match.
func VerifyPassword(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

// ConsumePasswordVerification performs one bcrypt comparison without exposing
// whether a username exists. Login handlers call it on account lookup misses
// to reduce username enumeration through response timing.
func ConsumePasswordVerification(password string) {
	_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(password))
}

// NewSessionToken generates a random session token and returns both the raw
// value, which is sent to the client once, and its SHA-256 hash, which is
// the only form persisted server-side.
func NewSessionToken() (raw string, hash string, err error) {
	buf := make([]byte, sessionTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("generate session token: %w", err)
	}
	raw = hex.EncodeToString(buf)
	return raw, HashToken(raw), nil
}

// HashToken returns the SHA-256 hex digest of a raw token. Only the digest
// is ever persisted; the raw token is never logged or stored.
func HashToken(raw string) string {
	digest := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(digest[:])
}

// ConstantTimeEqual reports whether a and b are equal using a comparison
// whose running time does not depend on where the strings first differ.
func ConstantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// rolePriority ranks roles from least to most privileged so callers can
// express "this role or higher" checks.
var rolePriority = map[types.Role]int{
	types.RoleReadOnly: 0,
	types.RoleOperator: 1,
	types.RoleAdmin:    2,
}

// HasAtLeast reports whether role meets or exceeds the privilege of
// minimum. An unrecognized role never satisfies any minimum.
func HasAtLeast(role, minimum types.Role) bool {
	rolePriorityValue, ok := rolePriority[role]
	if !ok {
		return false
	}
	minimumPriority, ok := rolePriority[minimum]
	if !ok {
		return false
	}
	return rolePriorityValue >= minimumPriority
}
