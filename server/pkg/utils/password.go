package utils

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// bcryptPrefix bcrypt 哈希前缀（$2a$/$2b$/$2y$）。
func isBcrypt(stored string) bool {
	return strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") || strings.HasPrefix(stored, "$2y$")
}

// HashPassword 生成 bcrypt 哈希（新注册/改密/登录升级用）。
func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// VerifyPassword 校验密码：bcrypt 哈希走 bcrypt，旧 MD5+salt 哈希走 legacy。
func VerifyPassword(password, stored, salt string) bool {
	if isBcrypt(stored) {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(password)) == nil
	}
	return MD5AndSalt(password, salt) == stored
}

// PasswordNeedsUpgrade 旧 MD5 哈希需要在登录成功后透明升级为 bcrypt。
func PasswordNeedsUpgrade(stored string) bool {
	return !isBcrypt(stored)
}
