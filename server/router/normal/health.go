package normal

import (
	"github.com/gin-gonic/gin"
	"server/controller/normal/health"
	"server/global"
	"server/router/middleware"
)

func InitHealth(group *gin.RouterGroup) {
	g := group.Group("gost/client/health", middleware.Auth(global.Jwt))
	g.POST("page", health.Page)
	g.POST("overview", health.Overview)
}
