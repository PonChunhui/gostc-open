package normal

import (
	"github.com/gin-gonic/gin"
	"server/controller/normal/failover"
	"server/global"
	"server/router/middleware"
)

func InitFailover(group *gin.RouterGroup) {
	g := group.Group("failover", middleware.Auth(global.Jwt))
	g.POST("log/page", failover.Page)
}
