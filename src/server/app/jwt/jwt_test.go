package jwt

import (
	"errors"
	"strings"
	"testing"
	"time"
)

const jwtSecret = "pWHPDKSX354kk21m2VRhIv42PNiOazOD"

func TestWrongSecret(t *testing.T) {
	_, err := Init("too-short-secret")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.HasPrefix(err.Error(), "invalid key size:") {
		t.Fatalf("expected 'invalid key size' error")
	}
}
func TestInvalidToken(t *testing.T) {
	jwt, err := Init(jwtSecret)
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Generate("issuer", "tenant", 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	_, err = jwt.Validate(token)
	if err != nil {
		t.Fatal(err)
	}
}
func TestExpiredToken(t *testing.T) {
	jwt, err := Init(jwtSecret)
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Generate("issuer", "tenant", 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	// here we wait until token expiration
	time.Sleep(150 * time.Millisecond)
	_, err = jwt.Validate(token)
	if err != nil {
		if !errors.Is(err, ErrExpiredToken) {
			t.Fatal(err)
		}
	}
}
func TestTokenPayload(t *testing.T) {
	jwt, err := Init(jwtSecret)
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Generate("issuer", "tenant", 1*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := jwt.Validate(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.GetIssuer() != "issuer" {
		t.Errorf("unexpected issuer found")
	}
}
