package middleware

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/model"
	"server/pkg/jwt"
	"server/repository"
	"strings"
	"go.uber.org/zap"
)

// Audit 记录管理/用户侧写操作（非 GET）审计日志。
// 仅记录元数据（操作者、方法、路径、IP、状态码），不记录请求体，避免泄露密码等敏感信息。
// 异步落库，失败仅告警，不影响主请求。
func Audit() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		method := c.Request.Method
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			return
		}
		path := c.Request.URL.Path
		if !strings.Contains(path, "/api/v1/admin/") && !strings.Contains(path, "/api/v1/normal/") {
			return
		}

		var userCode string
		if v, ok := c.Get("claims"); ok {
			if claims, ok2 := v.(jwt.Claims); ok2 {
				userCode = claims.Code
			}
		}
		ip := c.ClientIP()
		status := c.Writer.Status()

		go func(userCode, method, path, ip string, status int) {
			defer func() {
				if r := recover(); r != nil {
					global.Logger.Warn("audit log panic", zap.Any("recover", r))
				}
			}()
			db, _, _ := repository.Get("")
			account := ""
			if userCode != "" {
				if u, _ := db.SystemUser.Where(db.SystemUser.Code.Eq(userCode)).First(); u != nil {
					account = u.Account
				}
			}
			if err := db.SystemAuditLog.Create(&model.SystemAuditLog{
				UserCode: userCode,
				Account:  account,
				Method:   method,
				Path:     path,
				IP:       ip,
				Status:   status,
			}); err != nil {
				global.Logger.Warn("write audit log fail", zap.Error(err))
			}
		}(userCode, method, path, ip, status)
	}
}
