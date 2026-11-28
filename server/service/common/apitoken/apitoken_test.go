package apitoken

import (
	"strings"
	"testing"
)

func TestGenerateAndHash(t *testing.T) {
	tok, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(tok, tokenPrefixMark) {
		t.Fatalf("token missing prefix: %s", tok)
	}
	h1 := Hash(tok)
	h2 := Hash(tok)
	if h1 != h2 || len(h1) != 64 {
		t.Fatalf("hash not stable sha256: %s", h1)
	}
	if Hash(tok+"x") == h1 {
		t.Fatal("different token should hash differently")
	}
	tok2, _ := Generate()
	if tok2 == tok {
		t.Fatal("tokens should be unique")
	}
}

func TestPrefix(t *testing.T) {
	if Prefix("gostc_abcdef1234567890") != "gostc_abcdef..." {
		t.Fatalf("unexpected prefix: %s", Prefix("gostc_abcdef1234567890"))
	}
	if Prefix("short") != "short" {
		t.Fatal("short token should be returned as-is")
	}
}

func TestResolveRejectsNonApiToken(t *testing.T) {
	if _, ok := Resolve("eyJhbGciOi.some.jwt"); ok {
		t.Fatal("JWT-shaped token must not resolve as API token")
	}
	if _, ok := Resolve(""); ok {
		t.Fatal("empty token must not resolve")
	}
}
