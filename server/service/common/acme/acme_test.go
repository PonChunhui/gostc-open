package acme

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"server/model"
)

func TestBuildProvider(t *testing.T) {
	cases := []struct {
		name     string
		provider string
		cred     Credential
		wantErr  bool
	}{
		{"dnspod缺凭据", model.CERT_PROVIDER_DNSPOD, Credential{}, true},
		{"dnspod有凭据", model.CERT_PROVIDER_DNSPOD, Credential{LoginToken: "123,abc"}, false},
		{"cloudflare缺凭据", model.CERT_PROVIDER_CLOUDFLARE, Credential{}, true},
		{"cloudflare token", model.CERT_PROVIDER_CLOUDFLARE, Credential{AuthToken: "tok"}, false},
		{"cloudflare key模式", model.CERT_PROVIDER_CLOUDFLARE, Credential{AuthEmail: "a@b.c", AuthKey: "k"}, false},
		{"alidns缺凭据", model.CERT_PROVIDER_ALIDNS, Credential{APIKey: "x"}, true},
		{"alidns有凭据", model.CERT_PROVIDER_ALIDNS, Credential{APIKey: "x", SecretKey: "y"}, false},
		{"未知provider", "namecheap", Credential{}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := buildProvider(c.provider, c.cred)
			if (err != nil) != c.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, c.wantErr)
			}
		})
	}
}

func genCertPEM(t *testing.T, notAfter time.Time) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test.example.com"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     notAfter,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func TestParseExpiry(t *testing.T) {
	want := time.Now().Add(90 * 24 * time.Hour).UTC().Truncate(time.Second)
	pemBytes := genCertPEM(t, want)
	got, err := ParseExpiry(pemBytes)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if _, err := ParseExpiry([]byte("not a pem")); err == nil {
		t.Fatal("expected error for invalid pem")
	}
}
