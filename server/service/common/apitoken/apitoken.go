package apitoken

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"server/pkg/jwt"
	"server/repository"
)

const tokenPrefixMark = "gostc_"

// Generate 生成新的明文 API Token（仅创建时返回一次）。
func Generate() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return tokenPrefixMark + hex.EncodeToString(buf), nil
}

// Hash 计算 Token 的 SHA-256 哈希（存储用）。
func Hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Prefix 返回用于展示的 Token 前缀。
func Prefix(token string) string {
	if len(token) > 12 {
		return token[:12] + "..."
	}
	return token
}

// Resolve 将 API Token 解析为登录态 Claims；非 API Token 或无效/过期返回 false。
func Resolve(token string) (jwt.Claims, bool) {
	if !strings.HasPrefix(token, tokenPrefixMark) {
		return jwt.Claims{}, false
	}
	db, _, _ := repository.Get("")
	rec, err := db.SystemApiToken.Where(db.SystemApiToken.TokenHash.Eq(Hash(token))).First()
	if err != nil {
		return jwt.Claims{}, false
	}
	if rec.ExpAt > 0 && rec.ExpAt < time.Now().Unix() {
		return jwt.Claims{}, false
	}
	user, err := db.SystemUser.Where(db.SystemUser.Code.Eq(rec.UserCode)).First()
	if err != nil {
		return jwt.Claims{}, false
	}
	_, _ = db.SystemApiToken.Where(db.SystemApiToken.Code.Eq(rec.Code)).UpdateSimple(db.SystemApiToken.LastUsedAt.Value(time.Now().Unix()))
	return jwt.Claims{
		Code: user.Code,
		Data: map[string]string{"admin": strconv.Itoa(user.Admin)},
	}, true
}
