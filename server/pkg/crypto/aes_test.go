package crypto

import "testing"

func TestEncryptDecryptRoundtrip(t *testing.T) {
	key := "master-secret"
	plain := "-----BEGIN PRIVATE KEY-----\nabc123\n-----END PRIVATE KEY-----"
	enc, err := Encrypt(plain, key)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if !IsEncrypted(enc) {
		t.Fatalf("expected ciphertext prefix, got %q", enc)
	}
	if enc == plain {
		t.Fatal("ciphertext equals plaintext")
	}
	dec, err := Decrypt(enc, key)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if dec != plain {
		t.Fatalf("roundtrip mismatch: %q != %q", dec, plain)
	}
}

func TestEncryptEmptyKey(t *testing.T) {
	if _, err := Encrypt("x", ""); err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestEncryptIfEnabledDisabled(t *testing.T) {
	got, err := EncryptIfEnabled("plain", "")
	if err != nil || got != "plain" {
		t.Fatalf("disabled should passthrough: %q %v", got, err)
	}
}

func TestEncryptIfEnabledIdempotent(t *testing.T) {
	key := "k"
	enc, _ := EncryptIfEnabled("data", key)
	again, err := EncryptIfEnabled(enc, key)
	if err != nil || again != enc {
		t.Fatalf("should be idempotent: %v", err)
	}
}

func TestDecryptNonEncryptedPassthrough(t *testing.T) {
	got, err := Decrypt("legacy-plain", "k")
	if err != nil || got != "legacy-plain" {
		t.Fatalf("non-encrypted should passthrough: %q %v", got, err)
	}
}

func TestDecryptWrongKey(t *testing.T) {
	enc, _ := Encrypt("secret", "right")
	if _, err := Decrypt(enc, "wrong"); err == nil {
		t.Fatal("expected error decrypting with wrong key")
	}
}

func TestDecryptSafe(t *testing.T) {
	enc, _ := Encrypt("secret", "k")
	if got := DecryptSafe(enc, "k"); got != "secret" {
		t.Fatalf("DecryptSafe correct key: %q", got)
	}
	if got := DecryptSafe(enc, "wrong"); got != enc {
		t.Fatalf("DecryptSafe wrong key should return original: %q", got)
	}
	if got := DecryptSafe("plain", "k"); got != "plain" {
		t.Fatalf("DecryptSafe plain: %q", got)
	}
}
