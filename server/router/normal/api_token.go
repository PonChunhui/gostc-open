package normal

import (
	"github.com/gin-gonic/gin"
	"server/controller/normal/api_token"
	"server/global"
	"server/router/middleware"
)

func InitApiToken(group *gin.RouterGroup) {
	g := group.Group("api-token", middleware.Auth(global.Jwt))
	g.POST("create", api_token.Create)
	g.POST("delete", api_token.Delete)
	g.POST("page", api_token.Page)
}
