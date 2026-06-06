package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func init() {
	SetSecret("test-secret-key-for-unit-tests")
}

func TestGenerateToken_Success(t *testing.T) {
	userID := uuid.New()
	username := "testuser"
	token, err := GenerateToken(userID, username)
	if err != nil {
		t.Fatalf("GenerateToken() error: %v", err)
	}
	if token == "" {
		t.Error("GenerateToken() returned empty token")
	}
}

func TestParseToken_Valid(t *testing.T) {
	userID := uuid.New()
	username := "testuser"
	token, err := GenerateToken(userID, username)
	if err != nil {
		t.Fatalf("GenerateToken() error: %v", err)
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken() error: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.Username != username {
		t.Errorf("Username = %q, want %q", claims.Username, username)
	}
}

func TestParseToken_Expired(t *testing.T) {
	// Manually create an expired token
	userID := uuid.New()
	claims := Claims{
		UserID:   userID,
		Username: "expired",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString(jwtSecret)

	_, err := ParseToken(tokenStr)
	if err == nil {
		t.Error("ParseToken() expected error for expired token, got nil")
	}
}

func TestParseToken_Tampered(t *testing.T) {
	userID := uuid.New()
	token, _ := GenerateToken(userID, "user")
	// Tamper with the token by changing a character
	tampered := token[:len(token)-2] + "XX"
	_, err := ParseToken(tampered)
	if err == nil {
		t.Error("ParseToken() expected error for tampered token, got nil")
	}
}

func TestParseToken_Empty(t *testing.T) {
	_, err := ParseToken("")
	if err == nil {
		t.Error("ParseToken() expected error for empty token, got nil")
	}
}

func TestParseToken_InvalidString(t *testing.T) {
	_, err := ParseToken("not-a-valid-token")
	if err == nil {
		t.Error("ParseToken() expected error for invalid token string, got nil")
	}
}

func TestGenerateToken_72HourExpiry(t *testing.T) {
	before := time.Now()
	token, _ := GenerateToken(uuid.New(), "user")
	claims, _ := ParseToken(token)

	expiresAt := claims.ExpiresAt.Time
	expected := before.Add(72 * time.Hour)

	// Allow 5 second tolerance
	diff := expiresAt.Sub(expected)
	if diff > 5*time.Second || diff < -5*time.Second {
		t.Errorf("Expiry = %v, want ~%v (diff: %v)", expiresAt, expected, diff)
	}
}

func TestSetSecret_ChangesSigningKey(t *testing.T) {
	originalSecret := string(jwtSecret)
	defer SetSecret(originalSecret)

	SetSecret("original-key")
	token, _ := GenerateToken(uuid.New(), "user")

	SetSecret("different-key")
	_, err := ParseToken(token)
	if err == nil {
		t.Error("ParseToken() expected error when secret changes, got nil")
	}
}
