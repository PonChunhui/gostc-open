package normal

import (
	"github.com/gin-gonic/gin"
	"server/controller/normal/template"
	"server/global"
	"server/router/middleware"
)

func InitTemplate(group *gin.RouterGroup) {
	g := group.Group("template", middleware.Auth(global.Jwt))
	g.POST("create", template.Create)
	g.POST("delete", template.Delete)
	g.POST("page", template.Page)
	g.POST("apply", template.Apply)
}
