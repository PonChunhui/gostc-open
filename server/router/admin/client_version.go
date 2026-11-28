package admin

import (
	"github.com/gin-gonic/gin"
	"server/controller/admin/client_version"
	"server/global"
	"server/router/middleware"
)

func InitClientVersion(group *gin.RouterGroup) {
	g := group.Group("client/version", middleware.Auth(global.Jwt), middleware.AuthAdmin())
	g.POST("create", client_version.Create)
	g.POST("update", client_version.Update)
	g.POST("delete", client_version.Delete)
	g.POST("page", client_version.Page)
}
