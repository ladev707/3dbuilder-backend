package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	t.Parallel()
	password := "password"
	
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hashedPassword == "" {
		t.Fatalf("HashPassword returned empty string")
	}
	t.Logf("Hashed password: %s", hashedPassword)
	t.Logf("Password: %s", password)
}
