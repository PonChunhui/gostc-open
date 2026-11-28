// Package crypto 提供敏感凭据的对称加密（AES-256-GCM）。
// 密文带版本前缀 enc:v1:，便于将来算法轮换；主密钥来自配置/环境变量。
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"strings"
)

const prefix = "enc:v1:"

// deriveKey 将任意长度主密钥派生为 32 字节（AES-256）。
func deriveKey(key string) []byte {
	sum := sha256.Sum256([]byte(key))
	return sum[:]
}

// IsEncrypted 判断字符串是否已是本包密文。
func IsEncrypted(s string) bool {
	return strings.HasPrefix(s, prefix)
}

// Encrypt 用主密钥加密明文，返回 enc:v1:base64(nonce||ciphertext)。
// key 为空返回错误（调用方应先判断是否启用加密）。
func Encrypt(plain, key string) (string, error) {
	if key == "" {
		return "", errors.New("encrypt: empty master key")
	}
	block, err := aes.NewCipher(deriveKey(key))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return prefix + base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt 解密本包密文；非密文原样返回（兼容存量明文）。
// 密文损坏或密钥错误返回错误，绝不静默回退明文。
func Decrypt(s, key string) (string, error) {
	if !IsEncrypted(s) {
		return s, nil
	}
	if key == "" {
		return "", errors.New("decrypt: empty master key for encrypted value")
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(s, prefix))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(deriveKey(key))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ns := gcm.NonceSize()
	if len(raw) < ns {
		return "", errors.New("decrypt: ciphertext too short")
	}
	nonce, ct := raw[:ns], raw[ns:]
	plain, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// EncryptIfEnabled 启用加密（key 非空）时加密，否则原样返回。幂等：已加密不再加密。
func EncryptIfEnabled(plain, key string) (string, error) {
	if key == "" || IsEncrypted(plain) {
		return plain, nil
	}
	return Encrypt(plain, key)
}

// DecryptSafe 尽力解密：非密文或解密失败时原样返回，用于展示/比对等容错场景。
// 需要严格区分错误的调用方应使用 Decrypt。
func DecryptSafe(s, key string) string {
	plain, err := Decrypt(s, key)
	if err != nil {
		return s
	}
	return plain
}
