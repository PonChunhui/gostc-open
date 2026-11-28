package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v1.0.0", "v1.0.0", 0},
		{"v1.0.1", "v1.0.0", 1},
		{"v1.0.0", "v1.0.1", -1},
		{"v2.0.0", "v1.9.9", 1},
		{"v1.0.0", "v1.0.0-rc1", 1},  // 正式 > 预发布
		{"v1.0.0-rc2", "v1.0.0-rc1", 1},
		{"v1.0.0-rc1", "v1.0.0-rc2", -1},
		{"1.2.3", "v1.2.3", 0},
		{"v2.0.12-rc1", "v2.0.12-rc1", 0},
	}
	for _, c := range cases {
		if got := Compare(c.a, c.b); got != c.want {
			t.Errorf("Compare(%q,%q)=%d want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestSettingsRoundtrip(t *testing.T) {
	dir := t.TempDir()
	// 默认
	s := LoadSettings(dir)
	if s.Mode != ModeNotify || s.Channel != ChannelStable || s.IntervalHours != 24 {
		t.Fatalf("default settings wrong: %+v", s)
	}
	// 保存+读取
	want := Settings{Mode: ModeAuto, Channel: ChannelBeta, IntervalHours: 6, PanelURL: "https://x", AutoRestart: false, PublicKey: "ab"}
	if err := SaveSettingsFile(dir, want); err != nil {
		t.Fatal(err)
	}
	got := LoadSettings(dir)
	if got.Mode != ModeAuto || got.Channel != ChannelBeta || got.IntervalHours != 6 || got.PanelURL != "https://x" || got.AutoRestart || got.PublicKey != "ab" {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}
}

func TestVerifySha256(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	content := []byte("hello-gostc")
	os.WriteFile(p, content, 0644)
	sum := sha256.Sum256(content)
	hexSum := hex.EncodeToString(sum[:])
	if err := VerifySha256(p, hexSum); err != nil {
		t.Fatalf("valid sha failed: %v", err)
	}
	if err := VerifySha256(p, "deadbeef"); err == nil {
		t.Fatal("bad sha should fail")
	}
	if err := VerifySha256(p, ""); err == nil {
		t.Fatal("empty sha should fail (mandatory)")
	}
}

func makeZip(t *testing.T, path, binName string, content []byte) {
	f, _ := os.Create(path)
	defer f.Close()
	zw := zip.NewWriter(f)
	w, _ := zw.Create(binName)
	w.Write(content)
	zw.Close()
}

func makeTarGz(t *testing.T, path, binName string, content []byte) {
	f, _ := os.Create(path)
	defer f.Close()
	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)
	tw.WriteHeader(&tar.Header{Name: binName, Mode: 0755, Size: int64(len(content)), Typeflag: tar.TypeReg})
	tw.Write(content)
	tw.Close()
	gw.Close()
}

func TestExtractZipAndTarGz(t *testing.T) {
	dir := t.TempDir()
	content := []byte("BINARY-CONTENT")
	zipPath := filepath.Join(dir, "a.zip")
	makeZip(t, zipPath, "gostc.exe", content)
	dst := filepath.Join(dir, "out1")
	if err := Extract(zipPath, dst); err != nil {
		t.Fatalf("zip extract: %v", err)
	}
	if b, _ := os.ReadFile(dst); !bytes.Equal(b, content) {
		t.Fatal("zip content mismatch")
	}
	tgzPath := filepath.Join(dir, "a.tar.gz")
	makeTarGz(t, tgzPath, "gostc", content)
	dst2 := filepath.Join(dir, "out2")
	if err := Extract(tgzPath, dst2); err != nil {
		t.Fatalf("tar.gz extract: %v", err)
	}
	if b, _ := os.ReadFile(dst2); !bytes.Equal(b, content) {
		t.Fatal("tar.gz content mismatch")
	}
}

func TestAtomicReplaceAndRollback(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "gostc")
	newBin := filepath.Join(dir, "gostc-new")
	os.WriteFile(exe, []byte("OLD"), 0755)
	os.WriteFile(newBin, []byte("NEW"), 0755)
	if err := AtomicReplace(newBin, exe); err != nil {
		t.Fatalf("replace: %v", err)
	}
	if b, _ := os.ReadFile(exe); string(b) != "NEW" {
		t.Fatalf("exe should be NEW, got %q", b)
	}
	if err := Rollback(exe); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if b, _ := os.ReadFile(exe); string(b) != "OLD" {
		t.Fatalf("exe should be OLD after rollback, got %q", b)
	}
}

func TestCheckLatest(t *testing.T) {
	var resp latestResp
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	// 无更新（data nil）
	resp = latestResp{Code: 0}
	info, err := CheckLatest(srv.URL, "stable", "uid", "v1.0.0")
	if err != nil || info != nil {
		t.Fatalf("expected no update, got %v %v", info, err)
	}
	// 有更新
	resp = latestResp{Code: 0, Data: &Info{Version: "v1.1.0", DownloadURL: "https://x/a.zip", Sha256: "ab"}}
	info, err = CheckLatest(srv.URL, "stable", "uid", "v1.0.0")
	if err != nil || info == nil || info.Version != "v1.1.0" {
		t.Fatalf("expected update v1.1.0, got %v %v", info, err)
	}
	// 防降级：服务端版本更低
	resp = latestResp{Code: 0, Data: &Info{Version: "v0.9.0", DownloadURL: "https://x/a.zip"}}
	info, err = CheckLatest(srv.URL, "stable", "uid", "v1.0.0")
	if err != nil || info != nil {
		t.Fatalf("expected no downgrade, got %v %v", info, err)
	}
}
