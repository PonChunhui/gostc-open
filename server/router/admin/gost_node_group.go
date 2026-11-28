package admin

import (
	"github.com/gin-gonic/gin"
	"server/controller/admin/gost_node_group"
	"server/global"
	"server/router/middleware"
)

func InitGostNodeGroup(group *gin.RouterGroup) {
	g := group.Group("gost/node/group", middleware.Auth(global.Jwt), middleware.AuthAdmin())
	g.POST("create", gost_node_group.Create)
	g.POST("update", gost_node_group.Update)
	g.POST("delete", gost_node_group.Delete)
	g.POST("page", gost_node_group.Page)
}
