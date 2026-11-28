package task

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
)

func genCertPEM(t *testing.T, notAfter time.Time) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     notAfter,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

func TestParseCertNotAfter(t *testing.T) {
	want := time.Now().AddDate(0, 0, 20).Truncate(time.Second)
	got, ok := parseCertNotAfter(genCertPEM(t, want))
	if !ok {
		t.Fatal("解析证书失败")
	}
	if !got.Equal(want) {
		t.Fatalf("NotAfter=%v want %v", got, want)
	}
	if _, ok := parseCertNotAfter("not a pem"); ok {
		t.Fatal("非法 PEM 应返回 false")
	}
}
