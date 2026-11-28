package utils

import "testing"

func TestHashVerifyBcrypt(t *testing.T) {
	h, err := HashPassword("Passw0rd!")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !isBcrypt(h) {
		t.Fatalf("expected bcrypt hash, got %q", h)
	}
	if !VerifyPassword("Passw0rd!", h, "") {
		t.Fatal("correct password should verify")
	}
	if VerifyPassword("wrong", h, "") {
		t.Fatal("wrong password should not verify")
	}
	if PasswordNeedsUpgrade(h) {
		t.Fatal("bcrypt should not need upgrade")
	}
}

func TestVerifyLegacyMD5(t *testing.T) {
	salt := "somesalt"
	legacy := MD5AndSalt("oldpass", salt)
	if !VerifyPassword("oldpass", legacy, salt) {
		t.Fatal("legacy MD5 should verify")
	}
	if VerifyPassword("bad", legacy, salt) {
		t.Fatal("legacy wrong password should not verify")
	}
	if !PasswordNeedsUpgrade(legacy) {
		t.Fatal("legacy MD5 should need upgrade")
	}
}
