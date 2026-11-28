package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxDownloadBytes = 512 << 20 // 512MB 上限

var downloadClient = &http.Client{Timeout: time.Minute * 10}

// Download 下载文件到 dst（带大小上限）。
func Download(rawURL, dst string) error {
	resp, err := downloadClient.Get(rawURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载返回状态 %d", resp.StatusCode)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	n, err := io.Copy(f, io.LimitReader(resp.Body, maxDownloadBytes+1))
	if err != nil {
		return err
	}
	if n > maxDownloadBytes {
		return fmt.Errorf("下载文件超过大小上限")
	}
	return nil
}

// VerifySha256 校验文件 SHA256（hex，大小写不敏感）。expect 为空视为未提供，返回错误（强制校验）。
func VerifySha256(path, expect string) error {
	if strings.TrimSpace(expect) == "" {
		return fmt.Errorf("未提供 SHA256，拒绝升级")
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, strings.TrimSpace(expect)) {
		return fmt.Errorf("SHA256 校验失败: 期望 %s 实际 %s", expect, got)
	}
	return nil
}

// VerifySignature 用 ed25519 公钥校验文件签名。publicKeyHex 为空则跳过（返回 nil）。
func VerifySignature(path, sigHex, publicKeyHex string) error {
	if strings.TrimSpace(publicKeyHex) == "" {
		return nil // 未配置公钥，跳过签名校验（仍强制 SHA256）
	}
	pub, err := hex.DecodeString(strings.TrimSpace(publicKeyHex))
	if err != nil || len(pub) != ed25519.PublicKeySize {
		return fmt.Errorf("公钥格式错误")
	}
	sig, err := hex.DecodeString(strings.TrimSpace(sigHex))
	if err != nil {
		return fmt.Errorf("签名格式错误")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !ed25519.Verify(pub, data, sig) {
		return fmt.Errorf("签名校验失败")
	}
	return nil
}

// Extract 从归档（zip 或 tar.gz）中解出客户端二进制（gostc / gostc.exe）写入 dst。
func Extract(archivePath, dst string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	magic := make([]byte, 2)
	if _, err := io.ReadFull(f, magic); err != nil {
		return err
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if magic[0] == 0x1f && magic[1] == 0x8b {
		return extractTarGz(f, dst)
	}
	if magic[0] == 'P' && magic[1] == 'K' {
		return extractZip(archivePath, dst)
	}
	// 非归档：视为裸二进制
	return copyFile(archivePath, dst)
}

func isBinaryName(name string) bool {
	base := strings.ToLower(filepath.Base(name))
	return base == "gostc" || base == "gostc.exe"
}

func extractTarGz(r io.Reader, dst string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag != tar.TypeReg || !isBinaryName(hdr.Name) {
			continue
		}
		return writeReader(dst, tr)
	}
	return fmt.Errorf("归档中未找到 gostc 二进制")
}

func extractZip(archivePath, dst string) error {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer zr.Close()
	for _, zf := range zr.File {
		if zf.FileInfo().IsDir() || !isBinaryName(zf.Name) {
			continue
		}
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		err = writeReader(dst, rc)
		rc.Close()
		return err
	}
	return fmt.Errorf("归档中未找到 gostc 二进制")
}

func writeReader(dst string, r io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// moveFile 跨卷安全移动（rename 失败回退 copy+remove）。
func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	if err := copyFile(src, dst); err != nil {
		return err
	}
	return os.Remove(src)
}
