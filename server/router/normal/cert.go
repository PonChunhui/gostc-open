package normal

import (
	"github.com/gin-gonic/gin"
	"server/controller/normal/cert"
	"server/global"
	"server/router/middleware"
)

func InitCert(group *gin.RouterGroup) {
	g := group.Group("cert", middleware.Auth(global.Jwt))
	g.POST("create", cert.Create)
	g.POST("update", cert.Update)
	g.POST("delete", cert.Delete)
	g.POST("page", cert.Page)
	g.POST("issue", cert.Issue)
	g.POST("renew", cert.Renew)
	g.POST("upload", cert.Upload)
}
