package admin

import (
	"github.com/gin-gonic/gin"
	"server/controller/admin/failover"
	"server/global"
	"server/router/middleware"
)

func InitFailover(group *gin.RouterGroup) {
	g := group.Group("failover", middleware.Auth(global.Jwt), middleware.AuthAdmin())
	g.POST("log/page", failover.Page)
}
