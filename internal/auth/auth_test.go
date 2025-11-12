package auth

import (
	"bytes"
	"errors"
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	h, err := HashPassword("correct horse")
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckPassword(h, "correct horse"); err != nil {
		t.Fatalf("matching password rejected: %v", err)
	}
	if err := CheckPassword(h, "wrong horse"); !errors.Is(err, ErrWrongPassword) {
		t.Fatalf("err = %v, want ErrWrongPassword", err)
	}
}

func TestSessionTokens(t *testing.T) {
	tok1, hash1, err := NewSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	tok2, _, _ := NewSessionToken()
	if tok1 == tok2 {
		t.Fatal("tokens should be random")
	}
	if len(tok1) < 40 {
		t.Fatalf("token too short: %q", tok1)
	}
	if !bytes.Equal(HashToken(tok1), hash1) {
		t.Fatal("HashToken does not match the hash returned with the token")
	}
}
